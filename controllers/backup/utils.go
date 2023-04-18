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

package backup

import (
	"fmt"

	"github.com/radondb/radondb-mysql-kubernetes/api/v1beta1"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/rand"
)

// Define the label of backup.
const (
	labelPrefix    = "backups.mysql.radondb.com/"
	LabelCluster   = labelPrefix + "cluster"
	LableCronJob   = labelPrefix + "cronjob"
	LableManualJob = labelPrefix + "manualjob"
)

// Define the annotation of backup.
const (
	AnnotationPrefix = "backups.mysql.radondb.com/"
)

func BackupSelector(clusterName string) labels.Selector {
	return labels.SelectorFromSet(map[string]string{
		LabelCluster: clusterName,
	})
}

// jobCompleted returns "true" if the Job provided completed successfully.  Otherwise it returns
// "false".
func jobCompleted(job *batchv1.Job) bool {
	conditions := job.Status.Conditions
	for i := range conditions {
		if conditions[i].Type == batchv1.JobComplete {
			return (conditions[i].Status == corev1.ConditionTrue)
		}
	}
	return false
}

// jobFailed returns "true" if the Job provided has failed.  Otherwise it returns "false".
func jobFailed(job *batchv1.Job) bool {
	conditions := job.Status.Conditions
	for i := range conditions {
		if conditions[i].Type == batchv1.JobFailed {
			return (conditions[i].Status == corev1.ConditionTrue)
		}
	}
	return false
}

func ManualBackupJobMeta(cluster *v1beta1.MysqlCluster) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      cluster.GetName() + "-backup-" + rand.String(4),
		Namespace: cluster.GetNamespace(),
	}
}

func CronBackupJobMeta(cluster *v1beta1.MysqlCluster) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      cluster.GetName() + "-backup-" + rand.String(4),
		Namespace: cluster.GetNamespace(),
	}
}

func ManualBackupLabels(clusterName string) labels.Set {
	return map[string]string{
		LabelCluster:   clusterName,
		LableManualJob: "true",
	}
}

func CronBackupLabels(clusterName string) labels.Set {
	return map[string]string{
		LabelCluster: clusterName,
		LableCronJob: "true",
	}
}

func GetBackupHost(cluster *v1beta1.MysqlCluster) string {
	var host string
	nodeConditions := cluster.Status.Nodes
	for _, nodeCondition := range nodeConditions {
		host = nodeCondition.Name
		if nodeCondition.RaftStatus.Role == "FOLLOWER" && nodeCondition.Conditions[0].Status == "False" {
			host = nodeCondition.Name
		}
	}
	return host
}

func GetXtrabackupURL(backupHost string) string {
	xtrabackupPort := utils.XBackupPort
	url := fmt.Sprintf("%s:%d", backupHost, xtrabackupPort)
	return url
}

func getEnvVarFromSecret(sctName, name, key string, opt bool) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: sctName,
				},
				Key:      key,
				Optional: &opt,
			},
		},
	}
}
