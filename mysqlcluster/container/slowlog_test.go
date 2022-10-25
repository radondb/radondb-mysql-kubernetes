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

	mysqlv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
)

var (
	slowlogMysqlCluster = mysqlv1alpha1.MysqlCluster{
		Spec: mysqlv1alpha1.MysqlClusterSpec{
			PodPolicy: mysqlv1alpha1.PodPolicy{
				SidecarImage: "sidecar image",
				ExtraResources: corev1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
				},
			},
		},
	}
	testSlowlogCluster = mysqlcluster.MysqlCluster{
		MysqlCluster: &slowlogMysqlCluster,
	}
	slowlogCase = EnsureContainer("slowlog", &testSlowlogCluster)
)

func TestGetSlowlogName(t *testing.T) {
	assert.Equal(t, "slowlog", slowlogCase.Name)
}

func TestGetSlowlogImage(t *testing.T) {
	assert.Equal(t, mysqlcluster.GetImage("sidecar image"), slowlogCase.Image)
}

func TestGetSlowlogCommand(t *testing.T) {
	command := []string{"tail", "-f", "/var/log/mysql" + "/mysql-slow.log"}
	assert.Equal(t, command, slowlogCase.Command)
}

func TestGetSlowlogEnvVar(t *testing.T) {
	assert.Nil(t, slowlogCase.Env)
}

func TestGetSlowlogLifecycle(t *testing.T) {
	assert.Nil(t, slowlogCase.Lifecycle)
}

func TestGetSlowlogResources(t *testing.T) {
	assert.Equal(t, corev1.ResourceRequirements{
		Limits:   nil,
		Requests: nil,
	}, slowlogCase.Resources)
}

func TestGetSlowlogPorts(t *testing.T) {
	assert.Nil(t, slowlogCase.Ports)
}

func TestGetSlowlogLivenessProbe(t *testing.T) {
	assert.Nil(t, slowlogCase.LivenessProbe)
}

func TestGetSlowlogReadinessProbe(t *testing.T) {
	assert.Nil(t, slowlogCase.ReadinessProbe)
}

func TestGetSlowlogVolumeMounts(t *testing.T) {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "logs",
			MountPath: "/var/log/mysql",
		},
	}
	assert.Equal(t, volumeMounts, slowlogCase.VolumeMounts)
}
