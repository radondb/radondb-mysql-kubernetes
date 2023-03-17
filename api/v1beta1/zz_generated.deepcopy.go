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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Backup) DeepCopyInto(out *Backup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Backup.
func (in *Backup) DeepCopy() *Backup {
	if in == nil {
		return nil
	}
	out := new(Backup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Backup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupList) DeepCopyInto(out *BackupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Backup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupList.
func (in *BackupList) DeepCopy() *BackupList {
	if in == nil {
		return nil
	}
	out := new(BackupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BackupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupOps) DeepCopyInto(out *BackupOps) {
	*out = *in
	if in.S3 != nil {
		in, out := &in.S3, &out.S3
		*out = new(S3)
		**out = **in
	}
	if in.NFS != nil {
		in, out := &in.NFS, &out.NFS
		*out = new(NFS)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupOps.
func (in *BackupOps) DeepCopy() *BackupOps {
	if in == nil {
		return nil
	}
	out := new(BackupOps)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupOpts) DeepCopyInto(out *BackupOpts) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupOpts.
func (in *BackupOpts) DeepCopy() *BackupOpts {
	if in == nil {
		return nil
	}
	out := new(BackupOpts)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupSchedule) DeepCopyInto(out *BackupSchedule) {
	*out = *in
	if in.BackupRetention != nil {
		in, out := &in.BackupRetention, &out.BackupRetention
		*out = new(int32)
		**out = **in
	}
	if in.BackupJobHistoryLimit != nil {
		in, out := &in.BackupJobHistoryLimit, &out.BackupJobHistoryLimit
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupSchedule.
func (in *BackupSchedule) DeepCopy() *BackupSchedule {
	if in == nil {
		return nil
	}
	out := new(BackupSchedule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupSpec) DeepCopyInto(out *BackupSpec) {
	*out = *in
	if in.Manual != nil {
		in, out := &in.Manual, &out.Manual
		*out = new(ManualBackup)
		(*in).DeepCopyInto(*out)
	}
	if in.BackupSchedule != nil {
		in, out := &in.BackupSchedule, &out.BackupSchedule
		*out = new(BackupSchedule)
		(*in).DeepCopyInto(*out)
	}
	in.BackupOpts.DeepCopyInto(&out.BackupOpts)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupSpec.
func (in *BackupSpec) DeepCopy() *BackupSpec {
	if in == nil {
		return nil
	}
	out := new(BackupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupStatus) DeepCopyInto(out *BackupStatus) {
	*out = *in
	if in.ManualBackup != nil {
		in, out := &in.ManualBackup, &out.ManualBackup
		*out = new(ManualBackupStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.ScheduledBackups != nil {
		in, out := &in.ScheduledBackups, &out.ScheduledBackups
		*out = make([]ScheduledBackupStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupStatus.
func (in *BackupStatus) DeepCopy() *BackupStatus {
	if in == nil {
		return nil
	}
	out := new(BackupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterCondition) DeepCopyInto(out *ClusterCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterCondition.
func (in *ClusterCondition) DeepCopy() *ClusterCondition {
	if in == nil {
		return nil
	}
	out := new(ClusterCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataSource) DeepCopyInto(out *DataSource) {
	*out = *in
	if in.Remote != nil {
		in, out := &in.Remote, &out.Remote
		*out = new(RemoteDataSource)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataSource.
func (in *DataSource) DeepCopy() *DataSource {
	if in == nil {
		return nil
	}
	out := new(DataSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatabaseInitSQL) DeepCopyInto(out *DatabaseInitSQL) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatabaseInitSQL.
func (in *DatabaseInitSQL) DeepCopy() *DatabaseInitSQL {
	if in == nil {
		return nil
	}
	out := new(DatabaseInitSQL)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExporterSpec) DeepCopyInto(out *ExporterSpec) {
	*out = *in
	if in.CustomTLSSecret != nil {
		in, out := &in.CustomTLSSecret, &out.CustomTLSSecret
		*out = new(v1.SecretProjection)
		(*in).DeepCopyInto(*out)
	}
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExporterSpec.
func (in *ExporterSpec) DeepCopy() *ExporterSpec {
	if in == nil {
		return nil
	}
	out := new(ExporterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogOpts) DeepCopyInto(out *LogOpts) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogOpts.
func (in *LogOpts) DeepCopy() *LogOpts {
	if in == nil {
		return nil
	}
	out := new(LogOpts)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManualBackup) DeepCopyInto(out *ManualBackup) {
	*out = *in
	if in.BackupRetention != nil {
		in, out := &in.BackupRetention, &out.BackupRetention
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManualBackup.
func (in *ManualBackup) DeepCopy() *ManualBackup {
	if in == nil {
		return nil
	}
	out := new(ManualBackup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManualBackupStatus) DeepCopyInto(out *ManualBackupStatus) {
	*out = *in
	if in.StartTime != nil {
		in, out := &in.StartTime, &out.StartTime
		*out = (*in).DeepCopy()
	}
	if in.CompletionTime != nil {
		in, out := &in.CompletionTime, &out.CompletionTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManualBackupStatus.
func (in *ManualBackupStatus) DeepCopy() *ManualBackupStatus {
	if in == nil {
		return nil
	}
	out := new(ManualBackupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MonitoringSpec) DeepCopyInto(out *MonitoringSpec) {
	*out = *in
	in.Exporter.DeepCopyInto(&out.Exporter)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MonitoringSpec.
func (in *MonitoringSpec) DeepCopy() *MonitoringSpec {
	if in == nil {
		return nil
	}
	out := new(MonitoringSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MySQLStandbySpec) DeepCopyInto(out *MySQLStandbySpec) {
	*out = *in
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MySQLStandbySpec.
func (in *MySQLStandbySpec) DeepCopy() *MySQLStandbySpec {
	if in == nil {
		return nil
	}
	out := new(MySQLStandbySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MysqlCluster) DeepCopyInto(out *MysqlCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MysqlCluster.
func (in *MysqlCluster) DeepCopy() *MysqlCluster {
	if in == nil {
		return nil
	}
	out := new(MysqlCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MysqlCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MysqlClusterList) DeepCopyInto(out *MysqlClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MysqlCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MysqlClusterList.
func (in *MysqlClusterList) DeepCopy() *MysqlClusterList {
	if in == nil {
		return nil
	}
	out := new(MysqlClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MysqlClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MysqlClusterSpec) DeepCopyInto(out *MysqlClusterSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	in.Resources.DeepCopyInto(&out.Resources)
	in.CustomTLSSecret.DeepCopyInto(&out.CustomTLSSecret)
	in.Storage.DeepCopyInto(&out.Storage)
	if in.DatabaseInitSQL != nil {
		in, out := &in.DatabaseInitSQL, &out.DatabaseInitSQL
		*out = new(DatabaseInitSQL)
		**out = **in
	}
	in.Xenon.DeepCopyInto(&out.Xenon)
	in.Backup.DeepCopyInto(&out.Backup)
	in.Monitoring.DeepCopyInto(&out.Monitoring)
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]v1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(v1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.DataSource != nil {
		in, out := &in.DataSource, &out.DataSource
		*out = new(DataSource)
		(*in).DeepCopyInto(*out)
	}
	if in.Standby != nil {
		in, out := &in.Standby, &out.Standby
		*out = new(MySQLStandbySpec)
		(*in).DeepCopyInto(*out)
	}
	in.Log.DeepCopyInto(&out.Log)
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ServiceSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MysqlClusterSpec.
func (in *MysqlClusterSpec) DeepCopy() *MysqlClusterSpec {
	if in == nil {
		return nil
	}
	out := new(MysqlClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MysqlClusterStatus) DeepCopyInto(out *MysqlClusterStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ClusterCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]NodeStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MysqlClusterStatus.
func (in *MysqlClusterStatus) DeepCopy() *MysqlClusterStatus {
	if in == nil {
		return nil
	}
	out := new(MysqlClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NFS) DeepCopyInto(out *NFS) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NFS.
func (in *NFS) DeepCopy() *NFS {
	if in == nil {
		return nil
	}
	out := new(NFS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeCondition) DeepCopyInto(out *NodeCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeCondition.
func (in *NodeCondition) DeepCopy() *NodeCondition {
	if in == nil {
		return nil
	}
	out := new(NodeCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeStatus) DeepCopyInto(out *NodeStatus) {
	*out = *in
	in.RaftStatus.DeepCopyInto(&out.RaftStatus)
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]NodeCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeStatus.
func (in *NodeStatus) DeepCopy() *NodeStatus {
	if in == nil {
		return nil
	}
	out := new(NodeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RaftStatus) DeepCopyInto(out *RaftStatus) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RaftStatus.
func (in *RaftStatus) DeepCopy() *RaftStatus {
	if in == nil {
		return nil
	}
	out := new(RaftStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RemoteDataSource) DeepCopyInto(out *RemoteDataSource) {
	*out = *in
	if in.SourceConfig != nil {
		in, out := &in.SourceConfig, &out.SourceConfig
		*out = new(v1.SecretProjection)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RemoteDataSource.
func (in *RemoteDataSource) DeepCopy() *RemoteDataSource {
	if in == nil {
		return nil
	}
	out := new(RemoteDataSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *S3) DeepCopyInto(out *S3) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new S3.
func (in *S3) DeepCopy() *S3 {
	if in == nil {
		return nil
	}
	out := new(S3)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScheduledBackupStatus) DeepCopyInto(out *ScheduledBackupStatus) {
	*out = *in
	if in.StartTime != nil {
		in, out := &in.StartTime, &out.StartTime
		*out = (*in).DeepCopy()
	}
	if in.CompletionTime != nil {
		in, out := &in.CompletionTime, &out.CompletionTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScheduledBackupStatus.
func (in *ScheduledBackupStatus) DeepCopy() *ScheduledBackupStatus {
	if in == nil {
		return nil
	}
	out := new(ScheduledBackupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSpec) DeepCopyInto(out *ServiceSpec) {
	*out = *in
	if in.NodePort != nil {
		in, out := &in.NodePort, &out.NodePort
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSpec.
func (in *ServiceSpec) DeepCopy() *ServiceSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *XenonOpts) DeepCopyInto(out *XenonOpts) {
	*out = *in
	if in.AdmitDefeatHearbeatCount != nil {
		in, out := &in.AdmitDefeatHearbeatCount, &out.AdmitDefeatHearbeatCount
		*out = new(int32)
		**out = **in
	}
	if in.ElectionTimeout != nil {
		in, out := &in.ElectionTimeout, &out.ElectionTimeout
		*out = new(int32)
		**out = **in
	}
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new XenonOpts.
func (in *XenonOpts) DeepCopy() *XenonOpts {
	if in == nil {
		return nil
	}
	out := new(XenonOpts)
	in.DeepCopyInto(out)
	return out
}
