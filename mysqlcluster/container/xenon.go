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

// xenon used for xenon container.
type xenon struct {
	*mysqlcluster.MysqlCluster

	// The name of the xenon container.
	name string
}

// getName get the container name.
func (c *xenon) getName() string {
	return c.name
}

// getImage get the container image.
func (c *xenon) getImage() string {
	return c.Spec.XenonOpts.Image
}

// getCommand get the container command.
func (c *xenon) getCommand() []string {
	return nil
}

// getEnvVars get the container env.
func (c *xenon) getEnvVars() []corev1.EnvVar {
	autoRebuild := "false"
	if c.Spec.XenonOpts.EnableAutoRebuild {
		autoRebuild = "true"
	}
	return []corev1.EnvVar{
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
		{
			Name:  "AUTO_REBUILD",
			Value: autoRebuild,
		},
	}
}

// getLifecycle get the container lifecycle.
func (c *xenon) getLifecycle() *corev1.Lifecycle {
	return &corev1.Lifecycle{
		PreStop: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"/bin/bash",
					"-c",
					"/xenonchecker preStop",
				},
			},
		},
		PostStart: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"/bin/bash",
					"-c",
					"/xenonchecker postStart",
				},
			},
		},
	}
}

// getResources get the container resources.
func (c *xenon) getResources() corev1.ResourceRequirements {
	return c.Spec.XenonOpts.Resources
}

// getPorts get the container ports.
func (c *xenon) getPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          utils.XenonPortName,
			ContainerPort: utils.XenonPort,
		},
	}
}

// getLivenessProbe get the container livenessProbe.
func (c *xenon) getLivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"sh",
					"-c",
					"pgrep xenon && xenoncli xenon ping",
				},
			},
		},
		InitialDelaySeconds: 30,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

// getReadinessProbe get the container readinessProbe.
func (c *xenon) getReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"sh", "-c", "xenoncli xenon ping"},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

// getVolumeMounts get the container volumeMounts.
func (c *xenon) getVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      utils.ScriptsVolumeName,
			MountPath: utils.ScriptsVolumeMountPath,
		},
		{
			Name:      utils.XenonConfVolumeName,
			MountPath: utils.XenonConfVolumeMountPath,
		},
		{
			Name:      utils.XenonMetaVolumeName,
			MountPath: utils.XenonMetaVolumeMountPath,
		},
		{
			Name:      utils.SysLocalTimeZone,
			MountPath: utils.SysLocalTimeZoneMountPath,
		},
	}
}
