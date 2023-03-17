//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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
// Code generated by conversion-gen. DO NOT EDIT.

package v1beta1

import (
	unsafe "unsafe"

	v1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*Backup)(nil), (*v1alpha1.Backup)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_Backup_To_v1alpha1_Backup(a.(*Backup), b.(*v1alpha1.Backup), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.Backup)(nil), (*Backup)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Backup_To_v1beta1_Backup(a.(*v1alpha1.Backup), b.(*Backup), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*BackupList)(nil), (*v1alpha1.BackupList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_BackupList_To_v1alpha1_BackupList(a.(*BackupList), b.(*v1alpha1.BackupList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.BackupList)(nil), (*BackupList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_BackupList_To_v1beta1_BackupList(a.(*v1alpha1.BackupList), b.(*BackupList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ClusterCondition)(nil), (*v1alpha1.ClusterCondition)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_ClusterCondition_To_v1alpha1_ClusterCondition(a.(*ClusterCondition), b.(*v1alpha1.ClusterCondition), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.ClusterCondition)(nil), (*ClusterCondition)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ClusterCondition_To_v1beta1_ClusterCondition(a.(*v1alpha1.ClusterCondition), b.(*ClusterCondition), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MysqlCluster)(nil), (*v1alpha1.MysqlCluster)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_MysqlCluster_To_v1alpha1_MysqlCluster(a.(*MysqlCluster), b.(*v1alpha1.MysqlCluster), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.MysqlCluster)(nil), (*MysqlCluster)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MysqlCluster_To_v1beta1_MysqlCluster(a.(*v1alpha1.MysqlCluster), b.(*MysqlCluster), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MysqlClusterList)(nil), (*v1alpha1.MysqlClusterList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_MysqlClusterList_To_v1alpha1_MysqlClusterList(a.(*MysqlClusterList), b.(*v1alpha1.MysqlClusterList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.MysqlClusterList)(nil), (*MysqlClusterList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MysqlClusterList_To_v1beta1_MysqlClusterList(a.(*v1alpha1.MysqlClusterList), b.(*MysqlClusterList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MysqlClusterStatus)(nil), (*v1alpha1.MysqlClusterStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_MysqlClusterStatus_To_v1alpha1_MysqlClusterStatus(a.(*MysqlClusterStatus), b.(*v1alpha1.MysqlClusterStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.MysqlClusterStatus)(nil), (*MysqlClusterStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MysqlClusterStatus_To_v1beta1_MysqlClusterStatus(a.(*v1alpha1.MysqlClusterStatus), b.(*MysqlClusterStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*NodeCondition)(nil), (*v1alpha1.NodeCondition)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_NodeCondition_To_v1alpha1_NodeCondition(a.(*NodeCondition), b.(*v1alpha1.NodeCondition), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.NodeCondition)(nil), (*NodeCondition)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_NodeCondition_To_v1beta1_NodeCondition(a.(*v1alpha1.NodeCondition), b.(*NodeCondition), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*NodeStatus)(nil), (*v1alpha1.NodeStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_NodeStatus_To_v1alpha1_NodeStatus(a.(*NodeStatus), b.(*v1alpha1.NodeStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.NodeStatus)(nil), (*NodeStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_NodeStatus_To_v1beta1_NodeStatus(a.(*v1alpha1.NodeStatus), b.(*NodeStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*RaftStatus)(nil), (*v1alpha1.RaftStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_RaftStatus_To_v1alpha1_RaftStatus(a.(*RaftStatus), b.(*v1alpha1.RaftStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.RaftStatus)(nil), (*RaftStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_RaftStatus_To_v1beta1_RaftStatus(a.(*v1alpha1.RaftStatus), b.(*RaftStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*XenonOpts)(nil), (*v1alpha1.XenonOpts)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_XenonOpts_To_v1alpha1_XenonOpts(a.(*XenonOpts), b.(*v1alpha1.XenonOpts), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha1.XenonOpts)(nil), (*XenonOpts)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_XenonOpts_To_v1beta1_XenonOpts(a.(*v1alpha1.XenonOpts), b.(*XenonOpts), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*v1alpha1.BackupSpec)(nil), (*BackupSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_BackupSpec_To_v1beta1_BackupSpec(a.(*v1alpha1.BackupSpec), b.(*BackupSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*v1alpha1.BackupStatus)(nil), (*BackupStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_BackupStatus_To_v1beta1_BackupStatus(a.(*v1alpha1.BackupStatus), b.(*BackupStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*v1alpha1.MysqlClusterSpec)(nil), (*MysqlClusterSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MysqlClusterSpec_To_v1beta1_MysqlClusterSpec(a.(*v1alpha1.MysqlClusterSpec), b.(*MysqlClusterSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*BackupSpec)(nil), (*v1alpha1.BackupSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_BackupSpec_To_v1alpha1_BackupSpec(a.(*BackupSpec), b.(*v1alpha1.BackupSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*BackupStatus)(nil), (*v1alpha1.BackupStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_BackupStatus_To_v1alpha1_BackupStatus(a.(*BackupStatus), b.(*v1alpha1.BackupStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*MysqlClusterSpec)(nil), (*v1alpha1.MysqlClusterSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1beta1_MysqlClusterSpec_To_v1alpha1_MysqlClusterSpec(a.(*MysqlClusterSpec), b.(*v1alpha1.MysqlClusterSpec), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1beta1_Backup_To_v1alpha1_Backup(in *Backup, out *v1alpha1.Backup, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1beta1_BackupSpec_To_v1alpha1_BackupSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1beta1_BackupStatus_To_v1alpha1_BackupStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1beta1_Backup_To_v1alpha1_Backup is an autogenerated conversion function.
func Convert_v1beta1_Backup_To_v1alpha1_Backup(in *Backup, out *v1alpha1.Backup, s conversion.Scope) error {
	return autoConvert_v1beta1_Backup_To_v1alpha1_Backup(in, out, s)
}

func autoConvert_v1alpha1_Backup_To_v1beta1_Backup(in *v1alpha1.Backup, out *Backup, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1alpha1_BackupSpec_To_v1beta1_BackupSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_BackupStatus_To_v1beta1_BackupStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_Backup_To_v1beta1_Backup is an autogenerated conversion function.
func Convert_v1alpha1_Backup_To_v1beta1_Backup(in *v1alpha1.Backup, out *Backup, s conversion.Scope) error {
	return autoConvert_v1alpha1_Backup_To_v1beta1_Backup(in, out, s)
}

func autoConvert_v1beta1_BackupList_To_v1alpha1_BackupList(in *BackupList, out *v1alpha1.BackupList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]v1alpha1.Backup, len(*in))
		for i := range *in {
			if err := Convert_v1beta1_Backup_To_v1alpha1_Backup(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

// Convert_v1beta1_BackupList_To_v1alpha1_BackupList is an autogenerated conversion function.
func Convert_v1beta1_BackupList_To_v1alpha1_BackupList(in *BackupList, out *v1alpha1.BackupList, s conversion.Scope) error {
	return autoConvert_v1beta1_BackupList_To_v1alpha1_BackupList(in, out, s)
}

func autoConvert_v1alpha1_BackupList_To_v1beta1_BackupList(in *v1alpha1.BackupList, out *BackupList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Backup, len(*in))
		for i := range *in {
			if err := Convert_v1alpha1_Backup_To_v1beta1_Backup(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

// Convert_v1alpha1_BackupList_To_v1beta1_BackupList is an autogenerated conversion function.
func Convert_v1alpha1_BackupList_To_v1beta1_BackupList(in *v1alpha1.BackupList, out *BackupList, s conversion.Scope) error {
	return autoConvert_v1alpha1_BackupList_To_v1beta1_BackupList(in, out, s)
}

func autoConvert_v1beta1_BackupSpec_To_v1alpha1_BackupSpec(in *BackupSpec, out *v1alpha1.BackupSpec, s conversion.Scope) error {
	out.ClusterName = in.ClusterName
	// WARNING: in.BackupMethod requires manual conversion: does not exist in peer-type
	// WARNING: in.Manual requires manual conversion: does not exist in peer-type
	// WARNING: in.BackupSchedule requires manual conversion: does not exist in peer-type
	// WARNING: in.BackupOpts requires manual conversion: does not exist in peer-type
	return nil
}

func autoConvert_v1alpha1_BackupSpec_To_v1beta1_BackupSpec(in *v1alpha1.BackupSpec, out *BackupSpec, s conversion.Scope) error {
	// WARNING: in.Image requires manual conversion: does not exist in peer-type
	// WARNING: in.HostName requires manual conversion: does not exist in peer-type
	// WARNING: in.NFSServerAddress requires manual conversion: does not exist in peer-type
	out.ClusterName = in.ClusterName
	// WARNING: in.HistoryLimit requires manual conversion: does not exist in peer-type
	return nil
}

func autoConvert_v1beta1_BackupStatus_To_v1alpha1_BackupStatus(in *BackupStatus, out *v1alpha1.BackupStatus, s conversion.Scope) error {
	// WARNING: in.ManualBackup requires manual conversion: does not exist in peer-type
	// WARNING: in.ScheduledBackups requires manual conversion: does not exist in peer-type
	return nil
}

func autoConvert_v1alpha1_BackupStatus_To_v1beta1_BackupStatus(in *v1alpha1.BackupStatus, out *BackupStatus, s conversion.Scope) error {
	// WARNING: in.Completed requires manual conversion: does not exist in peer-type
	// WARNING: in.BackupName requires manual conversion: does not exist in peer-type
	// WARNING: in.BackupDate requires manual conversion: does not exist in peer-type
	// WARNING: in.BackupType requires manual conversion: does not exist in peer-type
	// WARNING: in.Conditions requires manual conversion: does not exist in peer-type
	return nil
}

func autoConvert_v1beta1_ClusterCondition_To_v1alpha1_ClusterCondition(in *ClusterCondition, out *v1alpha1.ClusterCondition, s conversion.Scope) error {
	out.Type = v1alpha1.ClusterConditionType(in.Type)
	out.Status = v1.ConditionStatus(in.Status)
	out.LastTransitionTime = in.LastTransitionTime
	out.Reason = in.Reason
	out.Message = in.Message
	return nil
}

// Convert_v1beta1_ClusterCondition_To_v1alpha1_ClusterCondition is an autogenerated conversion function.
func Convert_v1beta1_ClusterCondition_To_v1alpha1_ClusterCondition(in *ClusterCondition, out *v1alpha1.ClusterCondition, s conversion.Scope) error {
	return autoConvert_v1beta1_ClusterCondition_To_v1alpha1_ClusterCondition(in, out, s)
}

func autoConvert_v1alpha1_ClusterCondition_To_v1beta1_ClusterCondition(in *v1alpha1.ClusterCondition, out *ClusterCondition, s conversion.Scope) error {
	out.Type = ClusterConditionType(in.Type)
	out.Status = v1.ConditionStatus(in.Status)
	out.LastTransitionTime = in.LastTransitionTime
	out.Reason = in.Reason
	out.Message = in.Message
	return nil
}

// Convert_v1alpha1_ClusterCondition_To_v1beta1_ClusterCondition is an autogenerated conversion function.
func Convert_v1alpha1_ClusterCondition_To_v1beta1_ClusterCondition(in *v1alpha1.ClusterCondition, out *ClusterCondition, s conversion.Scope) error {
	return autoConvert_v1alpha1_ClusterCondition_To_v1beta1_ClusterCondition(in, out, s)
}

func autoConvert_v1beta1_MysqlCluster_To_v1alpha1_MysqlCluster(in *MysqlCluster, out *v1alpha1.MysqlCluster, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1beta1_MysqlClusterSpec_To_v1alpha1_MysqlClusterSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1beta1_MysqlClusterStatus_To_v1alpha1_MysqlClusterStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1beta1_MysqlCluster_To_v1alpha1_MysqlCluster is an autogenerated conversion function.
func Convert_v1beta1_MysqlCluster_To_v1alpha1_MysqlCluster(in *MysqlCluster, out *v1alpha1.MysqlCluster, s conversion.Scope) error {
	return autoConvert_v1beta1_MysqlCluster_To_v1alpha1_MysqlCluster(in, out, s)
}

func autoConvert_v1alpha1_MysqlCluster_To_v1beta1_MysqlCluster(in *v1alpha1.MysqlCluster, out *MysqlCluster, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1alpha1_MysqlClusterSpec_To_v1beta1_MysqlClusterSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_MysqlClusterStatus_To_v1beta1_MysqlClusterStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_MysqlCluster_To_v1beta1_MysqlCluster is an autogenerated conversion function.
func Convert_v1alpha1_MysqlCluster_To_v1beta1_MysqlCluster(in *v1alpha1.MysqlCluster, out *MysqlCluster, s conversion.Scope) error {
	return autoConvert_v1alpha1_MysqlCluster_To_v1beta1_MysqlCluster(in, out, s)
}

func autoConvert_v1beta1_MysqlClusterList_To_v1alpha1_MysqlClusterList(in *MysqlClusterList, out *v1alpha1.MysqlClusterList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]v1alpha1.MysqlCluster, len(*in))
		for i := range *in {
			if err := Convert_v1beta1_MysqlCluster_To_v1alpha1_MysqlCluster(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

// Convert_v1beta1_MysqlClusterList_To_v1alpha1_MysqlClusterList is an autogenerated conversion function.
func Convert_v1beta1_MysqlClusterList_To_v1alpha1_MysqlClusterList(in *MysqlClusterList, out *v1alpha1.MysqlClusterList, s conversion.Scope) error {
	return autoConvert_v1beta1_MysqlClusterList_To_v1alpha1_MysqlClusterList(in, out, s)
}

func autoConvert_v1alpha1_MysqlClusterList_To_v1beta1_MysqlClusterList(in *v1alpha1.MysqlClusterList, out *MysqlClusterList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MysqlCluster, len(*in))
		for i := range *in {
			if err := Convert_v1alpha1_MysqlCluster_To_v1beta1_MysqlCluster(&(*in)[i], &(*out)[i], s); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

// Convert_v1alpha1_MysqlClusterList_To_v1beta1_MysqlClusterList is an autogenerated conversion function.
func Convert_v1alpha1_MysqlClusterList_To_v1beta1_MysqlClusterList(in *v1alpha1.MysqlClusterList, out *MysqlClusterList, s conversion.Scope) error {
	return autoConvert_v1alpha1_MysqlClusterList_To_v1beta1_MysqlClusterList(in, out, s)
}

func autoConvert_v1beta1_MysqlClusterSpec_To_v1alpha1_MysqlClusterSpec(in *MysqlClusterSpec, out *v1alpha1.MysqlClusterSpec, s conversion.Scope) error {
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	// WARNING: in.User requires manual conversion: does not exist in peer-type
	// WARNING: in.MySQLConfig requires manual conversion: does not exist in peer-type
	// WARNING: in.Resources requires manual conversion: does not exist in peer-type
	// WARNING: in.CustomTLSSecret requires manual conversion: does not exist in peer-type
	// WARNING: in.Storage requires manual conversion: does not exist in peer-type
	out.MysqlVersion = in.MysqlVersion
	// WARNING: in.DatabaseInitSQL requires manual conversion: does not exist in peer-type
	// WARNING: in.Xenon requires manual conversion: does not exist in peer-type
	// WARNING: in.Backup requires manual conversion: does not exist in peer-type
	// WARNING: in.Monitoring requires manual conversion: does not exist in peer-type
	// WARNING: in.Image requires manual conversion: does not exist in peer-type
	// WARNING: in.MaxLagSeconds requires manual conversion: does not exist in peer-type
	// WARNING: in.ImagePullPolicy requires manual conversion: does not exist in peer-type
	// WARNING: in.Tolerations requires manual conversion: does not exist in peer-type
	// WARNING: in.Affinity requires manual conversion: does not exist in peer-type
	// WARNING: in.PriorityClassName requires manual conversion: does not exist in peer-type
	out.MinAvailable = in.MinAvailable
	// WARNING: in.DataSource requires manual conversion: does not exist in peer-type
	// WARNING: in.Standby requires manual conversion: does not exist in peer-type
	// WARNING: in.EnableAutoRebuild requires manual conversion: does not exist in peer-type
	// WARNING: in.Log requires manual conversion: does not exist in peer-type
	// WARNING: in.Service requires manual conversion: does not exist in peer-type
	return nil
}

func autoConvert_v1alpha1_MysqlClusterSpec_To_v1beta1_MysqlClusterSpec(in *v1alpha1.MysqlClusterSpec, out *MysqlClusterSpec, s conversion.Scope) error {
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.MinAvailable = in.MinAvailable
	// WARNING: in.MysqlOpts requires manual conversion: does not exist in peer-type
	// WARNING: in.XenonOpts requires manual conversion: does not exist in peer-type
	// WARNING: in.MetricsOpts requires manual conversion: does not exist in peer-type
	out.MysqlVersion = in.MysqlVersion
	// WARNING: in.PodPolicy requires manual conversion: does not exist in peer-type
	// WARNING: in.Persistence requires manual conversion: does not exist in peer-type
	// WARNING: in.BackupSecretName requires manual conversion: does not exist in peer-type
	// WARNING: in.RestoreFrom requires manual conversion: does not exist in peer-type
	// WARNING: in.NFSServerAddress requires manual conversion: does not exist in peer-type
	// WARNING: in.BackupSchedule requires manual conversion: does not exist in peer-type
	// WARNING: in.BothS3NFS requires manual conversion: does not exist in peer-type
	// WARNING: in.BackupScheduleJobsHistoryLimit requires manual conversion: does not exist in peer-type
	// WARNING: in.TlsSecretName requires manual conversion: does not exist in peer-type
	return nil
}

func autoConvert_v1beta1_MysqlClusterStatus_To_v1alpha1_MysqlClusterStatus(in *MysqlClusterStatus, out *v1alpha1.MysqlClusterStatus, s conversion.Scope) error {
	out.ReadyNodes = in.ReadyNodes
	out.State = v1alpha1.ClusterState(in.State)
	out.Conditions = *(*[]v1alpha1.ClusterCondition)(unsafe.Pointer(&in.Conditions))
	out.Nodes = *(*[]v1alpha1.NodeStatus)(unsafe.Pointer(&in.Nodes))
	return nil
}

// Convert_v1beta1_MysqlClusterStatus_To_v1alpha1_MysqlClusterStatus is an autogenerated conversion function.
func Convert_v1beta1_MysqlClusterStatus_To_v1alpha1_MysqlClusterStatus(in *MysqlClusterStatus, out *v1alpha1.MysqlClusterStatus, s conversion.Scope) error {
	return autoConvert_v1beta1_MysqlClusterStatus_To_v1alpha1_MysqlClusterStatus(in, out, s)
}

func autoConvert_v1alpha1_MysqlClusterStatus_To_v1beta1_MysqlClusterStatus(in *v1alpha1.MysqlClusterStatus, out *MysqlClusterStatus, s conversion.Scope) error {
	out.ReadyNodes = in.ReadyNodes
	out.State = ClusterState(in.State)
	out.Conditions = *(*[]ClusterCondition)(unsafe.Pointer(&in.Conditions))
	out.Nodes = *(*[]NodeStatus)(unsafe.Pointer(&in.Nodes))
	return nil
}

// Convert_v1alpha1_MysqlClusterStatus_To_v1beta1_MysqlClusterStatus is an autogenerated conversion function.
func Convert_v1alpha1_MysqlClusterStatus_To_v1beta1_MysqlClusterStatus(in *v1alpha1.MysqlClusterStatus, out *MysqlClusterStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_MysqlClusterStatus_To_v1beta1_MysqlClusterStatus(in, out, s)
}

func autoConvert_v1beta1_NodeCondition_To_v1alpha1_NodeCondition(in *NodeCondition, out *v1alpha1.NodeCondition, s conversion.Scope) error {
	out.Type = v1alpha1.NodeConditionType(in.Type)
	out.Status = v1.ConditionStatus(in.Status)
	out.LastTransitionTime = in.LastTransitionTime
	return nil
}

// Convert_v1beta1_NodeCondition_To_v1alpha1_NodeCondition is an autogenerated conversion function.
func Convert_v1beta1_NodeCondition_To_v1alpha1_NodeCondition(in *NodeCondition, out *v1alpha1.NodeCondition, s conversion.Scope) error {
	return autoConvert_v1beta1_NodeCondition_To_v1alpha1_NodeCondition(in, out, s)
}

func autoConvert_v1alpha1_NodeCondition_To_v1beta1_NodeCondition(in *v1alpha1.NodeCondition, out *NodeCondition, s conversion.Scope) error {
	out.Type = NodeConditionType(in.Type)
	out.Status = v1.ConditionStatus(in.Status)
	out.LastTransitionTime = in.LastTransitionTime
	return nil
}

// Convert_v1alpha1_NodeCondition_To_v1beta1_NodeCondition is an autogenerated conversion function.
func Convert_v1alpha1_NodeCondition_To_v1beta1_NodeCondition(in *v1alpha1.NodeCondition, out *NodeCondition, s conversion.Scope) error {
	return autoConvert_v1alpha1_NodeCondition_To_v1beta1_NodeCondition(in, out, s)
}

func autoConvert_v1beta1_NodeStatus_To_v1alpha1_NodeStatus(in *NodeStatus, out *v1alpha1.NodeStatus, s conversion.Scope) error {
	out.Name = in.Name
	out.Message = in.Message
	if err := Convert_v1beta1_RaftStatus_To_v1alpha1_RaftStatus(&in.RaftStatus, &out.RaftStatus, s); err != nil {
		return err
	}
	out.Conditions = *(*[]v1alpha1.NodeCondition)(unsafe.Pointer(&in.Conditions))
	return nil
}

// Convert_v1beta1_NodeStatus_To_v1alpha1_NodeStatus is an autogenerated conversion function.
func Convert_v1beta1_NodeStatus_To_v1alpha1_NodeStatus(in *NodeStatus, out *v1alpha1.NodeStatus, s conversion.Scope) error {
	return autoConvert_v1beta1_NodeStatus_To_v1alpha1_NodeStatus(in, out, s)
}

func autoConvert_v1alpha1_NodeStatus_To_v1beta1_NodeStatus(in *v1alpha1.NodeStatus, out *NodeStatus, s conversion.Scope) error {
	out.Name = in.Name
	out.Message = in.Message
	if err := Convert_v1alpha1_RaftStatus_To_v1beta1_RaftStatus(&in.RaftStatus, &out.RaftStatus, s); err != nil {
		return err
	}
	out.Conditions = *(*[]NodeCondition)(unsafe.Pointer(&in.Conditions))
	return nil
}

// Convert_v1alpha1_NodeStatus_To_v1beta1_NodeStatus is an autogenerated conversion function.
func Convert_v1alpha1_NodeStatus_To_v1beta1_NodeStatus(in *v1alpha1.NodeStatus, out *NodeStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_NodeStatus_To_v1beta1_NodeStatus(in, out, s)
}

func autoConvert_v1beta1_RaftStatus_To_v1alpha1_RaftStatus(in *RaftStatus, out *v1alpha1.RaftStatus, s conversion.Scope) error {
	out.Role = in.Role
	out.Leader = in.Leader
	out.Nodes = *(*[]string)(unsafe.Pointer(&in.Nodes))
	return nil
}

// Convert_v1beta1_RaftStatus_To_v1alpha1_RaftStatus is an autogenerated conversion function.
func Convert_v1beta1_RaftStatus_To_v1alpha1_RaftStatus(in *RaftStatus, out *v1alpha1.RaftStatus, s conversion.Scope) error {
	return autoConvert_v1beta1_RaftStatus_To_v1alpha1_RaftStatus(in, out, s)
}

func autoConvert_v1alpha1_RaftStatus_To_v1beta1_RaftStatus(in *v1alpha1.RaftStatus, out *RaftStatus, s conversion.Scope) error {
	out.Role = in.Role
	out.Leader = in.Leader
	out.Nodes = *(*[]string)(unsafe.Pointer(&in.Nodes))
	return nil
}

// Convert_v1alpha1_RaftStatus_To_v1beta1_RaftStatus is an autogenerated conversion function.
func Convert_v1alpha1_RaftStatus_To_v1beta1_RaftStatus(in *v1alpha1.RaftStatus, out *RaftStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_RaftStatus_To_v1beta1_RaftStatus(in, out, s)
}

func autoConvert_v1beta1_XenonOpts_To_v1alpha1_XenonOpts(in *XenonOpts, out *v1alpha1.XenonOpts, s conversion.Scope) error {
	out.Image = in.Image
	out.AdmitDefeatHearbeatCount = (*int32)(unsafe.Pointer(in.AdmitDefeatHearbeatCount))
	out.ElectionTimeout = (*int32)(unsafe.Pointer(in.ElectionTimeout))
	out.EnableAutoRebuild = in.EnableAutoRebuild
	out.Resources = in.Resources
	return nil
}

// Convert_v1beta1_XenonOpts_To_v1alpha1_XenonOpts is an autogenerated conversion function.
func Convert_v1beta1_XenonOpts_To_v1alpha1_XenonOpts(in *XenonOpts, out *v1alpha1.XenonOpts, s conversion.Scope) error {
	return autoConvert_v1beta1_XenonOpts_To_v1alpha1_XenonOpts(in, out, s)
}

func autoConvert_v1alpha1_XenonOpts_To_v1beta1_XenonOpts(in *v1alpha1.XenonOpts, out *XenonOpts, s conversion.Scope) error {
	out.Image = in.Image
	out.AdmitDefeatHearbeatCount = (*int32)(unsafe.Pointer(in.AdmitDefeatHearbeatCount))
	out.ElectionTimeout = (*int32)(unsafe.Pointer(in.ElectionTimeout))
	out.EnableAutoRebuild = in.EnableAutoRebuild
	out.Resources = in.Resources
	return nil
}

// Convert_v1alpha1_XenonOpts_To_v1beta1_XenonOpts is an autogenerated conversion function.
func Convert_v1alpha1_XenonOpts_To_v1beta1_XenonOpts(in *v1alpha1.XenonOpts, out *XenonOpts, s conversion.Scope) error {
	return autoConvert_v1alpha1_XenonOpts_To_v1beta1_XenonOpts(in, out, s)
}
