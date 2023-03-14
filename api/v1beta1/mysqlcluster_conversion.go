package v1beta1

import (
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
	//TODO in.MysqlOpts.Database in.MysqlOpts.InitTokuDB
	out.Image = in.MysqlOpts.Image
	out.MaxLagSeconds = in.MysqlOpts.MaxLagSeconds
	out.MySQLConfig = in.MysqlOpts.MysqlConfTemplate
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
	out.Log.BusyboxImage = in.PodPolicy.SidecarImage
	out.Log.Resources = in.PodPolicy.ExtraResources
	out.Storage.AccessModes = in.Persistence.AccessModes
	out.Storage.Resources.Requests = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceStorage: resource.MustParse(in.Persistence.Size),
	}

	//TODO in.Backup

	return nil
}

func Convert_v1beta1_MysqlClusterSpec_To_v1alpha1_MysqlClusterSpec(in *MysqlClusterSpec, out *v1alpha1.MysqlClusterSpec, s apiconversion.Scope) error {
	if err := autoConvert_v1beta1_MysqlClusterSpec_To_v1alpha1_MysqlClusterSpec(in, out, s); err != nil {
		return err
	}
	out.MysqlOpts.User = in.User
	out.MysqlOpts.MysqlConfTemplate = in.MySQLConfig
	out.MysqlOpts.Resources = in.Resources
	out.TlsSecretName = in.CustomTLSSecret.Name
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
	//TODO in.DataSource in.Standby
	out.XenonOpts.EnableAutoRebuild = in.EnableAutoRebuild
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
