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

package cluster

import (
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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// nolint: megacheck, deadcode, varcheck
const (
	_         = iota // ignore first value by assigning to blank identifier
	kb uint64 = 1 << (10 * iota)
	mb
	gb
)

// Cluster is the wrapper for apiv1alpha1.Cluster type.
type Cluster struct {
	*apiv1alpha1.Cluster
}

// New returns a pointer to Cluster.
func New(m *apiv1alpha1.Cluster) *Cluster {
	return &Cluster{
		Cluster: m,
	}
}

// Unwrap returns the api mysqlcluster object
func (c *Cluster) Unwrap() *apiv1alpha1.Cluster {
	return c.Cluster
}

func (c *Cluster) Validate() error {
	if utils.StringInArray(c.Spec.MysqlOpts.User, []string{"root", utils.ReplicationUser, utils.OperatorUser, utils.MetricsUser}) {
		return fmt.Errorf("spec.mysqlOpts.user cannot be root|%s|%s|%s", utils.ReplicationUser, utils.OperatorUser, utils.MetricsUser)
	}

	return nil
}

// GetLabels returns cluster labels
func (c *Cluster) GetLabels() labels.Set {
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

// GetSelectorLabels returns the labels that will be used as selector
func (c *Cluster) GetSelectorLabels() labels.Set {
	return labels.Set{
		"mysql.radondb.com/cluster":    c.Name,
		"app.kubernetes.io/name":       "mysql",
		"app.kubernetes.io/managed-by": "mysql.radondb.com",
	}
}

// GetMySQLVersion returns the MySQL server version.
func (c *Cluster) GetMySQLVersion() string {
	version := c.Spec.MysqlVersion
	// lookup for an alias, usually this will solve 5.7 to 5.7.x
	if v, ok := utils.MySQLTagsToSemVer[version]; ok {
		version = v
	}

	if _, ok := utils.MysqlImageVersions[version]; !ok {
		version = utils.MySQLDefaultVersion
	}

	return version
}

// CreatePeers create peers for xenon.
func (c *Cluster) CreatePeers() string {
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
func (c *Cluster) GetPodHostName(p int) string {
	return fmt.Sprintf("%s-%d.%s.%s", c.GetNameForResource(utils.StatefulSet), p,
		c.GetNameForResource(utils.HeadlessSVC),
		c.Namespace)
}

// EnsureVolumes ensure the volumes.
func (c *Cluster) EnsureVolumes() []corev1.Volume {
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
func (c *Cluster) EnsureVolumeClaimTemplates(schema *runtime.Scheme) ([]corev1.PersistentVolumeClaim, error) {
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

	if err := controllerutil.SetControllerReference(c.Cluster, &data, schema); err != nil {
		return nil, fmt.Errorf("failed setting controller reference: %v", err)
	}

	return []corev1.PersistentVolumeClaim{data}, nil
}

// GetNameForResource returns the name of a resource from above
func (c *Cluster) GetNameForResource(name utils.ResourceName) string {
	switch name {
	case utils.StatefulSet, utils.ConfigMap, utils.HeadlessSVC:
		return fmt.Sprintf("%s-mysql", c.Name)
	case utils.LeaderService:
		return fmt.Sprintf("%s-leader", c.Name)
	case utils.FollowerService:
		return fmt.Sprintf("%s-follower", c.Name)
	case utils.Secret:
		return fmt.Sprintf("%s-secret", c.Name)
	default:
		return c.Name
	}
}

// EnsureMysqlConf set the mysql default configs.
func (c *Cluster) EnsureMysqlConf() {
	if len(c.Spec.MysqlOpts.MysqlConf) == 0 {
		c.Spec.MysqlOpts.MysqlConf = make(apiv1alpha1.MysqlConf)
	}

	var defaultSize, maxSize, innodbBufferPoolSize uint64
	innodbBufferPoolSize = 128 * mb
	mem := uint64(c.Spec.MysqlOpts.Resources.Requests.Memory().Value())
	cpu := c.Spec.PodSpec.Resources.Limits.Cpu().MilliValue()
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
