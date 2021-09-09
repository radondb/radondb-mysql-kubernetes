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
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// Nolint: megacheck, deadcode, varcheck.
const (
	_         = iota // ignore first value by assigning to blank identifier
	kb uint64 = 1 << (10 * iota)
	mb
	gb
)

var log = logf.Log.WithName("mysqlcluster")

// MysqlCluster is the wrapper for apiv1alpha1.MysqlCluster type.
type MysqlCluster struct {
	*apiv1alpha1.MysqlCluster
}

// New returns a pointer to MysqlCluster.
func New(mc *apiv1alpha1.MysqlCluster) *MysqlCluster {
	return &MysqlCluster{
		MysqlCluster: mc,
	}
}

// Unwrap returns the api mysqlcluster object.
func (c *MysqlCluster) Unwrap() *apiv1alpha1.MysqlCluster {
	return c.MysqlCluster
}

func (c *MysqlCluster) Validate() error {
	if utils.StringInArray(c.Spec.MysqlOpts.User, []string{"root", utils.ReplicationUser, utils.OperatorUser, utils.MetricsUser}) {
		return fmt.Errorf("spec.mysqlOpts.user cannot be root|%s|%s|%s", utils.ReplicationUser, utils.OperatorUser, utils.MetricsUser)
	}

	// https://github.com/percona/percona-docker/blob/main/percona-server-5.7/ps-entry.sh#L159
	// ERROR 1396 (HY000): Operation CREATE USER failed for 'root'@'127.0.0.1'.
	if c.Spec.MysqlOpts.RootHost == "127.0.0.1" {
		return fmt.Errorf("spec.mysqlOpts.rootHost cannot be 127.0.0.1")
	}

	return nil
}

// GetLabels returns mysqlcluster labels.
func (c *MysqlCluster) GetLabels() labels.Set {
	instance := c.Name
	if inst, ok := c.Annotations["app.kubernetes.io/instance"]; ok {
		instance = inst
	}

	component := "database"
	if comp, ok := c.Annotations["app.kubernetes.io/component"]; ok {
		component = comp
	}

	labels := labels.Set{
		"mysql.radondb.com/cluster":    c.Name,
		"app.kubernetes.io/name":       "mysql",
		"app.kubernetes.io/instance":   instance,
		"app.kubernetes.io/version":    c.GetMySQLVersion(),
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/managed-by": "mysql.radondb.com",
	}

	if part, ok := c.Annotations["app.kubernetes.io/part-of"]; ok {
		labels["app.kubernetes.io/part-of"] = part
	}

	return labels
}

// GetSelectorLabels returns the labels that will be used as selector.
func (c *MysqlCluster) GetSelectorLabels() labels.Set {
	return labels.Set{
		"mysql.radondb.com/cluster":    c.Name,
		"app.kubernetes.io/name":       "mysql",
		"app.kubernetes.io/managed-by": "mysql.radondb.com",
	}
}

// GetMySQLVersion returns the MySQL server version.
func (c *MysqlCluster) GetMySQLVersion() string {
	var version string
	// Lookup for an alias, this will solve MySQL tags: 5.7 --> 5.7.x
	if v, ok := utils.MySQLTagsToSemVer[c.Spec.MysqlVersion]; ok {
		version = v
	} else {
		errmsg := "Invalid mysql version option:" + c.Spec.MysqlVersion
		log.Error(errors.New(errmsg), "currently we do not support mysql 5.6 or earlier version, default mysql version option should be 5.7 or 8.0")
		return utils.InvalidMySQLVersion
	}

	// Check if has specified image.
	if _, ok := utils.MysqlImageVersions[version]; !ok {
		version = utils.MySQLDefaultVersion
	}

	return version
}

// CreatePeers create peers for xenon.
func (c *MysqlCluster) CreatePeers() string {
	str := ""
	for i := 0; i < int(*c.Spec.Replicas); i++ {
		if i > 0 {
			str += ","
		}
		str += fmt.Sprintf("%s:%d", c.GetPodHostName(i), utils.XenonPort)
	}
	return str
}

// GetPodHostName get the pod's hostname by the index.
func (c *MysqlCluster) GetPodHostName(p int) string {
	return fmt.Sprintf("%s-%d.%s.%s", c.GetNameForResource(utils.StatefulSet), p,
		c.GetNameForResource(utils.HeadlessSVC),
		c.Namespace)
}

// EnsureVolumes ensure the volumes.
func (c *MysqlCluster) EnsureVolumes() []corev1.Volume {
	var volumes []corev1.Volume
	if !c.Spec.Persistence.Enabled {
		volumes = append(volumes, corev1.Volume{
			Name: utils.DataVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	if c.Spec.MysqlOpts.InitTokuDB {
		volumes = append(volumes,
			corev1.Volume{
				Name: utils.SysVolumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/sys/kernel/mm/transparent_hugepage",
					},
				},
			},
		)
	}

	volumes = append(volumes,
		corev1.Volume{
			Name: utils.ConfVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: utils.LogsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: utils.ConfMapVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: c.GetNameForResource(utils.ConfigMap),
					},
				},
			},
		},
		corev1.Volume{
			Name: utils.ScriptsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: utils.XenonVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: utils.InitFileVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	)

	return volumes
}

// EnsureVolumeClaimTemplates ensure the volume claim templates.
func (c *MysqlCluster) EnsureVolumeClaimTemplates(schema *runtime.Scheme) ([]corev1.PersistentVolumeClaim, error) {
	if !c.Spec.Persistence.Enabled {
		return nil, nil
	}

	if c.Spec.Persistence.StorageClass != nil {
		if *c.Spec.Persistence.StorageClass == "-" {
			*c.Spec.Persistence.StorageClass = ""
		}
	}

	data := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.DataVolumeName,
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: c.Spec.Persistence.AccessModes,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(c.Spec.Persistence.Size),
				},
			},
			StorageClassName: c.Spec.Persistence.StorageClass,
		},
	}

	if err := controllerutil.SetControllerReference(c.MysqlCluster, &data, schema); err != nil {
		return nil, fmt.Errorf("failed setting controller reference: %v", err)
	}

	return []corev1.PersistentVolumeClaim{data}, nil
}

// GetNameForResource returns the name of a resource from above
func (c *MysqlCluster) GetNameForResource(name utils.ResourceName) string {
	switch name {
	case utils.StatefulSet, utils.ConfigMap, utils.HeadlessSVC, utils.PodDisruptionBudget:
		return fmt.Sprintf("%s-mysql", c.Name)
	case utils.LeaderService:
		return fmt.Sprintf("%s-leader", c.Name)
	case utils.FollowerService:
		return fmt.Sprintf("%s-follower", c.Name)
	case utils.MetricsService:
		return fmt.Sprintf("%s-metrics", c.Name)
	case utils.Secret:
		return fmt.Sprintf("%s-secret", c.Name)
	default:
		return c.Name
	}
}

// EnsureMysqlConf set the mysql default configs.
func (c *MysqlCluster) EnsureMysqlConf() {
	if len(c.Spec.MysqlOpts.MysqlConf) == 0 {
		c.Spec.MysqlOpts.MysqlConf = make(apiv1alpha1.MysqlConf)
	}

	var defaultSize, maxSize, innodbBufferPoolSize uint64
	innodbBufferPoolSize = 128 * mb
	mem := uint64(c.Spec.MysqlOpts.Resources.Requests.Memory().Value())
	cpu := c.Spec.MysqlOpts.Resources.Limits.Cpu().MilliValue()
	if mem <= 1*gb {
		defaultSize = uint64(0.45 * float64(mem))
		maxSize = uint64(0.6 * float64(mem))
	} else {
		defaultSize = uint64(0.6 * float64(mem))
		maxSize = uint64(0.8 * float64(mem))
	}

	conf, ok := c.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"]
	if !ok {
		innodbBufferPoolSize = utils.Max(defaultSize, innodbBufferPoolSize)
	} else {
		if nums, err := sizeToBytes(conf); err != nil {
			innodbBufferPoolSize = utils.Max(defaultSize, innodbBufferPoolSize)
		} else {
			innodbBufferPoolSize = utils.Min(utils.Max(nums, innodbBufferPoolSize), maxSize)
		}
	}

	instances := math.Max(math.Min(math.Ceil(float64(cpu)/float64(1000)), math.Floor(float64(innodbBufferPoolSize)/float64(gb))), 1)
	c.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"] = strconv.FormatUint(innodbBufferPoolSize, 10)
	c.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_instances"] = strconv.Itoa(int(instances))
}

// sizeToBytes parses a string formatted by ByteSize as bytes.
// K = 1024
// M = 1024 * K
// G = 1024 * M
func sizeToBytes(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	idx := strings.IndexFunc(s, unicode.IsLetter)
	if idx == -1 {
		return strconv.ParseUint(s, 10, 64)
	}

	nums, err := strconv.ParseUint(s[:idx], 10, 64)
	if err != nil {
		return 0, err
	}

	switch s[idx:] {
	case "K":
		return nums * kb, nil
	case "M":
		return nums * mb, nil
	case "G":
		return nums * gb, nil
	}
	return 0, fmt.Errorf("'%s' format error, must be a positive integer with a unit of measurement like K, M or G", s)
}

// IsMysqlClusterKind for the given kind checks if CRD kind is for MysqlCluster CRD.
func IsClusterKind(kind string) bool {
	switch kind {
	case "MysqlCluster", "mysqlcluster", "mysqlclusters":
		return true
	}
	return false
}

// GetClusterKey returns the MysqlUser's MySQLCluster key.
func (c *MysqlCluster) GetClusterKey() client.ObjectKey {
	return client.ObjectKey{
		Name:      c.Name,
		Namespace: c.Namespace,
	}
}

// GetKey return the user key. Usually used for logging or for runtime.Client.Get as key.
func (c *MysqlCluster) GetKey() client.ObjectKey {
	return types.NamespacedName{
		Namespace: c.Namespace,
		Name:      c.Name,
	}
}
