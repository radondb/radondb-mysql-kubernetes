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
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	mysqlv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
)

var (
	metricsMysqlCluster = mysqlv1alpha1.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sample",
		},
		Spec: mysqlv1alpha1.MysqlClusterSpec{
			PodPolicy: mysqlv1alpha1.PodPolicy{
				ExtraResources: corev1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
				},
			},
			MetricsOpts: mysqlv1alpha1.MetricsOpts{
				Image: "metrics-image",
			},
		},
	}
	testMetricsCluster = mysqlcluster.MysqlCluster{
		MysqlCluster: &metricsMysqlCluster,
	}
	metricsCase = EnsureContainer("metrics", &testMetricsCluster)
)

func TestGetMetricsName(t *testing.T) {
	assert.Equal(t, "metrics", metricsCase.Name)
}

func TestGetMetricsImage(t *testing.T) {
	assert.Equal(t, "metrics-image", metricsCase.Image)
}

func TestGetMetricsCommand(t *testing.T) {
	assert.Nil(t, metricsCase.Command)
}

func TestGetMetricsEnvVar(t *testing.T) {
	{
		optTrue := true
		env := []corev1.EnvVar{
			{
				Name: "DATA_SOURCE_NAME",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "sample-secret",
						},
						Key:      "data-source",
						Optional: &optTrue,
					},
				},
			},
		}
		assert.Equal(t, env, metricsCase.Env)
	}
}

func TestGetMetricsLifecycle(t *testing.T) {
	assert.Nil(t, metricsCase.Lifecycle)
}

func TestGetMetricsResources(t *testing.T) {
	assert.Equal(t, corev1.ResourceRequirements{
		Limits:   nil,
		Requests: nil,
	}, metricsCase.Resources)
}

func TestGetMetricsPorts(t *testing.T) {
	port := []corev1.ContainerPort{
		{
			Name:          "metrics",
			ContainerPort: 9104,
		},
	}
	assert.Equal(t, port, metricsCase.Ports)
}

func TestGetMetricsLivenessProbe(t *testing.T) {
	livenessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.IntOrString{
					Type:   0,
					IntVal: int32(9104),
				},
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	assert.Equal(t, livenessProbe, metricsCase.LivenessProbe)
}

func TestGetMetricsReadinessProbe(t *testing.T) {
	readinessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/",
				Port: intstr.IntOrString{
					Type:   0,
					IntVal: int32(9104),
				},
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	assert.Equal(t, readinessProbe, metricsCase.ReadinessProbe)
}

func TestGetMetricsVolumeMounts(t *testing.T) {
	assert.Nil(t, metricsCase.VolumeMounts)
}
