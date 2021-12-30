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

package syncer

import (
	// "reflect"
	"testing"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/internal"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestStatusSyncer_updateClusterStatus(t *testing.T) {
	zero := int32(0)
	two := int32(2)
	three := int32(3)
	five := int32(5)
	type fields struct {
		MysqlCluster     *mysqlcluster.MysqlCluster
		cli              client.Client
		SQLRunnerFactory internal.SQLRunnerFactory
		XenonExecutor    internal.XenonExecutor
	}
	tests := []struct {
		name   string
		fields fields
		want   apiv1alpha1.ClusterCondition
	}{
		// Expext Replicas: 2
		// Ready nodes: 0
		// Last replicas: 0
		// Cluster state:
		// Last condition:
		// Expect Result: Initializing
		{
			name: "test init, wait first pod ready",
			fields: fields{
				MysqlCluster: ExpectCluster(two, 0, 0, "", ""),
			},
			want: apiv1alpha1.ClusterCondition{
				Type: apiv1alpha1.ConditionInit,
			},
		},
		// Expext Replicas: 2
		// Ready nodes: 1
		// Last replicas: 0
		// Cluster state:
		// Last condition:
		// Expect Result: Initializing
		{
			name: "test init, half done",
			fields: fields{
				MysqlCluster: ExpectCluster(two, 1, 0, "", ""),
			},
			want: apiv1alpha1.ClusterCondition{
				Type: apiv1alpha1.ConditionInit,
			},
		},
		// Expext Replicas: 2
		// Ready nodes: 2
		// Last replicas: 1
		// Cluster state: Initializing
		// Last condition: Initializing
		// Expect Result: Initializing (ready can not be set in this function, only keep last state.)
		{
			name: "test init, done",
			fields: fields{
				MysqlCluster: ExpectCluster(two, 2, 1, apiv1alpha1.ClusterInitState, apiv1alpha1.ConditionInit),
			},
			want: apiv1alpha1.ClusterCondition{
				Type: apiv1alpha1.ConditionInit,
			},
		},
		// Expext Replicas: 0
		// Ready nodes: 0
		// Last replicas: 2
		// Cluster state: Ready
		// Last condition: Ready
		// Expect Result: Closed
		{
			name: "test closed",
			fields: fields{
				MysqlCluster: ExpectCluster(zero, 0, 2, apiv1alpha1.ClusterReadyState, apiv1alpha1.ConditionReady),
			},
			want: apiv1alpha1.ClusterCondition{
				Type: apiv1alpha1.ConditionClose,
			},
		},
		// Expext Replicas: 2
		// Ready nodes: 2
		// Last replicas: 3
		// Cluster state: Ready
		// Last condition: Ready
		// Expect Result: Scale in
		{
			name: "test scale in",
			fields: fields{
				MysqlCluster: ExpectCluster(two, 2, 3, apiv1alpha1.ClusterReadyState, apiv1alpha1.ConditionReady),
			},
			want: apiv1alpha1.ClusterCondition{
				Type: apiv1alpha1.ConditionScaleIn,
			},
		},
		// Expext Replicas: 3
		// Ready nodes: 2
		// Last replicas: 2
		// Cluster state: Ready
		// Last condition: Ready
		// Expect Result: Scale out
		{
			name: "test scale out",
			fields: fields{
				MysqlCluster: ExpectCluster(three, 2, 2, apiv1alpha1.ClusterReadyState, apiv1alpha1.ConditionReady),
			},
			want: apiv1alpha1.ClusterCondition{
				Type: apiv1alpha1.ConditionScaleOut,
			},
		},
		// Expext Replicas: 5
		// Ready nodes: 4
		// Last replicas: 3
		// Cluster state: Scale out
		// Last condition: Scale out
		// Expect Result: Scale out
		{
			name: "test scale out, half done",
			fields: fields{
				MysqlCluster: ExpectCluster(five, 4, 3, apiv1alpha1.ClusterScaleOutState, apiv1alpha1.ConditionScaleOut),
			},
			want: apiv1alpha1.ClusterCondition{
				Type: apiv1alpha1.ConditionScaleOut,
			},
		},
		// Expext Replicas: 3
		// Ready nodes: 2
		// Last replicas: 3
		// Cluster state: Ready
		// Last condition: Ready
		// Expect Result: Failover
		{
			name: "test failover",
			fields: fields{
				MysqlCluster: ExpectCluster(three, 2, 3, apiv1alpha1.ClusterReadyState, apiv1alpha1.ConditionReady),
			},
			want: apiv1alpha1.ClusterCondition{
				Type: apiv1alpha1.ConditionFailover,
			},
		},
		// Expext Replicas: 3
		// Ready nodes: 3
		// Last replicas: 2
		// Cluster state: Failover
		// Last condition: Failover
		// Expect Result: Failover
		{
			name: "test failover, recover",
			fields: fields{
				MysqlCluster: ExpectCluster(three, 3, 2, apiv1alpha1.ClusterFailoverState, apiv1alpha1.ConditionFailover),
			},
			want: apiv1alpha1.ClusterCondition{
				Type: apiv1alpha1.ConditionFailover,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StatusSyncer{
				MysqlCluster:     tt.fields.MysqlCluster,
				cli:              tt.fields.cli,
				SQLRunnerFactory: tt.fields.SQLRunnerFactory,
				XenonExecutor:    tt.fields.XenonExecutor,
			}
			got := s.updateClusterStatus()
			if got.Type != tt.want.Type {
				t.Errorf("want state: %s, actual state: %s", tt.want.Type, got.Type)
			}
		})
	}
}

func ExpectCluster(expect int32, readyNodes, lastReplicas int,
	clusterState apiv1alpha1.ClusterState, lastConditionType apiv1alpha1.ClusterConditionType) *mysqlcluster.MysqlCluster {
	return &mysqlcluster.MysqlCluster{
		MysqlCluster: &apiv1alpha1.MysqlCluster{
			Spec: apiv1alpha1.MysqlClusterSpec{
				Replicas: &expect,
			},
			Status: apiv1alpha1.MysqlClusterStatus{
				Conditions: []apiv1alpha1.ClusterCondition{
					{
						Type:     lastConditionType,
						Replicas: lastReplicas,
					},
				},
				ReadyNodes: readyNodes,
				State:      clusterState,
			},
		},
	}
}
