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

	"github.com/looplab/fsm"
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
	// Type of cluster condition.
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
	ClusterInitializing ClusterConditionType = "Initializing"
	ClusterScaling      ClusterConditionType = "Scaling"
	ClusterUpdating     ClusterConditionType = "Updating"
	ClusterError        ClusterConditionType = "Error"
	ClusterAvailable    ClusterConditionType = "Available"
	NodesReady          ClusterConditionType = "NodesReady"
)

const (	
	// Reasons for cluster conditions. 
	ReasonClose    = "Closed"
	ReasonScaleIn  = "ScaleIn"
	ReasonScaleOut = "ScaleOut"
	// PodReasonUnschedulable reason in PodScheduled PodCondition means that the scheduler
	// can't schedule the pod right now, for example due to insufficient resources in the cluster.
	PodReasonUnschedulable   = "Unschedulable"
	ReasonClusterUnavailable = "ClusterUnavailable"

	ToleratePodUnschedulableTime   = 5 * time.Minute
	TolerateClusterUnavailableTime = 5 * time.Second
)

func newClusterCondition(condType ClusterConditionType, status corev1.ConditionStatus, reason, message string) ClusterCondition {
	return ClusterCondition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             reason,
		Message:            message,
	}
}

// If condition exist, replace it.
// If not exist, append a new condition.
func (cs *MysqlClusterStatus) UpdateClusterConditions(condType ClusterConditionType, cond ClusterCondition) {
	conditionIndex := -1
	for index, condition := range cs.Conditions {
		if condition.Type == condType {
			conditionIndex = index
			break
		}
	}
	if conditionIndex >= 0 {
		if cs.Conditions[conditionIndex].Status != cond.Status ||
			cs.Conditions[conditionIndex].Reason != cond.Reason ||
			cs.Conditions[conditionIndex].Message != cond.Message {
			cs.Conditions[conditionIndex] = cond
		}
	} else {
		cs.Conditions = append(cs.Conditions, cond)
	}
}

func (cs *MysqlClusterStatus) GenerateAvailableCondition(isAvailable corev1.ConditionStatus) ClusterCondition {
	return newClusterCondition(ClusterAvailable, isAvailable, "", "")
}

func (cs *MysqlClusterStatus) GenerateNodesReadyCondition(isNodesReady corev1.ConditionStatus) ClusterCondition {
	return newClusterCondition(NodesReady, isNodesReady, "", "")
}

func (cs *MysqlClusterStatus) GenerateUpdateCondition() ClusterCondition {
	return newClusterCondition(ClusterUpdating, corev1.ConditionTrue, "", "")
}

func (cs *MysqlClusterStatus) GenerateScaleCondition(currentReplicas, specReplicas int32) ClusterCondition {
	reason := ""
	if specReplicas == 0 {
		reason = ReasonClose
	} else {
		if currentReplicas > specReplicas {
			reason = ReasonScaleIn
		}
		if currentReplicas < specReplicas {
			reason = ReasonScaleOut
		}
	}

	return newClusterCondition(ClusterScaling, corev1.ConditionTrue, reason, fmt.Sprintf("Replicas: %d to %d", currentReplicas, specReplicas))
}

func (cs *MysqlClusterStatus) GenerateErrorCondition(reason, message string) ClusterCondition {
	return newClusterCondition(ClusterError, corev1.ConditionTrue, reason, message)
}

func (cs *MysqlClusterStatus) GenerateInitCondition() ClusterCondition {
	return newClusterCondition(ClusterInitializing, corev1.ConditionTrue, "", "")
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
