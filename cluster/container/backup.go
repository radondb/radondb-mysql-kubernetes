package container

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/radondb/radondb-mysql-kubernetes/cluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

type backupSidecar struct {
	*cluster.Cluster

	name string
}

func (c *backupSidecar) getName() string {
	return c.name
}

func (c *backupSidecar) getImage() string {
	return c.Spec.PodSpec.SidecarImage
}

func (c *backupSidecar) getCommand() []string {
	return []string{"sidecar", "http"}
}

func (c *backupSidecar) getEnvVars() []corev1.EnvVar {
	sctName := c.GetNameForResource(utils.BackupSecret)

	envs := []corev1.EnvVar{

		{
			Name:  "NAMESPACE",
			Value: c.Namespace,
		},
		{
			Name:  "SERVICE_NAME",
			Value: c.GetNameForResource(utils.HeadlessSVC),
		},
		getEnvVarFromSecret(sctName, "S3_ENDPOINT", "s3-endpoint", false),
		getEnvVarFromSecret(sctName, "S3_ACCESSKEY", "s3-access-key", true),
		getEnvVarFromSecret(sctName, "S3_SECRETKEY", "s3-secret-key", true),
		getEnvVarFromSecret(sctName, "S3_BUCKET", "s3-bucket", true),
	}

	return envs
}

func (c *backupSidecar) getLifecycle() *corev1.Lifecycle {
	return nil
}

func (c *backupSidecar) getResources() corev1.ResourceRequirements {
	return c.Spec.PodSpec.Resources
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
	return nil
}

func (c *backupSidecar) getReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/health",
				Port: intstr.FromInt(utils.XBackupPort),
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func (c *backupSidecar) getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      utils.ConfVolumeName,
			MountPath: utils.ConfVolumeMountPath,
		},
		{
			Name:      utils.DataVolumeName,
			MountPath: utils.DataVolumeMountPath,
		},
		{
			Name:      utils.LogsVolumeName,
			MountPath: utils.LogsVolumeMountPath,
		},
	}

}
