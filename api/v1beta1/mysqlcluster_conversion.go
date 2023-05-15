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

package v1beta1

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apiconversion "k8s.io/apimachinery/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &MysqlCluster{}

var _ conversion.Convertible = &Backup{}

func (src *MysqlCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.MysqlCluster)
	return Convert_v1beta1_MysqlCluster_To_v1alpha1_MysqlCluster(src, dst, nil)
}

func (dst *MysqlCluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.MysqlCluster)
	return Convert_v1alpha1_MysqlCluster_To_v1beta1_MysqlCluster(src, dst, nil)
}

func (src *Backup) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.Backup)
	return Convert_v1beta1_Backup_To_v1alpha1_Backup(src, dst, nil)
}

func (dst *Backup) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.Backup)
	return Convert_v1alpha1_Backup_To_v1beta1_Backup(src, dst, nil)
}

func Convert_v1alpha1_MysqlClusterSpec_To_v1beta1_MysqlClusterSpec(in *v1alpha1.MysqlClusterSpec, out *MysqlClusterSpec, s apiconversion.Scope) error {
	if err := autoConvert_v1alpha1_MysqlClusterSpec_To_v1beta1_MysqlClusterSpec(in, out, s); err != nil {
		return err
	}
	// if err := Convert_v1alpha1_MysqlOpts_To_v1beta1_MysqlConfig(in, out, s); err != nil {
	// 	return err
	// }
	//TODO in.MysqlOpts.Database in.MysqlOpts.InitTokuDB
	out.Image = in.MysqlOpts.Image
	out.MaxLagSeconds = in.MysqlOpts.MaxLagSeconds
	out.MySQLConfig.ConfigMapName = in.MysqlOpts.MysqlConfTemplate
	out.MySQLConfig.MysqlConfig = *(*map[string]string)(unsafe.Pointer(&in.MysqlOpts.MysqlConf))
	out.MySQLConfig.PluginConfig = *(*map[string]string)(unsafe.Pointer(&in.MysqlOpts.PluginConf))
	//TODO in.MysqlOpts.Password in.MysqlOpts.PluginConf in.MysqlOpts.RootHost
	out.Resources = in.MysqlOpts.Resources
	out.User = in.MysqlOpts.User
	out.Xenon = XenonOpts(in.XenonOpts)
	out.Monitoring.Exporter.Image = in.MetricsOpts.Image
	out.Monitoring.Exporter.Enabled = in.MetricsOpts.Enabled
	out.Monitoring.Exporter.Resources = in.MetricsOpts.Resources
	out.Affinity = (*corev1.Affinity)(unsafe.Pointer(in.PodPolicy.Affinity))
	out.ImagePullPolicy = in.PodPolicy.ImagePullPolicy
	out.Backup.Image = in.PodPolicy.SidecarImage
	out.Backup.Resources = in.PodPolicy.ExtraResources
	out.Log.SlowLogTail = in.PodPolicy.SlowLogTail
	out.Tolerations = in.PodPolicy.Tolerations
	out.PriorityClassName = in.PodPolicy.PriorityClassName
	out.Log.BusyboxImage = in.PodPolicy.BusyboxImage
	out.Log.Resources = in.PodPolicy.ExtraResources
	out.Storage.AccessModes = in.Persistence.AccessModes
	out.Storage.Resources.Requests = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceStorage: resource.MustParse(in.Persistence.Size),
	}
	if len(in.BackupSecretName) != 0 {
		out.DataSource.S3Backup.Name = in.RestoreFrom
		out.DataSource.S3Backup.SecretName = in.BackupSecretName
	}
	if len(in.NFSServerAddress) != 0 {
		ipStr := strings.Split(in.NFSServerAddress, ":")
		out.DataSource.NFSBackup = &NFSBackupDataSource{
			Name: in.RestoreFrom,
			Volume: corev1.NFSVolumeSource{
				Server: ipStr[0],
				Path: func() string {
					if len(ipStr) == 2 {
						return ipStr[1]
					} else {
						return "/"
					}
				}(),
			},
		}

	}
	if in.TlsSecretName != "" {
		out.CustomTLSSecret = &corev1.SecretProjection{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: in.TlsSecretName,
			},
		}
	}

	//TODO in.Backup

	return nil
}

func Convert_v1beta1_MysqlClusterSpec_To_v1alpha1_MysqlClusterSpec(in *MysqlClusterSpec, out *v1alpha1.MysqlClusterSpec, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_MysqlClusterSpec_To_v1alpha1_MysqlClusterSpec(in, out, s); err != nil {
		return err
	}
	out.MysqlOpts.User = in.User
	out.MysqlOpts.MysqlConfTemplate = in.MySQLConfig.ConfigMapName
	out.MysqlOpts.MysqlConf = in.MySQLConfig.MysqlConfig
	out.MysqlOpts.PluginConf = in.MySQLConfig.PluginConfig
	out.MysqlOpts.Resources = in.Resources
	if in.CustomTLSSecret != nil {
		out.TlsSecretName = in.CustomTLSSecret.Name
	}
	out.Persistence.StorageClass = in.Storage.StorageClassName
	out.Persistence.Size = FormatQuantity(in.Storage.Resources.Requests[corev1.ResourceStorage])
	out.Persistence.AccessModes = in.Storage.AccessModes
	out.XenonOpts = v1alpha1.XenonOpts(in.Xenon)
	// //TODO in.Backup
	out.PodPolicy.ExtraResources = in.Backup.Resources
	out.PodPolicy.SidecarImage = in.Backup.Image
	out.MetricsOpts.Image = in.Monitoring.Exporter.Image
	out.MetricsOpts.Resources = in.Monitoring.Exporter.Resources
	out.MysqlOpts.Image = in.Image
	out.PodPolicy.SlowLogTail = in.Log.SlowLogTail
	out.PodPolicy.BusyboxImage = in.Log.BusyboxImage
	out.MetricsOpts.Enabled = in.Monitoring.Exporter.Enabled
	out.PodPolicy.ImagePullPolicy = in.ImagePullPolicy
	out.PodPolicy.Tolerations = in.Tolerations
	out.PodPolicy.Affinity = (*corev1.Affinity)(unsafe.Pointer(in.Affinity))
	out.PodPolicy.PriorityClassName = in.PriorityClassName
	// in.DataSource in.Standby
	out.XenonOpts.EnableAutoRebuild = in.EnableAutoRebuild
	if len(in.DataSource.S3Backup.Name) != 0 {
		out.RestoreFrom = in.DataSource.S3Backup.Name
		out.BackupSecretName = in.DataSource.S3Backup.SecretName
	}

	if in.DataSource.NFSBackup != nil {
		out.RestoreFrom = in.DataSource.NFSBackup.Name
		out.NFSServerAddress = fmt.Sprintf("%s:%s",
			in.DataSource.NFSBackup.Volume.Server, in.DataSource.NFSBackup.Volume.Path)
	}

	//TODO in.Log n.Service
	return nil
}

func Convert_v1alpha1_BackupSpec_To_v1beta1_BackupSpec(in *v1alpha1.BackupSpec, out *BackupSpec, s apiconversion.Scope) error {
	if err := autoConvert_v1alpha1_BackupSpec_To_v1beta1_BackupSpec(in, out, s); err != nil {
		return err
	}
	return nil
}

func Convert_v1beta1_BackupSpec_To_v1alpha1_BackupSpec(in *BackupSpec, out *v1alpha1.BackupSpec, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_BackupSpec_To_v1alpha1_BackupSpec(in, out, s); err != nil {
		return err
	}
	return nil
}

func Convert_v1beta1_BackupStatus_To_v1alpha1_BackupStatus(in *BackupStatus, out *v1alpha1.BackupStatus, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_BackupStatus_To_v1alpha1_BackupStatus(in, out, s); err != nil {
		return err
	}
	return nil
}

func Convert_v1alpha1_BackupStatus_To_v1beta1_BackupStatus(in *v1alpha1.BackupStatus, out *BackupStatus, s apiconversion.Scope) error {
	if err := autoConvert_v1alpha1_BackupStatus_To_v1beta1_BackupStatus(in, out, s); err != nil {
		return err
	}
	return nil
}

func FormatQuantity(q resource.Quantity) string {
	if q.IsZero() {
		return ""
	}
	return q.String()
}
