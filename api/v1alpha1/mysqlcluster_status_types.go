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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
)

// ClusterConditionType defines type for cluster condition type.
type ClusterConditionType string

const (
	ClusterInitialized ClusterConditionType = "Initialized"
	ClusterAvaliable   ClusterConditionType = "Avaliable"
	ClusterInUpgrade   ClusterConditionType = "InUpgrade"
	ClusterAllReady    ClusterConditionType = "AllReady"
)

// MySQLClusterCondition defines type for cluster conditions.
type MySQLClusterCondition struct {
	// Type of cluster condition, values in (\"Initializing\", \"Ready\", \"Error\").
	Type ClusterConditionType `json:"type"`
	// Status of the condition, one of (\"True\", \"False\", \"Unknown\").
	Status metav1.ConditionStatus `json:"status"`
}

// MysqlClusterStatus defines the observed state of MysqlCluster
type MysqlClusterStatus struct {
	// StatefulSetStatus is the status of the StatefulSet reconciled by the mysqlcluster controller.
	StatefulSetStatus *appsv1.StatefulSetStatus `json:"statefulSetStatus,omitempty"`
	// State is the state of the mysqlcluster.
	State ClusterState `json:"state,omitempty"`
	// Conditions contains the list of the mysqlcluster conditions.
	Conditions []MySQLClusterCondition `json:"conditions,omitempty"`
}

func (s *MysqlClusterStatus) GetStatefulSetStatus() *appsv1.StatefulSetStatus {
	return s.StatefulSetStatus
}

func (s *MysqlClusterStatus) GetState() ClusterState {
	return s.State
}

func (s *MysqlClusterStatus) GetConditions() []MySQLClusterCondition {
	return s.Conditions
}

func (s *MysqlClusterStatus) SetStatefulSetStatus(status *appsv1.StatefulSetStatus) {
	s.StatefulSetStatus = status
}

func (s *MysqlClusterStatus) SetState(state ClusterState) {
	s.State = state
}

func (s *MysqlClusterStatus) SetConditions(conditions []MySQLClusterCondition) {
	s.Conditions = conditions
}
