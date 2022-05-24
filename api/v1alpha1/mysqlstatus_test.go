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
	"testing"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

func TestMysqlClusterStatus_RaftReady(t *testing.T) {
	type fields struct {
		ReadyNodes int
		State      ClusterState
		Conditions []ClusterCondition
		Nodes      []NodeStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "three followers, no leader",
			fields: fields{
				Nodes: []NodeStatus{
					{
						Name: "node1",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
					{
						Name: "node2",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
					{
						Name: "node3",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "two leader, one follower",
			fields: fields{
				Nodes: []NodeStatus{
					{
						Name: "node1",
						RaftStatus: RaftStatus{
							Role: string(utils.Leader),
						},
					},
					{
						Name: "node2",
						RaftStatus: RaftStatus{
							Role: string(utils.Leader),
						},
					},
					{
						Name: "node3",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "one leader, two followers",
			fields: fields{
				Nodes: []NodeStatus{
					{
						Name: "node1",
						RaftStatus: RaftStatus{
							Role: string(utils.Leader),
						},
					},
					{
						Name: "node2",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
					{
						Name: "node3",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "one leader, two unknwon",
			fields: fields{
				Nodes: []NodeStatus{
					{
						Name: "node1",
						RaftStatus: RaftStatus{
							Role: string(utils.Unknown),
						},
					},
					{
						Name: "node2",
						RaftStatus: RaftStatus{
							Role: string(utils.Leader),
						},
					},
					{
						Name: "node3",
						RaftStatus: RaftStatus{
							Role: string(utils.Unknown),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "one leader, one follower, one unknown",
			fields: fields{
				Nodes: []NodeStatus{
					{
						Name: "node1",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
					{
						Name: "node2",
						RaftStatus: RaftStatus{
							Role: string(utils.Leader),
						},
					},
					{
						Name: "node3",
						RaftStatus: RaftStatus{
							Role: string(utils.Unknown),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "one leader, one follower",
			fields: fields{
				Nodes: []NodeStatus{
					{
						Name: "node1",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
					{
						Name: "node2",
						RaftStatus: RaftStatus{
							Role: string(utils.Leader),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "two follower",
			fields: fields{
				Nodes: []NodeStatus{
					{
						Name: "node1",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
					{
						Name: "node2",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "one follower, one candidate",
			fields: fields{
				Nodes: []NodeStatus{
					{
						Name: "node1",
						RaftStatus: RaftStatus{
							Role: string(utils.Candidate),
						},
					},
					{
						Name: "node2",
						RaftStatus: RaftStatus{
							Role: string(utils.Follower),
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &MysqlClusterStatus{
				ReadyNodes: tt.fields.ReadyNodes,
				State:      tt.fields.State,
				Conditions: tt.fields.Conditions,
				Nodes:      tt.fields.Nodes,
			}
			if got := cs.RaftReady(); got != tt.want {
				t.Errorf("MysqlClusterStatus.RaftReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
