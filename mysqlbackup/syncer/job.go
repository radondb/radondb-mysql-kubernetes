package syncer

import (
	"fmt"

	"github.com/presslabs/controller-util/syncer"
	apiv1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlbackup"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("mysqlbackup.syncer.job")

//TODO: sync job
type jobSyncer struct {
	job    *batch.Job
	backup *mysqlbackup.MysqlBackup
}

// NewJobSyncer returns a syncer for backup jobs
func NewJobSyncer(c client.Client, s *runtime.Scheme, backup *mysqlbackup.MysqlBackup) syncer.Interface {
	obj := &batch.Job{
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
		log.V(1).Info("backup already completed", "backup", s.backup)
		// skip doing anything
		return syncer.ErrIgnore
	}

	// check if job is already created an just update the status
	if !s.job.ObjectMeta.CreationTimestamp.IsZero() {
		s.updateStatus(s.job)
		return nil
	}

	s.job.Labels = map[string]string{
		"Host": s.backup.Spec.HostName,
	}

	s.job.Spec.Template.Spec = s.ensurePodSpec(s.job.Spec.Template.Spec)
	return nil
}
func (s *jobSyncer) updateStatus(job *batch.Job) {
	// check for completion condition
	if cond := jobCondition(batch.JobComplete, job); cond != nil {
		s.backup.UpdateStatusCondition(apiv1.BackupComplete, cond.Status, cond.Reason, cond.Message)

		if cond.Status == corev1.ConditionTrue {
			s.backup.Status.Completed = true
		}
	}

	// check for failed condition
	if cond := jobCondition(batch.JobFailed, job); cond != nil {
		s.backup.UpdateStatusCondition(apiv1.BackupFailed, cond.Status, cond.Reason, cond.Message)

		if cond.Status == corev1.ConditionTrue {
			s.backup.Status.Completed = true
		}
	}
}

func jobCondition(condType batch.JobConditionType, job *batch.Job) *batch.JobCondition {
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
	// in.ImagePullSecrets = []corev1.LocalObjectReference{
	// 	{Name: s.opt.ImagePullSecretName},
	// }

	in.Containers[0].Name = "backup"
	in.Containers[0].Image = utils.SideCarImage
	if len(s.backup.Spec.BackupToPVC) != 0 {
		//add volumn about pvc
		in.Volumes = []corev1.Volume{
			{
				Name: utils.XtrabackupPV,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: s.backup.Spec.BackupToPVC,
					},
				},
			},
		}
		//"rm -rf /backup/*;curl --user sys_backups:sys_backups sample-mysql-0.sample-mysql.default:8082/download|xbstream -x -C /backup"
		in.Containers[0].Command = []string{
			"/bin/bash", "-c", "--",
		}
		var backupToDir string = utils.BuildBackupName()
		in.Containers[0].Args = []string{
			fmt.Sprintf("mkdir -p /backup/%s;curl --user sys_backups:sys_backups %s/download|xbstream -x -C /backup/%s",
				backupToDir, s.backup.GetBackupURL(s.backup.Spec.ClusterName, s.backup.Spec.HostName), backupToDir),
		}
		in.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      utils.XtrabackupPV,
				MountPath: utils.XtrabckupLocal,
			},
		}
	} else {
		//in.Containers[0].ImagePullPolicy = s.opt.ImagePullPolicy
		in.Containers[0].Args = []string{
			"request_a_backup",
			s.backup.GetBackupURL(s.backup.Spec.ClusterName, s.backup.Spec.HostName),
		}

	}

	in.Containers[0].Env = []corev1.EnvVar{
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
	}

	return in
}
