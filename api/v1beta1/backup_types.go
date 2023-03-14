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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BackupSpec defines the desired state of Backup
type BackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// ClusterName is the name of the cluster to be backed up.
	ClusterName string `json:"clusterName,omitempty"`
	// BackupMethod represents the type of backup
	// +kubebuilder:default:="xtrabackup"
	BackupMethod string `json:"mothod,omitempty"`
	// Defines details for manual pgBackRest backup Jobs
	// +optional
	Manual *ManualBackup `json:"manual,omitempty"`
	// Backup Schedule
	// +optional
	BackupSchedule BackupSchedule `json:"schedule,omitempty"`
	// Backup Storage
	BackupOpts BackupOps `json:"backupops,omitempty"`
}

type BackupOps struct {
	// BackupHost
	// +optional
	BackupHost string `json:"host,omitempty"`
	S3         *S3    `json:"s3,omitempty"`
	NFS        *NFS   `json:"nfs,omitempty"`
}

type S3 struct {
	// S3 Bucket
	// +optional
	BackupSecretName string `json:"secretName,omitempty"`
}

type NFS struct {
	// NFS Server Address
	// +optional
	NFSServerAddress string `json:"nfsServerAddress,omitempty"`
}

type ManualBackup struct {
	BackupType string `json:"type,omitempty"`
	// Backup Retention
	// +optional
	// +kubebuilder:default:=7
	BackupRetention *int32 `json:"backupRetention,omitempty"`
}

type BackupSchedule struct {
	// Cron expression for backup schedule
	// +optional
	CronExpression string `json:"cronExpression,omitempty"`
	// Backup Retention
	// +optional
	// +kubebuilder:default:=7
	BackupRetention *int32 `json:"backupRetention,omitempty"`
	BackupType      string `json:"type,omitempty"`
	// History Limit of job
	// +optional
	// +kubebuilder:default:=3
	BackupJobHistoryLimit *int32 `json:"jobhistoryLimit,omitempty"`
}

// type BackupStatus struct {
// 	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
// 	// Important: Run "make" to regenerate code after modifying this file
// 	// +kubebuilder:default:=false
// 	Completed bool `json:"completed"`
// 	// Get the backup path.
// 	BackupName string `json:"backupName,omitempty"`
// 	// Get the backup Date
// 	BackupDate string `json:"backupDate,omitempty"`
// 	// Get the backup Type
// 	BackupType string `json:"backupType,omitempty"`
// 	// Conditions represents the backup resource conditions list.
// 	Conditions []BackupCondition `json:"conditions,omitempty"`
// }

type BackupStatus struct {
	ManualBackup     *ManualBackupStatus     `json:"manual,omitempty"`
	ScheduledBackups []ScheduledBackupStatus `json:"scheduled,omitempty"`
}

type ManualBackupStatus struct {
	// Specifies whether or not the Job is finished executing (does not indicate success or
	// failure).
	// +kubebuilder:validation:Required
	Finished   bool   `json:"finished"`
	BackupName string `json:"backupName,omitempty"`
	// Get the backup Date
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// Get the backup Type
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	// Conditions represents the backup resource conditions list.
	// +optional
	Succeeded bool `json:"succeeded,omitempty"`
	// +optional
	Failed int32  `json:"failed,omitempty"`
	Reason string `json:"reason"`
	// Get the backup Type
	BackupType string `json:"backupType,omitempty"`
}

type ScheduledBackupStatus struct {
	// The name of the associated pgBackRest scheduled backup CronJob
	// +kubebuilder:validation:Required
	CronJobName string `json:"cronJobName,omitempty"`
	// Get the backup path.
	BackupName string `json:"backupName,omitempty"`
	// Specifies whether or not the Job is finished executing (does not indicate success or
	// failure).
	// +kubebuilder:validation:Required
	Finished bool `json:"finished"`
	// Get the backup Type
	BackupType string `json:"backupType,omitempty"`
	// Get the backup Date
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// Get the backup Type
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	// Conditions represents the backup resource conditions list.
	// +optional
	Succeeded bool `json:"succeeded,omitempty"`
	// +optional
	Failed int32  `json:"failed,omitempty"`
	Reason string `json:"reason"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="BackupName",type="string",JSONPath=".status.backupName",description="The Backup name"
// +kubebuilder:printcolumn:name="BackupDate",type="string",JSONPath=".status.backupDate",description="The Backup Date time"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".status.backupType",description="The Backup Type"
// +kubebuilder:printcolumn:name="Success",type="string",JSONPath=".status.conditions[?(@.type==\"Complete\")].status",description="Whether the backup Success?"
// Backup is the Schema for the backups API
type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupSpec   `json:"spec,omitempty"`
	Status BackupStatus `json:"status,omitempty"`
}

// BackupCondition defines condition struct for backup resource
// type BackupCondition struct {
// 	// type of cluster condition, values in (\"Ready\")
// 	Type BackupConditionType `json:"type"`
// 	// Status of the condition, one of (\"True\", \"False\", \"Unknown\")
// 	Status corev1.ConditionStatus `json:"status"`
// 	// LastTransitionTime
// 	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
// 	// Reason
// 	Reason string `json:"reason"`
// 	// Message
// 	Message string `json:"message"`
// }

// BackupConditionType defines condition types of a backup resources
// type BackupConditionType string

// const (
// 	// BackupComplete means the backup has finished his execution
// 	BackupComplete BackupConditionType = "Complete"
// 	// BackupFailed means backup has failed
// 	BackupFailed BackupConditionType = "Failed"
// 	BackupStart  BackupConditionType = "Started"
// )

//+kubebuilder:object:root=true

// BackupList contains a list of Backup
type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backup{}, &BackupList{})
}
