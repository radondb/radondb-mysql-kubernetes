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
	"fmt"

	"github.com/presslabs/controller-util/pkg/syncer"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/backup"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

type jobSyncer struct {
	job    *batchv1.Job
	backup *backup.Backup
}

// Owner returns the object owner or nil if object does not have one.
func (s *jobSyncer) ObjectOwner() runtime.Object { return s.backup.Unwrap() }

// NewJobSyncer returns a syncer for backup jobs
func NewJobSyncer(c client.Client, s *runtime.Scheme, backup *backup.Backup) syncer.Interface {
	obj := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backup.GetNameForJob(),
			Namespace: backup.Namespace,
		},
	}

	sync := &jobSyncer{
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
	in.Containers[0].Image = fmt.Sprintf("%s%s", mysqlcluster.GetPrefixFromEnv(), s.backup.Spec.Image)
	in.ServiceAccountName = s.backup.Spec.ClusterName
	if len(s.backup.Spec.NFSServerAddress) != 0 {
		// add volumn about pvc
		in.Volumes = []corev1.Volume{
			{
				Name: utils.XtrabackupPV,
				VolumeSource: corev1.VolumeSource{
					NFS: &corev1.NFSVolumeSource{
						Server: s.backup.Spec.NFSServerAddress,
						Path:   "/",
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
		in.Containers[0].Args = []string{
			fmt.Sprintf("mkdir -p /backup/%s;"+
				"curl --user $BACKUP_USER:$BACKUP_PASSWORD %s/download|xbstream -x -C /backup/%s;"+
				strAnnonations+"exit ${PIPESTATUS[0]}",
				backupToDir,
				s.backup.GetBackupURL(s.backup.Spec.ClusterName, s.backup.Spec.HostName), backupToDir),
		}
		in.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      utils.XtrabackupPV,
				MountPath: utils.XtrabckupLocal,
			},
		}
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
