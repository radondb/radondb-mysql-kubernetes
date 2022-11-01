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
	auditlogMysqlCluster = mysqlv1alpha1.MysqlCluster{
		Spec: mysqlv1alpha1.MysqlClusterSpec{
			PodPolicy: mysqlv1alpha1.PodPolicy{
				BusyboxImage: "busybox",
				ExtraResources: corev1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
				},
			},
		},
	}
	testAuditlogCluster = mysqlcluster.MysqlCluster{
		MysqlCluster: &auditlogMysqlCluster,
	}
	auditLogCase         = EnsureContainer("auditlog", &testAuditlogCluster)
	auditlogCommand      = []string{"sh", "-c", "for i in {120..0}; do if [ -f /var/log/mysql/mysql-audit.log ] ; then break; fi;sleep 1; done; tail -f /var/log/mysql/mysql-audit.log"}
	auditlogVolumeMounts = []corev1.VolumeMount{
		{
			Name:      "logs",
			MountPath: "/var/log/mysql",
		},
	}
)

func TestGetAuditlogName(t *testing.T) {
	assert.Equal(t, "auditlog", auditLogCase.Name)
}

func TestGetAuditlogImage(t *testing.T) {
	// assert.Equal(t, fmt.Sprintf("%s%s", mysqlcluster.GetPrefixFromEnv(), "busybox"), auditLogCase.Image)
	assert.Equal(t, mysqlcluster.GetImage("busybox"), auditLogCase.Image)
}

func TestGetAuditlogCommand(t *testing.T) {
	assert.Equal(t, auditlogCommand, auditLogCase.Command)
}

func TestGetAuditlogEnvVar(t *testing.T) {
	assert.Nil(t, auditLogCase.Env)
}

func TestGetAuditlogLifecycle(t *testing.T) {
	assert.Nil(t, auditLogCase.Lifecycle)
}

func TestGetAuditlogResources(t *testing.T) {
	assert.Equal(t, corev1.ResourceRequirements{
		Limits:   nil,
		Requests: nil,
	}, auditLogCase.Resources)
}

func TestGetAuditlogPorts(t *testing.T) {
	assert.Nil(t, auditLogCase.Ports)
}

func TestGetAuditlogLivenessProbe(t *testing.T) {
	assert.Nil(t, auditLogCase.LivenessProbe)
}

func TestGetAuditlogReadinessProbe(t *testing.T) {
	assert.Nil(t, auditLogCase.ReadinessProbe)
}

func TestGetAuditlogVolumeMounts(t *testing.T) {
	assert.Equal(t, auditlogVolumeMounts, auditLogCase.VolumeMounts)
}
