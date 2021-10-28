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

	mysqlv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	initMysqlMysqlCluster = mysqlv1alpha1.MysqlCluster{
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
			MysqlVersion: "5.7",
			MysqlOpts: mysqlv1alpha1.MysqlOpts{
				RootHost:   "localhost",
				InitTokuDB: false,
			},
		},
	}
	testInitMysqlCluster = mysqlcluster.MysqlCluster{
		MysqlCluster: &initMysqlMysqlCluster,
	}
	initMysqlVolumeMounts = []corev1.VolumeMount{
		{
			Name:      utils.ConfVolumeName,
			MountPath: utils.ConfVolumeMountPath,
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
	optFalse      = false
	optTrue       = true
	sctName       = "sample-secret"
	initMysqlEnvs = []corev1.EnvVar{
		{
			Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
			Value: "yes",
		},
		{
			Name:  "MYSQL_ROOT_HOST",
			Value: "localhost",
		},
		{
			Name:  "MYSQL_INIT_ONLY",
			Value: "1",
		},
		{
			Name: "MYSQL_ROOT_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "root-password",
					Optional: &optFalse,
				},
			},
		},
	}
	initMysqlCase = EnsureContainer("init-mysql", &testInitMysqlCluster)
)

func TestGetInitMysqlName(t *testing.T) {
	assert.Equal(t, "init-mysql", initMysqlCase.Name)
}

func TestGetInitMysqlImage(t *testing.T) {
	assert.Equal(t, "percona/percona-server:5.7.34", initMysqlCase.Image)
}

func TestGetInitMysqlCommand(t *testing.T) {
	assert.Nil(t, initMysqlCase.Command)
}

func TestGetInitMysqlEnvVar(t *testing.T) {
	// base env
	{
		assert.Equal(t, initMysqlEnvs, initMysqlCase.Env)
	}
	// initTokuDB
	{
		testToKuDBMysqlCluster := initMysqlMysqlCluster
		testToKuDBMysqlCluster.Spec.MysqlOpts.InitTokuDB = true
		testTokuDBCluster := mysqlcluster.MysqlCluster{
			MysqlCluster: &testToKuDBMysqlCluster,
		}
		tokudbCase := EnsureContainer("init-mysql", &testTokuDBCluster)
		testEnv := append(initMysqlEnvs, corev1.EnvVar{
			Name:  "INIT_TOKUDB",
			Value: "1",
		})
		assert.Equal(t, testEnv, tokudbCase.Env)
	}
}

func TestGetInitMysqlLifecycle(t *testing.T) {
	assert.Nil(t, initMysqlCase.Lifecycle)
}

func TestGetInitMysqlResources(t *testing.T) {
	assert.Equal(t, corev1.ResourceRequirements{
		Limits:   nil,
		Requests: nil,
	}, initMysqlCase.Resources)
}

func TestGetInitMysqlPorts(t *testing.T) {
	assert.Nil(t, initMysqlCase.Ports)
}

func TestGetInitMysqlLivenessProbe(t *testing.T) {
	assert.Nil(t, initMysqlCase.LivenessProbe)
}

func TestGetInitMysqlReadinessProbe(t *testing.T) {
	assert.Nil(t, initMysqlCase.ReadinessProbe)
}

func TestGetInitMysqlVolumeMounts(t *testing.T) {
	assert.Equal(t, initMysqlVolumeMounts, initMysqlCase.VolumeMounts)
}
