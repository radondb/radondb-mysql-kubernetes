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

// auditLog used for auditlog container.
type auditLog struct {
	*mysqlcluster.MysqlCluster

	// The name of the auditlog container.
	name string
}

// getName get the container name.
func (c *auditLog) getName() string {
	return c.name
}

// getImage get the container image.
func (c *auditLog) getImage() string {
	return c.Spec.PodPolicy.BusyboxImage
}

// getCommand get the container command.
func (c *auditLog) getCommand() []string {
	logsName := "/mysql-audit.log"
	return []string{"sh", "-c", "for i in {120..0}; do if [ -f " + utils.LogsVolumeMountPath + logsName + " ] ; then break; fi;sleep 1; done; " +
		"tail -f " + utils.LogsVolumeMountPath + logsName}
}

// getEnvVars get the container env.
func (c *auditLog) getEnvVars() []corev1.EnvVar {
	return nil
}

// getLifecycle get the container lifecycle.
func (c *auditLog) getLifecycle() *corev1.Lifecycle {
	return nil
}

// getResources get the container resources.
func (c *auditLog) getResources() corev1.ResourceRequirements {
	return c.Spec.PodPolicy.ExtraResources
}

// getPorts get the container ports.
func (c *auditLog) getPorts() []corev1.ContainerPort {
	return nil
}

// getLivenessProbe get the container livenessProbe.
func (c *auditLog) getLivenessProbe() *corev1.Probe {
	return nil
}

// getReadinessProbe get the container readinessProbe.
func (c *auditLog) getReadinessProbe() *corev1.Probe {
	return nil
}

// getVolumeMounts get the container volumeMounts.
func (c *auditLog) getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      utils.LogsVolumeName,
			MountPath: utils.LogsVolumeMountPath,
		},
	}
}
