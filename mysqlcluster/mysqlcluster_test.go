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

package mysqlcluster

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	. "bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mysqlv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	mysqlCluster = mysqlv1alpha1.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sample",
		},
		Spec: mysqlv1alpha1.MysqlClusterSpec{
			MysqlVersion: "5.7",
		},
	}
	testCluster = MysqlCluster{
		&mysqlCluster,
	}
)

func TestNew(t *testing.T) {
	want := &MysqlCluster{
		&mysqlCluster,
	}
	assert.Equal(t, want, New(&mysqlCluster))
}

func TestUnwrap(t *testing.T) {
	assert.Equal(t, &mysqlCluster, testCluster.Unwrap())
}

func TestGetLabel(t *testing.T) {
	// when the instance label exist.
	{
		testMysqlCluster := mysqlCluster
		testMysqlCluster.Annotations = map[string]string{
			"app.kubernetes.io/instance": "instance",
		}
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := labels.Set{
			"mysql.radondb.com/cluster":    "sample",
			"app.kubernetes.io/name":       "mysql",
			"app.kubernetes.io/instance":   "instance",
			"app.kubernetes.io/version":    "5.7.34",
			"app.kubernetes.io/component":  "database",
			"app.kubernetes.io/managed-by": "mysql.radondb.com",
		}
		assert.Equal(t, want, testCase.GetLabels())
	}
	// when the component label exist.
	{
		testMysqlCluster := mysqlCluster
		testMysqlCluster.Annotations = map[string]string{
			"app.kubernetes.io/component": "component",
		}
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := labels.Set{
			"mysql.radondb.com/cluster":    "sample",
			"app.kubernetes.io/name":       "mysql",
			"app.kubernetes.io/instance":   "sample",
			"app.kubernetes.io/version":    "5.7.34",
			"app.kubernetes.io/component":  "component",
			"app.kubernetes.io/managed-by": "mysql.radondb.com",
		}
		assert.Equal(t, want, testCase.GetLabels())
	}
	// when the part-of label exist.
	{
		testMysqlCluster := mysqlCluster
		testMysqlCluster.Annotations = map[string]string{
			"app.kubernetes.io/part-of": "part-of",
		}
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := labels.Set{
			"mysql.radondb.com/cluster":    "sample",
			"app.kubernetes.io/name":       "mysql",
			"app.kubernetes.io/instance":   "sample",
			"app.kubernetes.io/version":    "5.7.34",
			"app.kubernetes.io/component":  "database",
			"app.kubernetes.io/managed-by": "mysql.radondb.com",
			"app.kubernetes.io/part-of":    "part-of",
		}
		assert.Equal(t, want, testCase.GetLabels())
	}
}

func TestGetSelectorLabels(t *testing.T) {
	want := labels.Set{
		"mysql.radondb.com/cluster":    "sample",
		"app.kubernetes.io/name":       "mysql",
		"app.kubernetes.io/managed-by": "mysql.radondb.com",
	}
	assert.Equal(t, want, testCluster.GetSelectorLabels())
}

func TestGetMySQLVersion(t *testing.T) {
	//other 8.0 ->  5.7.34
	{
		testMysqlCluster := mysqlCluster
		testMysqlCluster.Spec.MysqlVersion = "8.0"
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := "5.7.34"
		assert.Equal(t, want, testCase.GetMySQLVersion())
	}
	//MySQLTagsToSemVer 5.7 -> 5.7.34
	{
		testMysqlCluster := mysqlCluster
		testMysqlCluster.Spec.MysqlVersion = "5.7"
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := "5.7.34"
		assert.Equal(t, want, testCase.GetMySQLVersion())
	}
	//MysqlImageVersions 5.7.34 -> 5.7.34
	{
		testMysqlCluster := mysqlCluster
		testMysqlCluster.Spec.MysqlVersion = "5.7.34"
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := "5.7.34"
		assert.Equal(t, want, testCase.GetMySQLVersion())
	}
}

func TestCreatePeers(t *testing.T) {
	var replicas int32 = 2

	{
		testMysqlCluster := mysqlCluster
		testMysqlCluster.ObjectMeta.Namespace = "default"
		testMysqlCluster.Spec.Replicas = &replicas
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := "sample-mysql-0.sample-mysql.default:8801,sample-mysql-1.sample-mysql.default:8801"
		assert.Equal(t, want, testCase.CreatePeers())
	}
	{
		replicas = 3
		testMysqlCluster := mysqlCluster
		testMysqlCluster.ObjectMeta.Namespace = "default"
		testMysqlCluster.Spec.Replicas = &replicas
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := "sample-mysql-0.sample-mysql.default:8801,sample-mysql-1.sample-mysql.default:8801,sample-mysql-2.sample-mysql.default:8801"
		assert.Equal(t, want, testCase.CreatePeers())
	}
	{
		replicas = 666
		testMysqlCluster := mysqlCluster
		testMysqlCluster.ObjectMeta.Namespace = "default"
		testMysqlCluster.Spec.Replicas = &replicas
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := ""
		for i := 0; i < 666; i++ {
			if i > 0 {
				want += ","
			}
			want += fmt.Sprintf("%s:%d", "sample-mysql-"+strconv.Itoa(i)+".sample-mysql.default", 8801)
		}

		assert.Equal(t, want, testCase.CreatePeers())
	}
	{
		replicas = 0
		testMysqlCluster := mysqlCluster
		testMysqlCluster.ObjectMeta.Namespace = "default"
		testMysqlCluster.Spec.Replicas = &replicas
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := ""
		assert.Equal(t, want, testCase.CreatePeers())
	}
	{
		replicas = -1
		testMysqlCluster := mysqlCluster
		testMysqlCluster.ObjectMeta.Namespace = "default"
		testMysqlCluster.Spec.Replicas = &replicas
		testCase := MysqlCluster{
			&testMysqlCluster,
		}
		want := ""
		assert.Equal(t, want, testCase.CreatePeers())
	}
}

func TestGetPodHostName(t *testing.T) {
	testMysqlCluster := mysqlCluster
	testMysqlCluster.ObjectMeta.Namespace = "default"
	testCase := MysqlCluster{
		&testMysqlCluster,
	}
	want0 := "sample-mysql-0.sample-mysql.default"
	want1 := "sample-mysql-1.sample-mysql.default"
	assert.Equal(t, want0, testCase.GetPodHostName(0))
	assert.Equal(t, want1, testCase.GetPodHostName(1))
}

func TestEnsureVolumes(t *testing.T) {
	volume := []corev1.Volume{
		{
			Name: utils.ConfVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: utils.LogsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: utils.ConfMapVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "sample-mysql",
					},
				},
			},
		},
		{
			Name: utils.ScriptsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: utils.XenonVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: utils.InitFileVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	// when disable Persistence
	{
		testMysql := mysqlCluster
		testMysql.Spec.Persistence.Enabled = false
		testCase := MysqlCluster{
			&testMysql,
		}
		want := []corev1.Volume{
			{
				Name: utils.DataVolumeName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}
		want = append(want, volume...)
		assert.Equal(t, want, testCase.EnsureVolumes())
	}
	// when enable tokudb
	{
		testMysql := mysqlCluster
		testMysql.Spec.Persistence.Enabled = true
		testMysql.Spec.MysqlOpts.InitTokuDB = true
		testCase := MysqlCluster{
			&testMysql,
		}
		want := []corev1.Volume{
			{
				Name: utils.SysVolumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/sys/kernel/mm/transparent_hugepage",
					},
				},
			},
		}
		want = append(want, volume...)
		assert.Equal(t, want, testCase.EnsureVolumes())
	}
	// default(Persistence is turned on by default)
	{
		testMysql := mysqlCluster
		testMysql.Spec.Persistence.Enabled = true
		testCase := MysqlCluster{
			&testMysql,
		}
		assert.Equal(t, volume, testCase.EnsureVolumes())
	}
}

func TestEnsureVolumeClaimTemplates(t *testing.T) {
	var scheme runtime.Scheme
	// when disable persistence
	{
		result, err := testCluster.EnsureVolumeClaimTemplates(&scheme)
		assert.Nil(t, result)
		assert.Nil(t, err)
	}

	// when enable persistence
	{
		var cluster *MysqlCluster
		storageClass := "ssd"
		testMysql := mysqlv1alpha1.MysqlCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sample",
				Namespace: "default",
				Labels:    nil,
			},
			Spec: mysqlv1alpha1.MysqlClusterSpec{
				MysqlVersion: "5.7",
				Persistence: mysqlv1alpha1.Persistence{
					AccessModes:  nil,
					StorageClass: &storageClass,
					Enabled:      true,
				},
			},
		}
		testCase := MysqlCluster{
			&testMysql,
		}
		want := []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "data",
					Namespace: "default",
					Labels:    nil,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: nil,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.Quantity{},
						},
					},
					StorageClassName: &storageClass,
				},
			},
		}
		guard := PatchInstanceMethod(reflect.TypeOf(cluster), "GetLabels", func(*MysqlCluster) labels.Set {
			return nil
		})
		guard1 := Patch(resource.MustParse, func(_ string) resource.Quantity {
			return resource.Quantity{}
		})
		guard2 := Patch(controllerutil.SetControllerReference, func(_ metav1.Object, _ metav1.Object, _ *runtime.Scheme) error {
			return nil
		})
		defer guard.Unpatch()
		defer guard1.Unpatch()
		defer guard2.Unpatch()
		result, err := testCase.EnsureVolumeClaimTemplates(&scheme)
		assert.Equal(t, want, result)
		assert.Nil(t, err)
	}

	// when the StorageClass is "-"
	{
		storageClass := "-"
		testMysql := mysqlCluster
		testMysql.Spec.Persistence.Enabled = true
		testMysql.Spec.Persistence.Size = "10Gi"
		testMysql.Spec.Persistence.StorageClass = &storageClass
		testCase := MysqlCluster{
			&testMysql,
		}
		guard := Patch(controllerutil.SetControllerReference, func(_ metav1.Object, _ metav1.Object, _ *runtime.Scheme) error {
			return nil
		})
		defer guard.Unpatch()
		result, err := testCase.EnsureVolumeClaimTemplates(&scheme)

		assert.Equal(t, &storageClass, result[0].Spec.StorageClassName)
		assert.Nil(t, err)
	}

	// when SetControllerReference error
	{
		testMysql := mysqlCluster
		testMysql.Spec.Persistence.Enabled = true
		testMysql.Spec.Persistence.Size = "10Gi"
		testCase := MysqlCluster{
			&testMysql,
		}
		guard := Patch(controllerutil.SetControllerReference, func(_ metav1.Object, _ metav1.Object, _ *runtime.Scheme) error {
			return fmt.Errorf("test")
		})
		defer guard.Unpatch()
		result, err := testCase.EnsureVolumeClaimTemplates(&scheme)
		want := fmt.Errorf("failed setting controller reference: test")
		assert.Nil(t, result)
		assert.Equal(t, want, err)
	}
}

func TestGetNameForResource(t *testing.T) {
	// statefulset configMap headlessSvc
	{
		want := "sample-mysql"
		assert.Equal(t, want, testCluster.GetNameForResource(utils.StatefulSet))
		assert.Equal(t, want, testCluster.GetNameForResource(utils.ConfigMap))
		assert.Equal(t, want, testCluster.GetNameForResource(utils.HeadlessSVC))
	}
	// leaderSvc
	{
		want := "sample-leader"
		assert.Equal(t, want, testCluster.GetNameForResource(utils.LeaderService))
	}
	// folloerSvc
	{
		want := "sample-follower"
		assert.Equal(t, want, testCluster.GetNameForResource(utils.FollowerService))
	}
	// secret
	{
		want := "sample-secret"
		assert.Equal(t, want, testCluster.GetNameForResource(utils.Secret))
	}
	// others
	{
		want := "sample"
		assert.Equal(t, want, testCluster.GetNameForResource("others"))
	}
}

func TestEnsureMysqlConf(t *testing.T) {
	var (
		gb                     int64 = 1 << 30
		mb                     int64 = 1 << 20
		wantSize, wantInstance string
	)

	requestsMemory := resource.NewQuantity(gb, resource.BinarySI)
	LimitCpucorev1s := resource.NewQuantity(1, resource.DecimalSI)
	testMysql := mysqlCluster
	testMysql.Spec = mysqlv1alpha1.MysqlClusterSpec{
		MysqlOpts: mysqlv1alpha1.MysqlOpts{
			MysqlConf: mysqlv1alpha1.MysqlConf{},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"memory": *requestsMemory,
				},
				Limits: corev1.ResourceList{
					"cpu": *LimitCpucorev1s,
				},
			},
		},
	}
	// cpu 1 corev1s,memory 1 gb,innodb_buffer_pool_size not set
	{
		testMysqlCase := testMysql
		testCase := MysqlCluster{
			&testMysqlCase,
		}
		testCase.EnsureMysqlConf()
		wantSize = strconv.FormatUint(uint64(0.45*float64(gb)), 10)
		wantInstance = strconv.Itoa(int(1))
		assert.Equal(t, wantSize, testCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"])
		assert.Equal(t, wantInstance, testCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_instances"])
	}
	// cpu 1 corev1s,memory 1 gb,innodb_buffer_pool_size 600 mb
	{
		guard := Patch(sizeToBytes, func(s string) (uint64, error) {
			return uint64(600 * mb), nil
		})
		defer guard.Unpatch()

		testMysqlCase := testMysql
		testMysqlCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"] = strconv.FormatUint(uint64(600*mb), 10)
		testCase := MysqlCluster{
			&testMysqlCase,
		}
		testCase.EnsureMysqlConf()
		wantSize := strconv.FormatUint(uint64(600*float64(mb)), 10)
		wantInstance := strconv.Itoa(int(1))
		assert.Equal(t, wantSize, testCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"])
		assert.Equal(t, wantInstance, testCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_instances"])
	}
	// cpu 1 corev1s,memory 2 gb,innodb_buffer_pool_size 1.7 gb
	{
		guard := Patch(sizeToBytes, func(s string) (uint64, error) {
			return uint64(1700 * mb), nil
		})
		defer guard.Unpatch()

		memoryCase := resource.NewQuantity(2*gb, resource.BinarySI)
		testMysqlCase := testMysql
		testMysqlCase.Spec.MysqlOpts.Resources.Requests["memory"] = *memoryCase
		testMysqlCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"] = strconv.FormatUint(uint64(1.7*float64(gb)), 10)
		testCase := MysqlCluster{
			&testMysqlCase,
		}
		testCase.EnsureMysqlConf()
		wantSize := strconv.FormatUint(uint64(1.6*float64(gb)), 10)
		wantInstance := strconv.Itoa(int(1))
		assert.Equal(t, wantSize, testCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"])
		assert.Equal(t, wantInstance, testCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_instances"])
	}
	// cpu 1 corev1s,memory 2 gb,innodb_buffer_pool_size 1.7 gb, sizeToBytes error
	{
		guard := Patch(sizeToBytes, func(s string) (uint64, error) {
			return uint64(1700 * mb), fmt.Errorf("error")
		})
		defer guard.Unpatch()
		memoryCase := resource.NewQuantity(2*gb, resource.BinarySI)
		testMysqlCase := testMysql
		testMysqlCase.Spec.MysqlOpts.Resources.Requests["memory"] = *memoryCase
		testMysqlCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"] = strconv.FormatUint(uint64(1.7*float64(gb)), 10)
		testCase := MysqlCluster{
			&testMysqlCase,
		}
		testCase.EnsureMysqlConf()
		wantSize := strconv.FormatUint(uint64(1.2*float64(gb)), 10)
		wantInstance := strconv.Itoa(int(1))
		assert.Equal(t, wantSize, testCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"])
		assert.Equal(t, wantInstance, testCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_instances"])
	}
	// cpu 8 corev1s,memory 16 gb,innodb_buffer_pool_size 2 gb
	{
		guard := Patch(sizeToBytes, func(s string) (uint64, error) {
			return uint64(2 * gb), nil
		})
		defer guard.Unpatch()

		memoryCase := resource.NewQuantity(16*gb, resource.BinarySI)
		limitCpucorev1sCase := resource.NewQuantity(4, resource.DecimalSI)
		testMysqlCase := testMysql
		testMysqlCase.Spec.MysqlOpts.Resources.Limits["cpu"] = *limitCpucorev1sCase
		testMysqlCase.Spec.MysqlOpts.Resources.Requests["memory"] = *memoryCase
		testCase := MysqlCluster{
			&testMysqlCase,
		}
		testCase.EnsureMysqlConf()
		wantSize := strconv.FormatUint(uint64(2*float64(gb)), 10)
		wantInstance := strconv.Itoa(int(2))
		assert.Equal(t, wantSize, testCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"])
		assert.Equal(t, wantInstance, testCase.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_instances"])
	}
}

func TestSizeToBytes(t *testing.T) {
	var (
		gb int64 = 1 << 30
		mb int64 = 1 << 20
		kb int64 = 1 << 10
	)
	// kb
	{
		testCase := "1000k"
		want := uint64(1000 * kb)
		result, err := sizeToBytes(testCase)
		assert.Equal(t, want, result)
		assert.Nil(t, err)
	}
	// mb
	{
		testCase := "1000m"
		want := uint64(1000 * mb)
		result, err := sizeToBytes(testCase)
		assert.Equal(t, want, result)
		assert.Nil(t, err)
	}
	// gb
	{
		testCase := "1000g"
		want := uint64(1000 * gb)
		result, err := sizeToBytes(testCase)
		assert.Equal(t, want, result)
		assert.Nil(t, err)
	}
	// others
	{
		testCase := "1000a"
		want := uint64(0)
		wantError := fmt.Errorf("'1000A' format error, must be a positive integer with a unit of measurement like K, M or G")
		result, err := sizeToBytes(testCase)
		assert.Equal(t, want, result)
		assert.Equal(t, wantError, err)
	}
	// it will return the result of ParseUint() when the parameter without unit
	{
		guard := Patch(strconv.ParseUint, func(s string, base int, bitSize int) (uint64, error) {
			return uint64(666), nil
		})
		defer guard.Unpatch()

		testCase := "1000"
		want := uint64(666)
		result, err := sizeToBytes(testCase)
		assert.Equal(t, want, result)
		assert.Nil(t, err)
	}
	// ParseUint error
	{
		guard := Patch(strconv.ParseUint, func(s string, base int, bitSize int) (uint64, error) {
			return uint64(777), fmt.Errorf("error")
		})
		defer guard.Unpatch()

		testCase := "1000k"
		want := uint64(0)
		result, err := sizeToBytes(testCase)
		assert.Equal(t, want, result)
		assert.Equal(t, err, fmt.Errorf("error"))
	}
}
