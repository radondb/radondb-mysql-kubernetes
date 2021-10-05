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

// metrics used for metrics container.
type metrics struct {
	*mysqlcluster.MysqlCluster

	// The name of the metrics container.
	name string
}

// getName get the container name.
func (c *metrics) getName() string {
	return c.name
}

// getImage get the container image.
func (c *metrics) getImage() string {
	return c.Spec.MetricsOpts.Image
}

// getCommand get the container command.
func (c *metrics) getCommand() []string {
	return nil
}

// getEnvVars get the container env.
func (c *metrics) getEnvVars() []corev1.EnvVar {
	return []corev1.EnvVar{
		getEnvVarFromSecret(c.GetNameForResource(utils.Secret), "DATA_SOURCE_NAME", "data-source", true),
	}
}

// getLifecycle get the container lifecycle.
func (c *metrics) getLifecycle() *corev1.Lifecycle {
	return nil
}

// getResources get the container resources.
func (c *metrics) getResources() corev1.ResourceRequirements {
	return c.Spec.MetricsOpts.Resources
}

// getPorts get the container ports.
func (c *metrics) getPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          utils.MetricsPortName,
			ContainerPort: utils.MetricsPort,
		},
	}
}

// getLivenessProbe get the container livenessProbe.
func (c *metrics) getLivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(utils.MetricsPort),
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

// getReadinessProbe get the container readinessProbe.
func (c *metrics) getReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.FromInt(utils.MetricsPort),
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

// getVolumeMounts get the container volumeMounts.
func (c *metrics) getVolumeMounts() []corev1.VolumeMount {
	return nil
}
