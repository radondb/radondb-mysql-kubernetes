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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This is the backup Job CRD.
// BackupSpec defines the desired state of Backup
type BackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// To specify the image that will be used for sidecar container.
	// +optional
	// +kubebuilder:default:="radondb/mysql57-sidecar:v2.2.1"
	Image string `json:"image"`

	// HostName represents the host for which to take backup
	// If is empty, is use leader HostName
	HostName string `json:"hostName,omitempty"`

	// Represents the ip address of the nfs server.
	// +optional
	NFSServerAddress string `json:"nfsServerAddress,omitempty"`

	// ClusterName represents the cluster name to backup
	ClusterName string `json:"clusterName"`

	// History Limit of job
	// +optional
	// +kubebuilder:default:=3
	HistoryLimit *int32 `json:"historyLimit,omitempty"`
}

// BackupStatus defines the observed state of Backup
type BackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Completed indicates whether the backup is in a final state,
	// no matter whether its' corresponding job failed or succeeded
	// +kubebuilder:default:=false
	Completed bool `json:"completed"`
	// Get the backup path.
	BackupName string `json:"backupName,omitempty"`
	// Get the backup Date
	BackupDate string `json:"backupDate,omitempty"`
	// Get the backup Type
	BackupType string `json:"backupType,omitempty"`
	// Conditions represents the backup resource conditions list.
	Conditions []BackupCondition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
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
type BackupCondition struct {
	// type of cluster condition, values in (\"Ready\")
	Type BackupConditionType `json:"type"`
	// Status of the condition, one of (\"True\", \"False\", \"Unknown\")
	Status corev1.ConditionStatus `json:"status"`
	// LastTransitionTime
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// Reason
	Reason string `json:"reason"`
	// Message
	Message string `json:"message"`
}

// BackupConditionType defines condition types of a backup resources
type BackupConditionType string

const (
	// BackupComplete means the backup has finished his execution
	BackupComplete BackupConditionType = "Complete"
	// BackupFailed means backup has failed
	BackupFailed BackupConditionType = "Failed"
	BackupStart  BackupConditionType = "Started"
)

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
