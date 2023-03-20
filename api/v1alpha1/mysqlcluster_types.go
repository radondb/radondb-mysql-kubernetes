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

// MysqlClusterSpec defines the desired state of MysqlCluster
type MysqlClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas is the number of pods.
	// +optional
	// +kubebuilder:validation:Enum=0;1;2;3;5
	// +kubebuilder:default:=3
	Replicas *int32 `json:"replicas,omitempty"`
	// Readonlys Info.
	// +optional
	ReadOnlys *ReadOnlyType `json:"readonlys,omitempty"`
	// The number of pods from that set that must still be available after the
	// eviction, even in the absence of the evicted pod
	// +optional
	// +kubebuilder:default:="50%"
	MinAvailable string `json:"minAvailable,omitempty"`

	// MysqlOpts is the options of MySQL container.
	// +optional
	// +kubebuilder:default:={rootPassword: "", rootHost: "localhost", user: "radondb_usr", password: "RadonDB@123", database: "radondb", initTokuDB: false, resources: {limits: {cpu: "500m", memory: "1Gi"}, requests: {cpu: "100m", memory: "256Mi"}}}
	MysqlOpts MysqlOpts `json:"mysqlOpts,omitempty"`

	// XenonOpts is the options of xenon container.
	// +optional
	// +kubebuilder:default:={image: "radondb/xenon:v2.3.0", admitDefeatHearbeatCount: 5, electionTimeout: 10000, resources: {limits: {cpu: "100m", memory: "256Mi"}, requests: {cpu: "50m", memory: "128Mi"}}}
	XenonOpts XenonOpts `json:"xenonOpts,omitempty"`

	// MetricsOpts is the options of metrics container.
	// +optional
	// +kubebuilder:default:={image: "prom/mysqld-exporter:v0.12.1", resources: {limits: {cpu: "100m", memory: "128Mi"}, requests: {cpu: "10m", memory: "32Mi"}}, enabled: false}
	MetricsOpts MetricsOpts `json:"metricsOpts,omitempty"`

	// Represents the MySQL version that will be run. The available version can be found here:
	// This field should be set even if the Image is set to let the operator know which mysql version is running.
	// Based on this version the operator can take decisions which features can be used.
	// +optional
	// +kubebuilder:default:="5.7"
	MysqlVersion string `json:"mysqlVersion,omitempty"`

	// PodPolicy defines the policy to extra specification.
	// +optional
	// +kubebuilder:default:={imagePullPolicy: "IfNotPresent", extraResources: {requests: {cpu: "10m", memory: "32Mi"}}, sidecarImage: "radondb/mysql57-sidecar:v2.3.0", busyboxImage: "busybox:1.32"}
	PodPolicy PodPolicy `json:"podPolicy,omitempty"`

	// PVC extra specifiaction.
	// +optional
	// +kubebuilder:default:={enabled: true, accessModes: {"ReadWriteOnce"}, size: "10Gi"}
	Persistence Persistence `json:"persistence,omitempty"`

	// Represents the name of the secret that contains credentials to connect to
	// the storage provider to store backups.
	// +optional
	BackupSecretName string `json:"backupSecretName,omitempty"`

	// Represents the name of the cluster restore from backup path.
	// +optional
	RestoreFrom string `json:"restoreFrom,omitempty"`

	// Represents NFS ip address where cluster restore from.
	// +optional
	NFSServerAddress string `json:"nfsServerAddress,omitempty"`

	// Specify under crontab format interval to take backups
	// leave it empty to deactivate the backup process
	// Defaults to ""
	// +optional
	BackupSchedule string `json:"backupSchedule,omitempty"`

	// Specify that crontab job backup both on NFS and S3 storage.
	// +optional
	BothS3NFS *BothS3NFSOpt `json:"bothS3NFS,omitempty"`

	// If set keeps last BackupScheduleJobsHistoryLimit Backups
	// +optional
	// +kubebuilder:default:=6
	BackupScheduleJobsHistoryLimit *int `json:"backupScheduleJobsHistoryLimit,omitempty"`

	// Containing CA (ca.crt) and server cert (tls.crt), server private key (tls.key) for SSL
	// +optional
	TlsSecretName string `json:"tlsSecretName,omitempty"`
}

// ReadOnly define the ReadOnly pods
type ReadOnlyType struct {
	// ReadOnlys is the number of readonly pods.
	Num int32 `json:"num"`
	// When the host name is empty, use the leader to change master
	// +optional
	Host string `json:"hostname"`
	// The compute resource requirements.
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// MysqlOpts defines the options of MySQL container.
type MysqlOpts struct {
	// Specifies mysql image to use.
	// +optional
	// +kubebuilder:default:="percona/percona-server:5.7.34"
	Image string `json:"image,omitempty"`
	// Unchangeable: Use super users instead
	// Password for the root user, can be empty or 8~32 characters long.
	// Only be a combination of uppercase letters, lowercase letters, numbers or special characters.
	// Special characters are supported: @#$%^&*_+-=.
	// +optional
	// +kubebuilder:default:=""
	// +kubebuilder:validation:Enum=""
	RootPassword string `json:"rootPassword,omitempty"`

	// Unchangeable: Use super users instead.
	// The root user's host.
	// +optional
	// +kubebuilder:validation:Enum=localhost
	// +kubebuilder:default:="localhost"
	RootHost string `json:"rootHost,omitempty"`

	// Username of new user to create.
	// Only be a combination of letters, numbers or underlines. The length can not exceed 26 characters.
	// +optional
	// +kubebuilder:default:="radondb_usr"
	// +kubebuilder:validation:Pattern="^[A-Za-z0-9_]{2,26}$"
	User string `json:"user,omitempty"`

	// Password for the new user, must be 8~32 characters long.
	// Only be a combination of uppercase letters, lowercase letters, numbers or special characters.
	// Special characters are supported: @#$%^&*_+-=.
	// +optional
	// +kubebuilder:default:="RadonDB@123"
	// +kubebuilder:validation:Pattern="^[A-Za-z0-9@#$%^&*_+\\-=]{8,32}$"
	Password string `json:"password,omitempty"`

	// Name for new database to create.
	// +optional
	// +kubebuilder:default:="radondb"
	Database string `json:"database,omitempty"`

	// InitTokuDB represents if install tokudb engine.
	// +optional
	// +kubebuilder:default:=false
	InitTokuDB bool `json:"initTokuDB,omitempty"`

	// MysqlConfTemplate is the configmap name of the template for mysql config.
	// The configmap should contain the keys `mysql.cnf` and `plugin.cnf` at least, key `init.sql` is optional.
	// If empty, operator will generate a default template named <spec.metadata.name>-mysql.
	// +optional
	MysqlConfTemplate string `json:"mysqlConfTemplate,omitempty"`

	// A map[string]string that will be passed to my.cnf file.
	// The key/value pairs is persisted in the configmap.
	// Delete key is not valid, it is recommended to edit the configmap directly.
	// +optional
	MysqlConf MysqlConf `json:"mysqlConf,omitempty"`

	// A map[string]string that will be passed to plugin.cnf file.
	// The key/value pairs is persisted in the configmap.
	// Delete key is not valid, it is recommended to edit the configmap directly.
	// +optional
	PluginConf MysqlConf `json:"pluginConf,omitempty"`

	// The compute resource requirements.
	// +optional
	// +kubebuilder:default:={limits: {cpu: "500m", memory: "1Gi"}, requests: {cpu: "100m", memory: "256Mi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// XenonOpts defines the options of xenon container.
type XenonOpts struct {
	// To specify the image that will be used for xenon container.
	// +optional
	// +kubebuilder:default:="radondb/xenon:v2.3.0"
	Image string `json:"image,omitempty"`

	// High available component admit defeat heartbeat count.
	// +optional
	// +kubebuilder:default:=5
	AdmitDefeatHearbeatCount *int32 `json:"admitDefeatHearbeatCount,omitempty"`

	// High available component election timeout. The unit is millisecond.
	// +optional
	// +kubebuilder:default:=10000
	ElectionTimeout *int32 `json:"electionTimeout,omitempty"`

	// If true, when the data is inconsistent, Xenon will automatically rebuild the invalid node.
	// +optional
	// +kubebuilder:default:=false
	EnableAutoRebuild bool `json:"enableAutoRebuild,omitempty"`

	// The compute resource requirements.
	// +optional
	// +kubebuilder:default:={limits: {cpu: "100m", memory: "256Mi"}, requests: {cpu: "50m", memory: "128Mi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// MetricsOpts defines the options of metrics container.
type MetricsOpts struct {
	// To specify the image that will be used for metrics container.
	// +optional
	// +kubebuilder:default:="prom/mysqld-exporter:v0.12.1"
	Image string `json:"image,omitempty"`

	// The compute resource requirements.
	// +optional
	// +kubebuilder:default:={limits: {cpu: "100m", memory: "128Mi"}, requests: {cpu: "10m", memory: "32Mi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Enabled represents if start a metrics container.
	// +optional
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`
}

// MysqlConf defines type for extra cluster configs. It's a simple map between
// string and string.
type MysqlConf map[string]string

// PodPolicy defines the general configuration and extra resources of pod.
type PodPolicy struct {
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +kubebuilder:default:="IfNotPresent"
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	Labels            map[string]string   `json:"labels,omitempty"`
	Annotations       map[string]string   `json:"annotations,omitempty"`
	Affinity          *corev1.Affinity    `json:"affinity,omitempty"`
	PriorityClassName string              `json:"priorityClassName,omitempty"`
	Tolerations       []corev1.Toleration `json:"tolerations,omitempty"`
	SchedulerName     string              `json:"schedulerName,omitempty"`

	// ExtraResources defines quotas for containers other than mysql or xenon.
	// These containers take up less resources, so quotas are set uniformly.
	// +optional
	// +kubebuilder:default:={requests: {cpu: "10m", memory: "32Mi"}}
	ExtraResources corev1.ResourceRequirements `json:"extraResources,omitempty"`

	// To specify the image that will be used for sidecar container.
	// +optional
	// +kubebuilder:default:="radondb/mysql57-sidecar:v2.3.0"
	SidecarImage string `json:"sidecarImage,omitempty"`

	// The busybox image.
	// +optional
	// +kubebuilder:default:="busybox:1.32"
	BusyboxImage string `json:"busyboxImage,omitempty"`

	// SlowLogTail represents if tail the mysql slow log.
	// +optional
	// +kubebuilder:default:=false
	SlowLogTail bool `json:"slowLogTail,omitempty"`

	// AuditLogTail represents if tail the mysql audit log.
	// +optional
	// +kubebuilder:default:=false
	AuditLogTail bool `json:"auditLogTail,omitempty"`
}

// Persistence is the desired spec for storing mysql data. Only one of its
// members may be specified.
type Persistence struct {
	// Create a volume to store data.
	// +optional
	// +kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	// AccessModes contains the desired access modes the volume should have.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1
	// +optional
	// +kubebuilder:default:={"ReadWriteOnce"}
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// Name of the StorageClass required by the claim.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// Size of persistent volume claim.
	// +optional
	// +kubebuilder:default:="10Gi"
	Size string `json:"size,omitempty"`
}

// bothS3NFS opt
type BothS3NFSOpt struct {
	// NFS schedule.
	NFSSchedule string `json:"nfsSchedule,omitempty"`
	S3Schedule  string `json:"s3Schedule,omitempty"`
}

// ClusterState defines cluster state.
type ClusterState string

const (
	// ClusterInitState indicates whether the cluster is initializing.
	ClusterInitState ClusterState = "Initializing"
	// ClusterUpdateState indicates whether the cluster is being updated.
	ClusterUpdateState ClusterState = "Updating"
	// ClusterReadyState indicates whether all containers in the pod are ready.
	ClusterReadyState ClusterState = "Ready"
	// ClusterCloseState indicates whether the cluster is closed.
	ClusterCloseState ClusterState = "Closed"
	// ClusterScaleInState indicates whether the cluster replicas is decreasing.
	ClusterScaleInState ClusterState = "ScaleIn"
	// ClusterScaleOutState indicates whether the cluster replicas is increasing.
	ClusterScaleOutState ClusterState = "ScaleOut"
)

// ClusterConditionType defines type for cluster condition type.
type ClusterConditionType string

const (
	// ConditionInit indicates whether the cluster is initializing.
	ConditionInit ClusterConditionType = "Initializing"
	// ConditionUpdate indicates whether the cluster is being updated.
	ConditionUpdate ClusterConditionType = "Updating"
	// ConditionReady indicates whether all containers in the pod are ready.
	ConditionReady ClusterConditionType = "Ready"
	// ConditionClose indicates whether the cluster is closed.
	ConditionClose ClusterConditionType = "Closed"
	// ConditionError indicates whether there is an error in the cluster.
	ConditionError ClusterConditionType = "Error"
	// ConditionScaleIn indicates whether the cluster replicas is decreasing.
	ConditionScaleIn ClusterConditionType = "ScaleIn"
	// ConditionScaleOut indicates whether the cluster replicas is increasing.
	ConditionScaleOut ClusterConditionType = "ScaleOut"
)

// ClusterCondition defines type for cluster conditions.
type ClusterCondition struct {
	// Type of cluster condition, values in (\"Initializing\", \"Ready\", \"Error\").
	Type ClusterConditionType `json:"type"`
	// Status of the condition, one of (\"True\", \"False\", \"Unknown\").
	Status corev1.ConditionStatus `json:"status"`

	// The last time this Condition type changed.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	// One word, camel-case reason for current status of the condition.
	Reason string `json:"reason,omitempty"`
	// Full text reason for current status of the condition.
	Message string `json:"message,omitempty"`
}

// NodeStatus defines type for status of a node into cluster.
type NodeStatus struct {
	// Name of the node.
	Name string `json:"name"`
	// Full text reason for current status of the node.
	Message string `json:"message,omitempty"`
	// RaftStatus is the raft status of the node.
	RaftStatus RaftStatus `json:"raftStatus,omitempty"`
	// Conditions contains the list of the node conditions fulfilled.
	Conditions []NodeCondition `json:"conditions,omitempty"`
}

type RaftStatus struct {
	// Role is one of (LEADER/CANDIDATE/FOLLOWER/IDLE/INVALID)
	Role string `json:"role,omitempty"`
	// Leader is the name of the Leader of the current node.
	Leader string `json:"leader,omitempty"`
	// Nodes is a list of nodes that can be identified by the current node.
	Nodes []string `json:"nodes,omitempty"`
}

// NodeCondition defines type for representing node conditions.
type NodeCondition struct {
	// Type of the node condition.
	Type NodeConditionType `json:"type"`
	// Status of the node, one of (\"True\", \"False\", \"Unknown\").
	Status corev1.ConditionStatus `json:"status"`
	// The last time this Condition type changed.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}

// The index of the NodeStatus.Conditions.
type NodeConditionsIndex uint8

const (
	IndexLagged NodeConditionsIndex = iota
	IndexLeader
	IndexReadOnly
	IndexReplicating
)

// NodeConditionType defines type for node condition type.
type NodeConditionType string

const (
	// NodeConditionLagged represents if the node is lagged.
	NodeConditionLagged NodeConditionType = "Lagged"
	// NodeConditionLeader represents if the node is leader or not.
	NodeConditionLeader NodeConditionType = "Leader"
	// NodeConditionReadOnly repesents if the node is read only or not
	NodeConditionReadOnly NodeConditionType = "ReadOnly"
	// NodeConditionReplicating represents if the node is replicating or not.
	NodeConditionReplicating NodeConditionType = "Replicating"
)

// MysqlClusterStatus defines the observed state of MysqlCluster
type MysqlClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ReadyNodes represents number of the nodes that are in ready state.
	ReadyNodes int `json:"readyNodes,omitempty"`
	// State
	State ClusterState `json:"state,omitempty"`
	// Conditions contains the list of the cluster conditions fulfilled.
	Conditions []ClusterCondition `json:"conditions,omitempty"`
	// Nodes contains the list of the node status fulfilled.
	Nodes []NodeStatus `json:"nodes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.readyNodes
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="The cluster status"
// +kubebuilder:printcolumn:name="Desired",type="integer",JSONPath=".spec.replicas",description="The number of desired replicas"
// +kubebuilder:printcolumn:name="Current",type="integer",JSONPath=".status.readyNodes",description="The number of current replicas"
// +kubebuilder:printcolumn:name="Leader",type="string",JSONPath=".status.nodes[?(@.raftStatus.role == 'LEADER')].name",description="Name of the leader node"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=mysql
// MysqlCluster is the Schema for the mysqlclusters API
type MysqlCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlClusterSpec   `json:"spec,omitempty"`
	Status MysqlClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// MysqlClusterList contains a list of MysqlCluster
type MysqlClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MysqlCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MysqlCluster{}, &MysqlClusterList{})
}
