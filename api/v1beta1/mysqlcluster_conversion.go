package v1beta1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &MysqlCluster{}

func (src *MysqlCluster) ConvertTo(dstRaw conversion.Hub) error {

	return nil
}
func (dst *MysqlCluster) ConvertFrom(srcRaw conversion.Hub) error {

	return nil
}
