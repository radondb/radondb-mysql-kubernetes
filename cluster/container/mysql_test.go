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
	"github.com/radondb/radondb-mysql-kubernetes/cluster"
)

var (
	mysqlMysqlCluster = mysqlv1alpha1.Cluster{
		Spec: mysqlv1alpha1.ClusterSpec{
			PodSpec: mysqlv1alpha1.PodSpec{
				Resources: corev1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
				},
			},
			MysqlVersion: "5.7",
			MysqlOpts: mysqlv1alpha1.MysqlOpts{
				InitTokuDB: false,
			},
		},
	}
	testMysqlCluster = cluster.Cluster{
		Cluster: &mysqlMysqlCluster,
	}
	mysqlCase = EnsureContainer("mysql", &testMysqlCluster)
)

func TestGetMysqlName(t *testing.T) {
	assert.Equal(t, "mysql", mysqlCase.Name)
}

func TestGetMysqlImage(t *testing.T) {
	assert.Equal(t, "percona/percona-server:5.7.33", mysqlCase.Image)
}

func TestGetMysqlCommand(t *testing.T) {
	assert.Nil(t, mysqlCase.Command)
}

func TestGetMysqlEnvVar(t *testing.T) {
	//base env
	{
		assert.Nil(t, mysqlCase.Env)
	}
	//initTokuDB
	{
		volumeMounts := []corev1.EnvVar{
			{
				Name:  "INIT_TOKUDB",
				Value: "1",
			},
		}
		mysqlCluster := mysqlMysqlCluster
		mysqlCluster.Spec.MysqlOpts.InitTokuDB = true
		testCluster := cluster.Cluster{
			Cluster: &mysqlCluster,
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
				Command: []string{"sh", "-c", "mysqladmin ping -uroot -p${MYSQL_ROOT_PASSWORD}"},
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
				Command: []string{"sh", "-c", `mysql -uroot -p${MYSQL_ROOT_PASSWORD} -e "SELECT 1"`},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	assert.Equal(t, readinessProbe, mysqlCase.ReadinessProbe)
}

func TestGetMysqlVolumeMounts(t *testing.T) {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "conf",
			MountPath: "/etc/mysql/conf.d",
		},
		{
			Name:      "config-map",
			MountPath: "/etc/mysql/my.cnf",
			SubPath:   "my.cnf",
		},
		{
			Name:      "data",
			MountPath: "/var/lib/mysql",
		},
		{
			Name:      "logs",
			MountPath: "/var/log/mysql",
		},
	}
	assert.Equal(t, volumeMounts, mysqlCase.VolumeMounts)
}
