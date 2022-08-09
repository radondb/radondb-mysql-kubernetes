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

package container

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

type backupSidecar struct {
	*mysqlcluster.MysqlCluster
	name string
}

func (c *backupSidecar) getName() string {
	return c.name
}

func (c *backupSidecar) getImage() string {
	return c.Spec.PodPolicy.SidecarImage
}

func (c *backupSidecar) getCommand() []string {
	return []string{"sidecar", "http"}
}

func (c *backupSidecar) getEnvVars() []corev1.EnvVar {
	sctNameBakup := c.Spec.BackupSecretName
	sctName := c.GetNameForResource(utils.Secret)
	envs := []corev1.EnvVar{
		{
			Name:  "CONTAINER_TYPE",
			Value: utils.ContainerBackupName,
		},
		{
			Name:  "NAMESPACE",
			Value: c.Namespace,
		},
		{
			Name:  "CLUSTER_NAME",
			Value: c.Name,
		},
		{
			Name:  "SERVICE_NAME",
			Value: c.GetNameForResource(utils.HeadlessSVC),
		},
		{
			Name:  "MYSQL_ROOT_PASSWORD",
			Value: c.Spec.MysqlOpts.RootPassword,
		},
		// backup user password for sidecar http server
		getEnvVarFromSecret(sctName, "BACKUP_USER", "backup-user", true),
		getEnvVarFromSecret(sctName, "BACKUP_PASSWORD", "backup-password", true),
	}
	if len(sctNameBakup) != 0 {
		envs = append(envs,
			getEnvVarFromSecret(sctNameBakup, "S3_ENDPOINT", "s3-endpoint", false),
			getEnvVarFromSecret(sctNameBakup, "S3_ACCESSKEY", "s3-access-key", true),
			getEnvVarFromSecret(sctNameBakup, "S3_SECRETKEY", "s3-secret-key", true),
			getEnvVarFromSecret(sctNameBakup, "S3_BUCKET", "s3-bucket", true),
		)
	}
	return envs
}

func (c *backupSidecar) getLifecycle() *corev1.Lifecycle {
	return nil
}

func (c *backupSidecar) getResources() corev1.ResourceRequirements {
	return c.Spec.PodPolicy.ExtraResources
}

func (c *backupSidecar) getPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          utils.XBackupPortName,
			ContainerPort: utils.XBackupPort,
		},
	}
}

func (c *backupSidecar) getLivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/health",
				Port: intstr.FromInt(utils.XBackupPort),
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func (c *backupSidecar) getReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/health",
				Port: intstr.FromInt(utils.XBackupPort),
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func (c *backupSidecar) getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      utils.MysqlConfVolumeName,
			MountPath: utils.MysqlConfVolumeMountPath,
		},
		{
			Name:      utils.DataVolumeName,
			MountPath: utils.DataVolumeMountPath,
		},
		{
			Name:      utils.LogsVolumeName,
			MountPath: utils.LogsVolumeMountPath,
		},
		{
			Name:      utils.SysLocalTimeZone,
			MountPath: utils.SysLocalTimeZoneMountPath,
		},
	}
}
