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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UserSpec defines the desired state of User
type UserSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Username is the name of user to be operated.
	// This field should be immutable.
	// +optional
	// +kubebuilder:validation:Pattern="^[A-Za-z0-9_]{2,26}$"
	User string `json:"user,omitempty"`

	// Hosts is the grants hosts.
	// +optional
	Hosts []string `json:"hosts,omitempty"`

	// EnableSsl represents whether to open the ssl connection.
	// +optional
	// +kubebuilder:default:=false
	// TODO: add logic.
	EnableSsl bool `json:"enableSsl,omitempty"`

	// ClusterBinder Contains parameters about the cluster bound by user.
	// +optional
	ClusterBinder ClusterBinder `json:"clusterBinder,omitempty"`

	// SecretBinder Contains parameters about the secret object bound by user.
	// +optional
	SecretBinder SecretBinder `json:"secretBinder,omitempty"`

	// Permissions is the list of roles that user has in the specified database.
	// +optional
	Permissions []UserPermission `json:"permissions,omitempty"`
}

type ClusterBinder struct {
	// ClusterName is the name of cluster.
	ClusterName string `json:"clusterName,omitempty"`

	// NameSpace is the nameSpace of cluster.
	NameSpace string `json:"nameSpace,omitempty"`
}

type SecretBinder struct {
	// SecretName is the name of secret object.
	SecretName string `json:"secretName,omitempty"`

	// SecretKey is the key of secret object.
	SecretKey string `json:"secretKey,omitempty"`
}

// UserPermission defines a UserPermission permission.
type UserPermission struct {
	// Database is the grants database.
	// +optional
	Database string `json:"database,omitempty"`
	// Tables is the grants tables inside the database.
	// +optional
	Tables []string `json:"tables,omitempty"`
	// Privileges is the normal privileges(comma delimited, such as "SELECT,CREATE").
	// +optional
	// TODO: add validation.
	Privileges []string `json:"privileges,omitempty"`
}

// UserStatus defines the observed state of MysqlUser.
type UserStatus struct {

	// Conditions represents the MysqlUser resource conditions list.
	// +optional
	Conditions []UserCondition `json:"conditions,omitempty"`

	// AllowedHosts contains the list of hosts that the user is allowed to connect from.
	AllowedHosts []string `json:"allowedHosts,omitempty"`
}

// UserConditionType defines the condition types of a MysqlUser resource.
type UserConditionType string

const (
	// MySQLUserReady means the MySQL user is ready when database exists.
	MySQLUserReady UserConditionType = "Ready"
)

// UserCondition defines the condition struct for a MysqlUser resource.
type UserCondition struct {
	// Type of MysqlUser condition.
	Type UserConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// The reason for the condition's last transition.
	Reason string `json:"reason"`
	// A human readable message indicating details about the transition.
	Message string `json:"message"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:finalizers
// MysqlUser is the Schema for the users API.
type MysqlUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// MysqlUserList contains a list of MysqlUser.
type MysqlUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MysqlUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MysqlUser{}, &MysqlUserList{})
}
