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

	"github.com/presslabs/controller-util/syncer"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/backup"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var log = logf.Log.WithName("backup.syncer.job")

type jobSyncer struct {
	job    *batchv1.Job
	backup *backup.Backup
}

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
		log.V(1).Info("backup already completed", "backup", s.backup)
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
	}

	s.job.Spec.Template.Spec = s.ensurePodSpec(s.job.Spec.Template.Spec)
	return nil
}

func (s *jobSyncer) updateStatus(job *batchv1.Job) {
	// check for completion condition
	if cond := jobCondition(batchv1.JobComplete, job); cond != nil {
		s.backup.UpdateStatusCondition(v1alpha1.BackupComplete, cond.Status, cond.Reason, cond.Message)

		if cond.Status == corev1.ConditionTrue {
			s.backup.Status.Completed = true
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
	in.Containers[0].Image = s.backup.Spec.Image
	in.Containers[0].Args = []string{
		"request_a_backup",
		s.backup.GetBackupURL(s.backup.Spec.ClusterName, s.backup.Spec.HostName),
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
	}
	return in
}
