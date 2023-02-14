package v1beta1

import (
	"github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// var _ conversion.Convertible = &Backup{}

func (src *Backup) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.Backup)
	dst.Spec.HostName = src.Spec.HostFrom
	//dst.TypeMeta = src.TypeMeta  do not copy this
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.ClusterName = src.ClusterName
	dst.Spec.HistoryLimit = src.Spec.HistoryLimit
	dst.Spec.Image = src.Spec.Image
	dst.Spec.NFSServerAddress = src.Spec.NFSServerAddress
	return nil
}
func (dst *Backup) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.Backup)
	dst.Spec.HostFrom = src.Spec.HostName
	//dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.ClusterName = src.ClusterName
	dst.Spec.HistoryLimit = src.Spec.HistoryLimit
	dst.Spec.Image = src.Spec.Image
	dst.Spec.NFSServerAddress = src.Spec.NFSServerAddress
	return nil
}
