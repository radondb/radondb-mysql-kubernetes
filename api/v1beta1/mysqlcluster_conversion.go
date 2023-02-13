package v1beta1

import (
	"unsafe"

	"github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &MysqlCluster{}

func (src *MysqlCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.MysqlCluster)
	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.MysqlVersion = src.Spec.MysqlVersion
	dst.Status.Conditions = *(*[]v1alpha1.ClusterCondition)(unsafe.Pointer(&src.Status.Conditions))

	dst.Status.ReadyNodes = src.Status.ReadyNodes
	dst.Status.State = v1alpha1.ClusterState(src.Status.State)
	dst.Status.Nodes = *(*[]v1alpha1.NodeStatus)(unsafe.Pointer(&src.Status.Nodes))
	return nil
}
func (dst *MysqlCluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.MysqlCluster)
	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.MysqlVersion = src.Spec.MysqlVersion

	dst.Status.Conditions = *(*[]ClusterCondition)(unsafe.Pointer(&src.Status.Conditions))
	dst.Status.ReadyNodes = src.Status.ReadyNodes
	dst.Status.State = ClusterState(src.Status.State)
	dst.Status.Nodes = *(*[]NodeStatus)(unsafe.Pointer(&src.Status.Nodes))
	return nil
}
