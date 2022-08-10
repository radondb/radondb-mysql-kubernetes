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

	mysqlv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	mysqlMysqlCluster = mysqlv1alpha1.MysqlCluster{
		Spec: mysqlv1alpha1.MysqlClusterSpec{
			PodPolicy: mysqlv1alpha1.PodPolicy{
				ExtraResources: corev1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
				},
			},
			MysqlOpts: mysqlv1alpha1.MysqlOpts{
				InitTokuDB: false,
				Image:      "percona/percona-server:5.7.34",
			},
		},
	}
	testMysqlCluster = mysqlcluster.MysqlCluster{
		MysqlCluster: &mysqlMysqlCluster,
	}
	mysqlCase = EnsureContainer("mysql", &testMysqlCluster)
)

func TestGetMysqlName(t *testing.T) {
	assert.Equal(t, "mysql", mysqlCase.Name)
}

func TestGetMysqlImage(t *testing.T) {
	assert.Equal(t, fmt.Sprintf("%s%s", mysqlcluster.GetPrefixFromEnv(), "percona/percona-server:5.7.34"), mysqlCase.Image)
}

func TestGetMysqlCommand(t *testing.T) {
	assert.Equal(t, mysqlCase.Command,
		[]string{"sh", "-c", "while  [ -f '/var/lib/mysql/sleep-forever' ] ;do sleep 2 ; done; /docker-entrypoint.sh mysqld"})
}

func TestGetMysqlEnvVar(t *testing.T) {
	// base env
	{
		assert.Nil(t, mysqlCase.Env)
	}
	// initTokuDB
	{
		volumeMounts := []corev1.EnvVar{
			{
				Name:  "INIT_TOKUDB",
				Value: "1",
			},
		}
		mysqlCluster := mysqlMysqlCluster
		mysqlCluster.Spec.MysqlOpts.InitTokuDB = true
		testCluster := mysqlcluster.MysqlCluster{
			MysqlCluster: &mysqlCluster,
		}
		mysqlCase = EnsureContainer("mysql", &testCluster)
		assert.Equal(t, volumeMounts, mysqlCase.Env)
	}
}

func TestGetMysqlLifecycle(t *testing.T) {
	assert.Nil(t, mysqlCase.Lifecycle)
}

func TestGetMysqlResources(t *testing.T) {
	assert.Equal(t, corev1.ResourceRequirements{
		Limits:   nil,
		Requests: nil,
	}, mysqlCase.Resources)
}

func TestGetMysqlPorts(t *testing.T) {
	port := []corev1.ContainerPort{
		{
			Name:          "mysql",
			ContainerPort: 3306,
		},
	}
	assert.Equal(t, port, mysqlCase.Ports)
}

func TestGetMysqlLivenessProbe(t *testing.T) {
	livenessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"sh", "-c", "if [ -f '/var/lib/mysql/sleep-forever' ] ;then exit 0 ; fi; pgrep mysqld"},
			},
		},
		InitialDelaySeconds: 30,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	assert.Equal(t, livenessProbe, mysqlCase.LivenessProbe)
}

func TestGetMysqlReadinessProbe(t *testing.T) {
	readinessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"sh", "-c", `if [ -f '/var/lib/mysql/sleep-forever' ] ;then exit 0 ; fi; test $(mysql --defaults-file=/etc/mysql/client.conf -NB -e "SELECT 1") -eq 1`},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	assert.Equal(t, readinessProbe, mysqlCase.ReadinessProbe)
}

func TestGetMysqlVolumeMounts(t *testing.T) {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "mysql-conf",
			MountPath: "/etc/mysql",
		},
		{
			Name:      "data",
			MountPath: "/var/lib/mysql",
		},
		{
			Name:      "logs",
			MountPath: "/var/log/mysql",
		},
		{
			Name:      utils.SysLocalTimeZone,
			MountPath: "/etc/localtime",
		},
	}
	assert.Equal(t, volumeMounts, mysqlCase.VolumeMounts)
}
