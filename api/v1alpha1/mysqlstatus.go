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
	"fmt"
	"time"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// ClusterErrorState indicates whether the cluster is in error state.
	ClusterErrorState ClusterState = "Error"
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

const (
	RaftNotReadyReason = "RaftNotReady"
)

// The max quantity of the statuses.
const maxStatusesQuantity = 10

func NewClusterCondition(condType ClusterConditionType, status corev1.ConditionStatus, reason, message string) ClusterCondition {
	return ClusterCondition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             reason,
		Message:            message,
	}
}

// error -> ready, close, error(different)
// close -> scaleout, update
// ready -> scale, update, close, error, ready(different)
// scale -> ready, close, error
// update -> ready, close, error
func (cs *MysqlClusterStatus) appendClusterCondition(condition ClusterCondition) {
	if len(cs.Conditions) == 0 {
		cs.Conditions = append(cs.Conditions, condition)
	} else {
		lastCond := cs.Conditions[len(cs.Conditions)-1]

		switch lastCond.Type {
		case ConditionError:
			if condition.Type == ConditionReady || condition.Type == ConditionClose {
				cs.Conditions = append(cs.Conditions, condition)
			}
			if condition.Type == ConditionError && lastCond.Reason != condition.Reason {
				cs.Conditions = append(cs.Conditions, condition)
			}
		case ConditionClose:
			if condition.Type == ConditionScaleOut || condition.Type == ConditionUpdate {
				cs.Conditions = append(cs.Conditions, condition)
			}
		case ConditionReady:
			if condition.Type == ConditionReady && condition.Message == lastCond.Message {
				break
			}
			cs.Conditions = append(cs.Conditions, condition)
		default:
			if condition.Type == ConditionReady || condition.Type == ConditionClose || condition.Type == ConditionError {
				cs.Conditions = append(cs.Conditions, condition)
			}
		}
	}
	// Truncate outdate cluster condition.
	if len(cs.Conditions) > maxStatusesQuantity {
		cs.Conditions = cs.Conditions[len(cs.Conditions)-maxStatusesQuantity:]
	}
}

func (cs *MysqlClusterStatus) AppendReadyCondition() {
	c := NewClusterCondition(ConditionReady, corev1.ConditionTrue, "", fmt.Sprintf("Ready nodes: %d", cs.ReadyNodes))
	cs.appendClusterCondition(c)
}

func (cs *MysqlClusterStatus) AppendUpdateCondition() {
	c := NewClusterCondition(ConditionUpdate, corev1.ConditionTrue, "", "")
	cs.appendClusterCondition(c)
}

func (cs *MysqlClusterStatus) AppendScaleInCondition(currentReplicas, specReplicas int32) {
	c := NewClusterCondition(ConditionScaleIn, corev1.ConditionTrue, "", fmt.Sprintf("Replicas: %d to %d", currentReplicas, specReplicas))
	cs.appendClusterCondition(c)
}

func (cs *MysqlClusterStatus) AppendScaleOutCondition(currentReplicas, specReplicas int32) {
	c := NewClusterCondition(ConditionScaleOut, corev1.ConditionTrue, "", fmt.Sprintf("Replicas: %d to %d", currentReplicas, specReplicas))
	cs.appendClusterCondition(c)
}

func (cs *MysqlClusterStatus) AppendClosedCondition() {
	c := NewClusterCondition(ConditionClose, corev1.ConditionTrue, "", "")
	cs.appendClusterCondition(c)
}

func (cs *MysqlClusterStatus) AppendErrorCondition(reason, message string) {
	c := NewClusterCondition(ConditionError, corev1.ConditionTrue, reason, message)
	cs.appendClusterCondition(c)
}

func (cs *MysqlClusterStatus) AppendInitCondition(specReplicas int32) {
	c := NewClusterCondition(ConditionInit, corev1.ConditionTrue, "", fmt.Sprintf("Replicas: 0 to %d", specReplicas))
	cs.appendClusterCondition(c)
}

// tolerateError return true if the error still exists after timeout.
func tolerateError(condition ClusterCondition, reason string) bool {
	switch condition.Reason {
	case RaftNotReadyReason:
		return errorTimeOut(condition.LastTransitionTime.Time, -10*time.Second)
	case corev1.PodReasonUnschedulable:
		return errorTimeOut(condition.LastTransitionTime.Time, -1*time.Minute)
	}
	return false
}

func errorTimeOut(LastTransitionTime time.Time, timeOut time.Duration) bool {
	return LastTransitionTime.Before(time.Now().Add(timeOut))
}

func (cs *MysqlClusterStatus) ClusterReady(podReady bool) bool {
	if cs.State == ClusterReadyState || cs.State == ClusterErrorState {
		if cs.RaftReady() {
			return true
		}
		cs.AppendErrorCondition(RaftNotReadyReason, "Raft is not ready")
		return false
	}
	return podReady && cs.RaftReady()
}

func (cs *MysqlClusterStatus) ClusterClosed(replicas int32) bool {
	return replicas == 0 && cs.ReadyNodes == 0
}

func (cs *MysqlClusterStatus) ClusterError() bool {
	if len(cs.Conditions) > 0 {
		condition := cs.Conditions[len(cs.Conditions)-1]
		if condition.Type == ConditionError {
			return tolerateError(condition, condition.Reason)
		}
	}
	return false
}

func (cs *MysqlClusterStatus) ClusterUpdating() bool {
	if len(cs.Conditions) > 0 {
		conditionType := cs.Conditions[len(cs.Conditions)-1].Type
		return conditionType == ConditionUpdate
	}
	return false
}

func (cs *MysqlClusterStatus) ClusterScaling() bool {
	if len(cs.Conditions) > 0 {
		conditionType := cs.Conditions[len(cs.Conditions)-1].Type
		return conditionType == ConditionScaleOut || conditionType == ConditionScaleIn
	}
	return false
}

func (cs *MysqlClusterStatus) ClusterInitializing() bool {
	return cs.State == ClusterInitState || cs.State == ""
}

// Only one leader.
// Must have one leader.
func (cs *MysqlClusterStatus) RaftReady() bool {
	leader := ""
	raftReadyNodes := 0

	for _, nodeStatus := range cs.Nodes {
		switch nodeStatus.RaftStatus.Role {
		case string(utils.Leader):
			if leader != "" {
				return false
			}
			leader = nodeStatus.Name
			raftReadyNodes++
		case string(utils.Follower):
			raftReadyNodes++
		default:
		}
	}
	return leader != "" && raftReadyNodes >= 2
}
