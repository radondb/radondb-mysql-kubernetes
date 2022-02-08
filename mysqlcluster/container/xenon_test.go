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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mysqlv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	xenonReplicas     int32 = 3
	xenonMysqlCluster       = mysqlv1alpha1.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "default",
		},
		Spec: mysqlv1alpha1.MysqlClusterSpec{
			PodPolicy: mysqlv1alpha1.PodPolicy{
				ExtraResources: corev1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
				},
			},
			XenonOpts: mysqlv1alpha1.XenonOpts{
				Image: "xenon image",
			},
			Replicas: &xenonReplicas,
		},
	}
	testXenonCluster = mysqlcluster.MysqlCluster{
		MysqlCluster: &xenonMysqlCluster,
	}
	xenonCase = EnsureContainer("xenon", &testXenonCluster)
)

func TestGetXenonName(t *testing.T) {
	assert.Equal(t, "xenon", xenonCase.Name)
}

func TestGetXenonImage(t *testing.T) {
	assert.Equal(t, fmt.Sprintf("%s%s", mysqlcluster.GetPrefixFromEnv(), "xenon image"), xenonCase.Image)
}

func TestGetXenonCommand(t *testing.T) {
	assert.Nil(t, xenonCase.Command)
}

func TestGetXenonEnvVar(t *testing.T) {
	assert.Nil(t, xenonCase.Env)
}

func TestGetXenonLifecycle(t *testing.T) {
	assert.Nil(t, xenonCase.Lifecycle)
}

func TestGetXenonResources(t *testing.T) {
	assert.Equal(t, corev1.ResourceRequirements{
		Limits:   nil,
		Requests: nil,
	}, xenonCase.Resources)
}

func TestGetXenonPorts(t *testing.T) {
	port := []corev1.ContainerPort{
		{
			Name:          "xenon",
			ContainerPort: 8801,
		},
	}
	assert.Equal(t, port, xenonCase.Ports)
}

func TestGetXenonLivenessProbe(t *testing.T) {
	livenessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"pgrep", "xenon"},
			},
		},
		InitialDelaySeconds: 30,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	assert.Equal(t, livenessProbe, xenonCase.LivenessProbe)
}

func TestGetXenonReadinessProbe(t *testing.T) {
	readinessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"sh", "-c", "xenoncli xenon ping"},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	assert.Equal(t, readinessProbe, xenonCase.ReadinessProbe)
}

func TestGetXenonVolumeMounts(t *testing.T) {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "scripts",
			MountPath: "/scripts",
		},
		{
			Name:      "xenon",
			MountPath: "/etc/xenon",
		},
		{
			Name:      utils.SysLocalTimeZone,
			MountPath: "/etc/localtime",
		},
	}
	assert.Equal(t, volumeMounts, xenonCase.VolumeMounts)
}
