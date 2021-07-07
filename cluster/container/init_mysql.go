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

	"github.com/radondb/radondb-mysql-kubernetes/cluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// initMysql used for init-mysql container.
type initMysql struct {
	*cluster.Cluster

	// The name of the init-mysql container.
	name string
}

// getName get the container name.
func (c *initMysql) getName() string {
	return c.name
}

// getImage get the container image.
func (c *initMysql) getImage() string {
	img := utils.MysqlImageVersions[c.GetMySQLVersion()]
	return img
}

// getCommand get the container command.
func (c *initMysql) getCommand() []string {
	return nil
}

// getEnvVars get the container env.
func (c *initMysql) getEnvVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{
		{
			Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
			Value: "yes",
		},
		{
			Name:  "MYSQL_ROOT_HOST",
			Value: "127.0.0.1",
		},
		{
			Name:  "MYSQL_INIT_ONLY",
			Value: "1",
		},
	}

	sctName := c.GetNameForResource(utils.Secret)
	envs = append(
		envs,
		getEnvVarFromSecret(sctName, "MYSQL_ROOT_PASSWORD", "root-password", false),
		getEnvVarFromSecret(sctName, "MYSQL_DATABASE", "mysql-database", true),
		getEnvVarFromSecret(sctName, "MYSQL_USER", "mysql-user", true),
		getEnvVarFromSecret(sctName, "MYSQL_PASSWORD", "mysql-password", true),
	)

	if c.Spec.MysqlOpts.InitTokuDB {
		envs = append(envs, corev1.EnvVar{
			Name:  "INIT_TOKUDB",
			Value: "1",
		})
	}

	return envs
}

// getLifecycle get the container lifecycle.
func (c *initMysql) getLifecycle() *corev1.Lifecycle {
	return nil
}

// getResources get the container resources.
func (c *initMysql) getResources() corev1.ResourceRequirements {
	return c.Spec.MysqlOpts.Resources
}

// getPorts get the container ports.
func (c *initMysql) getPorts() []corev1.ContainerPort {
	return nil
}

// getLivenessProbe get the container livenessProbe.
func (c *initMysql) getLivenessProbe() *corev1.Probe {
	return nil
}

// getReadinessProbe get the container readinessProbe.
func (c *initMysql) getReadinessProbe() *corev1.Probe {
	return nil
}

// getVolumeMounts get the container volumeMounts.
func (c *initMysql) getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      utils.ConfVolumeName,
			MountPath: utils.ConfVolumeMountPath,
		},
		{
			Name:      utils.ConfMapVolumeName,
			MountPath: utils.MyCnfMountPath,
			SubPath:   "my.cnf",
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
			Name:      utils.InitFileVolumeName,
			MountPath: utils.InitFileVolumeMountPath,
		},
	}
}
