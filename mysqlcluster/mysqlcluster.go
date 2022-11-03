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
	"math"
	"os"
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

	"github.com/go-logr/logr"

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

// MysqlCluster is the wrapper for apiv1alpha1.MysqlCluster type.
type MysqlCluster struct {
	*apiv1alpha1.MysqlCluster
	log logr.Logger
}

// New returns a pointer to MysqlCluster.
func New(mc *apiv1alpha1.MysqlCluster) *MysqlCluster {
	return &MysqlCluster{
		MysqlCluster: mc,
		log:          logf.Log.WithName("mysqlcluster"),
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
	// MySQL8 nerver support TokuDB
	// https://www.percona.com/blog/2021/05/21/tokudb-support-changes-and-future-removal-from-percona-server-for-mysql-8-0/
	if strings.Contains(c.Spec.MysqlOpts.Image, "8.0") && c.Spec.MysqlOpts.InitTokuDB {
		c.log.Info("TokuDB is not supported in MySQL 8.0 any more, the value in Cluster.spec.mysqlOpts.initTokuDB should be set false")
		return nil
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
	// version := ""
	// if _, ok := c.Labels["app.kubernetes.io/version"]; !ok {
	// 	version = c.GetMySQLVersion()
	// }
	labels := labels.Set{
		"mysql.radondb.com/cluster":  c.Name,
		"app.kubernetes.io/name":     "mysql",
		"app.kubernetes.io/instance": instance,
		// Notice: if app.kubernetes.io/version changed, then statefulset update will failure, It is not need to do this.
		// So delete this label.
		//"app.kubernetes.io/version":    version,
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

// GetMySQLVersion returns the MySQL server version. If MySQLVersion is  set, then validate the coreect verison.
func (c *MysqlCluster) GetMySQLVersion() string {
	if _, _, imageTag, err := utils.ParseImageName(c.Spec.MysqlOpts.Image); err == nil {
		return imageTag
	} else {
		return utils.InvalidMySQLVersion
	}
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
			Name: utils.MysqlConfVolumeName,
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
			Name: utils.MysqlCMVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: c.GetNameForResource(utils.ConfigMap),
					},
				},
			},
		},
		corev1.Volume{
			Name: utils.XenonCMVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: c.GetNameForResource(utils.XenonMetaData),
					},
				},
			},
		},
		corev1.Volume{
			Name: utils.XenonMetaVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: utils.ScriptsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: utils.XenonConfVolumeName,
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
		corev1.Volume{
			Name: utils.SysLocalTimeZone,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: utils.SysLocalTimeZoneHostPath,
				},
			},
		},
	)
	// add the nfs volumn mount
	if len(c.Spec.NFSServerAddress) != 0 {
		volumes = append(volumes, corev1.Volume{
			Name: utils.XtrabackupPV,
			VolumeSource: corev1.VolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: c.Spec.NFSServerAddress,
					Path:   "/",
				},
			},
		})
	}
	// Add the ssl secret mounts.
	if len(c.Spec.TlsSecretName) != 0 {
		volumes = append(volumes, corev1.Volume{
			Name: utils.TlsVolumeName + "-sidecar",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: c.Spec.TlsSecretName,
				},
			},
		}, corev1.Volume{
			Name: utils.TlsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}
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
	case utils.StatefulSet, utils.HeadlessSVC, utils.PodDisruptionBudget:
		return fmt.Sprintf("%s-mysql", c.Name)
	case utils.LeaderService:
		return fmt.Sprintf("%s-leader", c.Name)
	case utils.FollowerService:
		return fmt.Sprintf("%s-follower", c.Name)
	case utils.MetricsService:
		return fmt.Sprintf("%s-metrics", c.Name)
	case utils.Secret:
		return fmt.Sprintf("%s-secret", c.Name)
	case utils.XenonMetaData:
		return fmt.Sprintf("%s-xenon", c.Name)
	case utils.ConfigMap:
		if template := c.Spec.MysqlOpts.MysqlConfTemplate; template != "" {
			return template
		}
		return fmt.Sprintf("%s-mysql", c.Name)
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

	// innodb_log_file_size = 25 % of innodb_buffer_pool_size
	// Minimum Value (≥ 5.7.11)	4194304
	// Minimum Value (≤ 5.7.10)	1048576
	// Maximum Value	512GB / innodb_log_files_in_group
	// but wet set it not over than 8G, you should set it in the config file when you want over 8G.
	// See https://dev.mysql.com/doc/refman/5.7/en/innodb-parameters.html#sysvar_innodb_log_file_size

	const innodbDefaultLogFileSize uint64 = 1073741824
	var innodbLogFileSize uint64 = innodbDefaultLogFileSize // 1GB, default value
	// if innodb_log_file_size is not set, calculate it
	if _, ok := c.Spec.MysqlOpts.MysqlConf["innodb_log_file_size"]; !ok {
		logGroups, err := strconv.Atoi(c.Spec.MysqlOpts.MysqlConf["innodb_log_file_groups"])
		if err != nil {
			logGroups = 1
		}

		// https://dev.mysql.com/doc/refman/8.0/en/innodb-dedicated-server.html
		// Table 15.9 Automatically Configured Log File Size
		// Buffer Pool Size	Log File Size
		// Less than 8GB	512MiB
		// 8GB to 128GB	1024MiB
		// Greater than 128GB	2048MiB
		if innodbBufferPoolSize < (8 * gb) {
			innodbLogFileSize = (512 * mb) / (uint64(logGroups))
		} else if innodbBufferPoolSize <= (128 * gb) {
			innodbLogFileSize = 1 * gb
		} else {
			innodbLogFileSize = 2 * gb
		}
		// Check if the innodb_log_file_size is bigger than persistent volume size
		q, err := resource.ParseQuantity(c.Spec.Persistence.Size)
		if err != nil {
			c.log.Error(err, "failed to parse persistent volume size")
		} else {
			if uint64(q.Value()/2) < innodbLogFileSize {
				c.log.Error(err, "log file size too larger than persistent volume size")
				innodbLogFileSize = innodbDefaultLogFileSize
			}
			c.Spec.MysqlOpts.MysqlConf["innodb_log_file_size"] = strconv.FormatUint(innodbLogFileSize, 10)
		}

	}
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

// GetPrefixFromEnv get the image prefix from the environment variable.
func GetPrefixFromEnv() string {
	prefix := os.Getenv("IMAGE_PREFIX")
	if len(prefix) == 0 {
		return ""
	}
	return prefix + "/"
}

// GetImage returns the image name with the prefix and override.
func GetImage(name string) string {
	var image_namespace string
	prefix := GetPrefixFromEnv()
	override := os.Getenv("IMAGE_NAMESPACE_OVERRIDE")
	imageArray := strings.Split(name, "/")
	if len(imageArray) == 1 {
		image_namespace = ""
	} else {
		image_namespace = strings.Join(imageArray[0:len(imageArray)-1], "/") + "/"
	}
	if len(override) > 0 {
		image_namespace = override + "/"
	}

	return prefix + image_namespace + imageArray[len(imageArray)-1]
}
