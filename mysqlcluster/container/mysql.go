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

	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// mysql used for mysql container.
type mysql struct {
	*mysqlcluster.MysqlCluster

	// The name of the mysql container.
	name string
}

// getName get the container name.
func (c *mysql) getName() string {
	return c.name
}

// getImage get the container image.
func (c *mysql) getImage() string {
	img := utils.MysqlImageVersions[c.GetMySQLVersion()]
	return img
}

// getCommand get the container command.
func (c *mysql) getCommand() []string {
	return []string{
		"sh",
		"-c",
		"while  [ -f '/var/lib/mysql/sleep-forever' ] ;do sleep 2 ; done; /docker-entrypoint.sh mysqld --safe-user-create --skip-symbolic-links",
	}
}

// getEnvVars get the container env.
func (c *mysql) getEnvVars() []corev1.EnvVar {
	envVar := []corev1.EnvVar{
		{
			Name: "NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
	}
	if c.Spec.MysqlOpts.InitTokuDB {
		envVar = append(envVar, corev1.EnvVar{
			Name:  "INIT_TOKUDB",
			Value: "1",
		})
	}

	return envVar
}

// getLifecycle get the container lifecycle.
func (c *mysql) getLifecycle() *corev1.Lifecycle {
	return nil
}

// getResources get the container resources.
func (c *mysql) getResources() corev1.ResourceRequirements {
	return c.Spec.MysqlOpts.Resources
}

// getPorts get the container ports.
func (c *mysql) getPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          utils.MysqlPortName,
			ContainerPort: utils.MysqlPort,
		},
	}
}

// getProbeSet get the set of livenessProbe, ReadinessProbe and StartupProbe.
func (c *mysql) getProbeSet() *ProbeSet {
	return &ProbeSet{
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{

					/* /var/lib/mysql/sleep-forever is used to prevent mysql's container from exiting.
					kubectl exec -it sample-mysql-0 -c mysql -- sh -c 'touch /var/lib/mysql/sleep-forever'
					*/
					Command: []string{
						"sh",
						"-c",
						"/mysqlchecker liveness",
					},
				},
			},
			InitialDelaySeconds: 30,
			TimeoutSeconds:      5,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			FailureThreshold:    3,
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"sh",
						"-c",
						`/mysqlchecker readiness`,
					},
				},
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      5,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			FailureThreshold:    3,
		},
		// TODO: Can choose to open/close.
		StartupProbe: &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"sh",
						"-c",
						`if test $(mysql -usuper -h127.0.0.1 -NB -e "SELECT 1") -eq 1; then cat /dev/null > /etc/mysql/conf.d/init.sql; fi`,
					},
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      5,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			FailureThreshold:    5,
		},
	}
}

// getVolumeMounts get the container volumeMounts.
func (c *mysql) getVolumeMounts() []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
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
	if c.Spec.TlsSecretName != "" {
		volumeMounts = append(volumeMounts,
			corev1.VolumeMount{
				Name:      utils.TlsVolumeName,
				MountPath: utils.TlsMountPath,
			},
		)
	}
	return volumeMounts
}
