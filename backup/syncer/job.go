/*
Copyright 2021 RadonDB.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package syncer

import (
	"context"
	"fmt"
	"strings"

	"github.com/presslabs/controller-util/pkg/syncer"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/backup"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

type jobSyncer struct {
	client client.Client
	job    *batchv1.Job
	backup *backup.Backup
}

// Owner returns the object owner or nil if object does not have one.
func (s *jobSyncer) ObjectOwner() runtime.Object { return s.backup.Unwrap() }

// NewJobSyncer returns a syncer for backup jobs
func NewJobSyncer(c client.Client, backup *backup.Backup) syncer.Interface {
	obj := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backup.GetNameForJob(),
			Namespace: backup.Namespace,
		},
	}

	sync := &jobSyncer{
		client: c,
		job:    obj,
		backup: backup,
	}

	return syncer.NewObjectSyncer("Job", backup.Unwrap(), obj, c, sync.SyncFn)
}

func (s *jobSyncer) SyncFn() error {
	if s.backup.Status.Completed {
		s.backup.Log.V(1).Info("backup already completed", "backup", s.backup)
		// skip doing anything
		return syncer.ErrIgnore
	}

	// check if job is already created and just update the status
	if !s.job.ObjectMeta.CreationTimestamp.IsZero() {
		s.updateStatus(s.job)
		return nil
	}

	s.job.Labels = map[string]string{
		"Host": s.backup.Spec.HostName,
		"Type": utils.BackupJobTypeName,

		// Cluster used as selector.
		"Cluster": s.backup.Spec.ClusterName,
	}
	var backoff int32 = 3
	s.job.Spec.Template.Spec = s.ensurePodSpec(s.job.Spec.Template.Spec)
	s.job.Spec.BackoffLimit = &backoff
	return nil
}

func (s *jobSyncer) updateStatus(job *batchv1.Job) {
	// check for completion condition
	s.backup.Status.Completed = false
	s.backup.UpdateStatusCondition(v1alpha1.BackupStart, corev1.ConditionTrue, "backup has started", "backup has started")
	if cond := jobCondition(batchv1.JobComplete, job); cond != nil {
		s.backup.UpdateStatusCondition(v1alpha1.BackupComplete, cond.Status, cond.Reason, cond.Message)
		if cond.Status == corev1.ConditionTrue {
			s.backup.Status.Completed = true
		}
		if backupName := s.job.Annotations[utils.JobAnonationName]; backupName != "" {
			s.backup.Status.BackupName = backupName
		}
		if backDate := s.job.Annotations[utils.JobAnonationDate]; backDate != "" {
			s.backup.Status.BackupDate = backDate
		}
		if backType := s.job.Annotations[utils.JobAnonationType]; backType != "" {
			s.backup.Status.BackupType = backType
		}
	}

	// check for failed condition
	if cond := jobCondition(batchv1.JobFailed, job); cond != nil {
		s.backup.UpdateStatusCondition(v1alpha1.BackupFailed, cond.Status, cond.Reason, cond.Message)
		s.backup.UpdateStatusCondition(v1alpha1.BackupComplete, corev1.ConditionFalse, cond.Reason, cond.Message)
		if cond.Status == corev1.ConditionTrue {
			s.backup.Status.Completed = true
		}
	}

}

func jobCondition(condType batchv1.JobConditionType, job *batchv1.Job) *batchv1.JobCondition {
	for _, c := range job.Status.Conditions {
		if c.Type == condType {
			return &c
		}
	}

	return nil
}

func (s *jobSyncer) ensurePodSpec(in corev1.PodSpec) corev1.PodSpec {
	if len(in.Containers) == 0 {
		in.Containers = make([]corev1.Container, 1)
	}

	in.RestartPolicy = corev1.RestartPolicyNever
	sctName := fmt.Sprintf("%s-secret", s.backup.Spec.ClusterName)
	in.Containers[0].Name = utils.ContainerBackupName
	in.Containers[0].Image = mysqlcluster.GetImage(s.backup.Spec.Image)
	in.ServiceAccountName = s.backup.Spec.ClusterName
	if len(s.backup.Spec.NFSServerAddress) != 0 {
		//parse NFSServerAddress to IP:/Path
		ip, path := utils.ParseIPAndPath(s.backup.Spec.NFSServerAddress)
		// add volumn about pvc
		in.Volumes = []corev1.Volume{
			{
				Name: utils.XtrabackupPV,
				VolumeSource: corev1.VolumeSource{
					NFS: &corev1.NFSVolumeSource{
						Server: ip,
						Path:   path,
					},
				},
			},
		}
		//"rm -rf /backup/*;curl --user sys_backups:sys_backups sample-mysql-0.sample-mysql.default:8082/download|xbstream -x -C /backup"
		in.Containers[0].Command = []string{
			"/bin/bash", "-c", "--",
		}
		backupToDir, DateTime := utils.BuildBackupName(s.backup.Spec.ClusterName)
		strAnnonations := fmt.Sprintf(`curl -X PATCH -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" -H "Content-Type: application/json-patch+json" \
		--cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/apis/batch/v1/namespaces/%s/jobs/%s \
		 -d '[{"op": "add", "path": "/metadata/annotations/backupName", "value": "%s"}, {"op": "add", "path": "/metadata/annotations/backupDate", "value": "%s"}, {"op": "add", "path": "/metadata/annotations/backupType", "value": "NFS"}]';`,
			s.backup.Namespace, s.backup.GetNameForJob(), backupToDir, DateTime)
		// Add the check DiskUsage
		// use expr because shell cannot compare float number
		checkUsage := `[ $(echo "$(df /backup|awk 'NR>1 {print $4}') > $(du  /backup |awk 'END {if (NR > 1) {print $1 /(NR-1)} else print 0}')"|bc) -eq '1' ] || { echo disk available may be too small; exit 1;};`
		in.Containers[0].Args = []string{
			checkUsage + fmt.Sprintf("mkdir -p /backup/%s;"+
				"curl --user $BACKUP_USER:$BACKUP_PASSWORD %s/download|xbstream -x -C /backup/%s; err1=${PIPESTATUS[0]};"+
				strAnnonations+"retval_final=$?; exit $err1||$retval_final",
				backupToDir,
				s.backup.GetBackupURL(s.backup.Spec.ClusterName, s.backup.Spec.HostName), backupToDir),
		}
		in.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      utils.XtrabackupPV,
				MountPath: utils.XtrabckupLocal,
			},
		}
	} else if s.backup.Spec.JuiceOpt != nil {
		// Deal it for juiceOpt
		s.buildJuicefsBackPod(&in)

	} else {
		// in.Containers[0].ImagePullPolicy = s.opt.ImagePullPolicy
		in.Containers[0].Args = []string{
			"request_a_backup",
			s.backup.GetBackupURL(s.backup.Spec.ClusterName, s.backup.Spec.HostName),
		}
	}
	var optTrue bool = true
	in.Containers[0].Env = []corev1.EnvVar{
		{
			Name:  "CONTAINER_TYPE",
			Value: utils.ContainerBackupJobName,
		},
		{
			Name:  "NAMESPACE",
			Value: s.backup.Namespace,
		},
		{
			Name:  "SERVICE_NAME",
			Value: fmt.Sprintf("%s-mysql", s.backup.Spec.ClusterName),
		},
		{
			Name: "HOST_NAME",

			Value: s.backup.Spec.HostName,
		},
		{
			Name:  "REPLICAS",
			Value: "1",
		},
		// backup user for sidecar http server.
		{
			Name: "BACKUP_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "backup-user",
					Optional: &optTrue,
				},
			},
		},
		// backup password for sidecar http server.
		{
			Name: "BACKUP_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "backup-password",
					Optional: &optTrue,
				},
			},
		},
		// Cluster Name for set Anotations.
		{
			Name:  "JOB_NAME",
			Value: s.job.Name,
		},
	}
	return in
}

func (s *jobSyncer) buildJuicefsBackPod(in *corev1.PodSpec) error {
	// add volumn about pvc
	var defMode int32 = 0600
	var err error
	var cmdstr string
	in.Volumes = []corev1.Volume{
		{
			Name: utils.SShVolumnName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  fmt.Sprintf("%s-ssh-key", s.backup.Spec.ClusterName),
					DefaultMode: &defMode,
				},
			},
		},
	}

	in.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      utils.SShVolumnName,
			MountPath: utils.SshVolumnPath,
		},
	}

	// PodName.clusterName-mysql.Namespace
	// sample-mysql-0.sample-mysql.default
	hostname := fmt.Sprintf("%s.%s-mysql.%s", s.backup.Spec.HostName, s.backup.Spec.ClusterName, s.backup.Namespace)
	if cmdstr, err = s.buildJuicefsCmd(s.backup.Spec.JuiceOpt.BackupSecretName); err != nil {
		return err
	}

	in.Containers[0].Command = []string{"bash", "-c", "--", `cp  /etc/secret-ssh/* /root/.ssh
chmod 600 /root/.ssh/authorized_keys ;` +
		strings.Join([]string{
			"ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", hostname, cmdstr,
		}, " ")}

	return nil
}

func (s *jobSyncer) buildJuicefsCmd(secName string) (string, error) {
	juiceopt := s.backup.Spec.JuiceOpt
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secName,
			Namespace: s.backup.Namespace,
		},
	}
	err := s.client.Get(context.TODO(),
		types.NamespacedName{Namespace: s.backup.Namespace,
			Name: secName}, secret)

	if err != nil {
		return "", err
	}
	url, bucket := secret.Data["s3-endpoint"], secret.Data["s3-bucket"]
	accesskey, secretkey := secret.Data["s3-access-key"], secret.Data["s3-secret-key"]
	juicebucket := utils.InstallBucket(string(url), string(bucket))
	cmdstr := fmt.Sprintf(`<<EOF
	export CLUSTER_NAME=%s
	juicefs format --storage s3 \
    --bucket  %s \
    --access-key %s \
    --secret-key %s \
    %s \
    %s`, s.backup.Spec.ClusterName, juicebucket, accesskey, secretkey, juiceopt.JuiceMeta, juiceopt.JuiceName)
	cmdstr += fmt.Sprintf(`
	juicefs mount -d %s /%s/
	`, juiceopt.JuiceMeta, juiceopt.JuiceName)
	cmdstr += fmt.Sprintf(`
	source /backup.sh
    backup
	juicefs umount /%s/
EOF`, juiceopt.JuiceName)
	return cmdstr, nil
}
