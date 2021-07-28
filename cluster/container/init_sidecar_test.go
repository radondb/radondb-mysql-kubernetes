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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mysqlv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/cluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	defeatCount             int32 = 1
	electionTimeout         int32 = 5
	replicas                int32 = 3
	initSidecarMysqlCluster       = mysqlv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sample",
		},
		Spec: mysqlv1alpha1.ClusterSpec{
			Replicas: &replicas,
			PodSpec: mysqlv1alpha1.PodSpec{
				SidecarImage: "sidecar image",
				Resources: corev1.ResourceRequirements{
					Limits:   nil,
					Requests: nil,
				},
			},
			XenonOpts: mysqlv1alpha1.XenonOpts{
				AdmitDefeatHearbeatCount: &defeatCount,
				ElectionTimeout:          &electionTimeout,
			},
			MetricsOpts: mysqlv1alpha1.MetricsOpts{
				Enabled: false,
			},
			MysqlOpts: mysqlv1alpha1.MysqlOpts{
				InitTokuDB: false,
			},
			Persistence: mysqlv1alpha1.Persistence{
				Enabled: false,
			},
		},
	}
	testInitSidecarCluster = cluster.Cluster{
		Cluster: &initSidecarMysqlCluster,
	}
	defaultInitSidecarEnvs = []corev1.EnvVar{
		{
			Name: "POD_HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name:  "NAMESPACE",
			Value: testInitSidecarCluster.Namespace,
		},
		{
			Name:  "SERVICE_NAME",
			Value: "sample-mysql",
		},
		{
			Name:  "STATEFULSET_NAME",
			Value: "sample-mysql",
		},
		{
			Name:  "REPLICAS",
			Value: fmt.Sprintf("%d", *testInitSidecarCluster.Spec.Replicas),
		},
		{
			Name:  "ADMIT_DEFEAT_HEARBEAT_COUNT",
			Value: strconv.Itoa(int(*testInitSidecarCluster.Spec.XenonOpts.AdmitDefeatHearbeatCount)),
		},
		{
			Name:  "ELECTION_TIMEOUT",
			Value: strconv.Itoa(int(*testInitSidecarCluster.Spec.XenonOpts.ElectionTimeout)),
		},
		{
			Name:  "MYSQL_VERSION",
			Value: "5.7.33",
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
		{
			Name: "MYSQL_DATABASE",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "mysql-database",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "MYSQL_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "mysql-user",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "MYSQL_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "mysql-password",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "MYSQL_REPL_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "replication-user",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "MYSQL_REPL_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "replication-password",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "METRICS_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "metrics-user",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "METRICS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "metrics-password",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "OPERATOR_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "operator-user",
					Optional: &optTrue,
				},
			},
		},
		{
			Name: "OPERATOR_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: sctName,
					},
					Key:      "operator-password",
					Optional: &optTrue,
				},
			},
		},
	}
	defaultInitsidecarVolumeMounts = []corev1.VolumeMount{
		{
			Name:      utils.ConfVolumeName,
			MountPath: utils.ConfVolumeMountPath,
		},
		{
			Name:      utils.ConfMapVolumeName,
			MountPath: utils.ConfMapVolumeMountPath,
		},
		{
			Name:      utils.ScriptsVolumeName,
			MountPath: utils.ScriptsVolumeMountPath,
		},
		{
			Name:      utils.XenonVolumeName,
			MountPath: utils.XenonVolumeMountPath,
		},
		{
			Name:      utils.InitFileVolumeName,
			MountPath: utils.InitFileVolumeMountPath,
		},
	}
	initSidecarCase = EnsureContainer("init-sidecar", &testInitSidecarCluster)
)

func TestGetInitSidecarName(t *testing.T) {
	assert.Equal(t, "init-sidecar", initSidecarCase.Name)
}

func TestGetInitSidecarImage(t *testing.T) {
	assert.Equal(t, "sidecar image", initSidecarCase.Image)
}

func TestGetInitSidecarCommand(t *testing.T) {
	command := []string{"sidecar", "init"}
	assert.Equal(t, command, initSidecarCase.Command)
}

func TestGetInitSidecarEnvVar(t *testing.T) {
	//default
	{
		assert.Equal(t, defaultInitSidecarEnvs, initSidecarCase.Env)
	}
	//initTokuDB
	{
		testToKuDBMysqlCluster := initSidecarMysqlCluster
		testToKuDBMysqlCluster.Spec.MysqlOpts.InitTokuDB = true
		testTokuDBCluster := cluster.Cluster{
			Cluster: &testToKuDBMysqlCluster,
		}
		tokudbCase := EnsureContainer("init-sidecar", &testTokuDBCluster)
		testTokuDBEnv := make([]corev1.EnvVar, 18)
		copy(testTokuDBEnv, defaultInitSidecarEnvs)
		testTokuDBEnv = append(testTokuDBEnv, corev1.EnvVar{
			Name:  "INIT_TOKUDB",
			Value: "1",
		})
		assert.Equal(t, testTokuDBEnv, tokudbCase.Env)
	}
}

func TestGetInitSidecarLifecycle(t *testing.T) {
	assert.Nil(t, initSidecarCase.Lifecycle)
}

func TestGetInitSidecarResources(t *testing.T) {
	assert.Equal(t, corev1.ResourceRequirements{
		Limits:   nil,
		Requests: nil,
	}, initSidecarCase.Resources)
}

func TestGetInitSidecarPorts(t *testing.T) {
	assert.Nil(t, initSidecarCase.Ports)
}

func TestGetInitSidecarLivenessProbe(t *testing.T) {
	assert.Nil(t, initSidecarCase.LivenessProbe)
}

func TestGetInitSidecarReadinessProbe(t *testing.T) {
	assert.Nil(t, initSidecarCase.ReadinessProbe)
}

func TestGetInitSidecarVolumeMounts(t *testing.T) {
	//default
	{
		assert.Equal(t, defaultInitsidecarVolumeMounts, initSidecarCase.VolumeMounts)
	}
	//init tokudb
	{
		testToKuDBMysqlCluster := initSidecarMysqlCluster
		testToKuDBMysqlCluster.Spec.MysqlOpts.InitTokuDB = true
		testTokuDBCluster := cluster.Cluster{
			Cluster: &testToKuDBMysqlCluster,
		}
		tokudbCase := EnsureContainer("init-sidecar", &testTokuDBCluster)
		tokuDBVolumeMounts := make([]corev1.VolumeMount, 5, 6)
		copy(tokuDBVolumeMounts, defaultInitsidecarVolumeMounts)
		tokuDBVolumeMounts = append(tokuDBVolumeMounts, corev1.VolumeMount{
			Name:      utils.SysVolumeName,
			MountPath: utils.SysVolumeMountPath,
		})
		assert.Equal(t, tokuDBVolumeMounts, tokudbCase.VolumeMounts)
	}
	//enable persistence
	{
		testPersistenceMysqlCluster := initSidecarMysqlCluster
		testPersistenceMysqlCluster.Spec.Persistence.Enabled = true
		testPersistenceCluster := cluster.Cluster{
			Cluster: &testPersistenceMysqlCluster,
		}
		persistenceCase := EnsureContainer("init-sidecar", &testPersistenceCluster)
		persistenceVolumeMounts := make([]corev1.VolumeMount, 5, 6)
		copy(persistenceVolumeMounts, defaultInitsidecarVolumeMounts)
		persistenceVolumeMounts = append(persistenceVolumeMounts, corev1.VolumeMount{
			Name:      utils.DataVolumeName,
			MountPath: utils.DataVolumeMountPath,
		})
		assert.Equal(t, persistenceVolumeMounts, persistenceCase.VolumeMounts)
	}
}
