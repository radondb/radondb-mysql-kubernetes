/*
Copyright 2021 zhyass.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MysqlBackupSpec defines the desired state of MysqlBackup
type MysqlBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MysqlBackup. Edit mysqlbackup_types.go to remove/update

	// HostName represents the host for which to take backup
	// +optional
	HostName string `json:"hostname,omitempty"`
	//clustName , as same as Release in helm
	ClusterName string `json:"clustname"`
	// Represents the name of backup to PVC
	// +optional
	BackupToNFS string `json:"BackupToNFS,omitempty"`
}

// MysqlBackupStatus defines the observed state of MysqlBackup
type MysqlBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Completed bool `json:"completed,omitempty"`

	// Conditions represents the backup resource conditions list.
	Conditions []BackupCondition `json:"conditions,omitempty"`
	//BackupDir  string            `json:"backupdir,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MysqlBackup is the Schema for the mysqlbackups API
type MysqlBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlBackupSpec   `json:"spec,omitempty"`
	Status MysqlBackupStatus `json:"status,omitempty"`
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
)

//+kubebuilder:object:root=true

// MysqlBackupList contains a list of MysqlBackup
type MysqlBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MysqlBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MysqlBackup{}, &MysqlBackupList{})
}
