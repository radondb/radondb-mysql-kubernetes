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
	"reflect"
	"testing"
	"time"

	. "bou.ke/monkey"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// todo: test truncate outdate cluster condition.
func TestMysqlClusterStatus_appendClusterCondition(t *testing.T) {
	type fields struct {
		ReadyNodes int
		State      ClusterState
		Conditions []ClusterCondition
		Nodes      []NodeStatus
	}
	type testCase struct {
		condition ClusterCondition
		success   bool
	}
	type nextCondition struct {
		canInit,
		canReady,
		canScaleOut,
		canScaleIn,
		canUpdate,
		canClose,
		canError bool
	}
	lastCondition := func(conditions []ClusterCondition) ClusterCondition {
		if len(conditions) == 0 {
			return ClusterCondition{}
		}
		return conditions[len(conditions)-1]
	}
	buildArgs := func(n nextCondition) []testCase {
		return []testCase{
			{
				condition: ClusterCondition{
					Type: ConditionInit,
				},
				success: n.canInit,
			},
			{
				condition: ClusterCondition{
					Type:    ConditionReady,
					Message: "Test message",
				},
				success: n.canReady,
			},
			{
				condition: ClusterCondition{
					Type: ConditionClose,
				},
				success: n.canClose,
			},
			{
				condition: ClusterCondition{
					Type: ConditionScaleIn,
				},
				success: n.canScaleIn,
			},
			{
				condition: ClusterCondition{
					Type: ConditionScaleOut,
				},
				success: n.canScaleOut,
			},
			{
				condition: ClusterCondition{
					Type: ConditionUpdate,
				},
				success: n.canUpdate,
			},
			{
				condition: ClusterCondition{
					Type:   ConditionError,
					Reason: "TestReason",
				},
				success: n.canError,
			},
		}
	}
	tests := []struct {
		name   string
		fields fields
		args   []testCase
	}{
		{
			name: "empty conditions",
			fields: fields{
				ReadyNodes: 0,
				State:      "",
				Conditions: []ClusterCondition{},
				Nodes:      []NodeStatus{},
			},
			args: buildArgs(nextCondition{
				canInit:     true,
				canReady:    true,
				canScaleOut: true,
				canScaleIn:  true,
				canUpdate:   true,
				canClose:    true,
				canError:    true,
			}),
		},
		{
			name: "lastest ready conditions",
			fields: fields{
				ReadyNodes: 2,
				State:      "",
				Conditions: []ClusterCondition{
					{
						Type:    ConditionReady,
						Status:  corev1.ConditionTrue,
						Message: "Ready nodes: 2",
						LastTransitionTime: metav1.Time{
							Time: time.Now(),
						},
					},
				},
				Nodes: []NodeStatus{},
			},
			args: append(buildArgs(nextCondition{
				canInit:     true,
				canReady:    true,
				canScaleOut: true,
				canScaleIn:  true,
				canUpdate:   true,
				canClose:    true,
				canError:    true,
			}),
				testCase{
					condition: ClusterCondition{
						Type:    ConditionReady,
						Status:  corev1.ConditionTrue,
						Message: "Ready nodes: 2",
						LastTransitionTime: metav1.Time{
							Time: time.Now(),
						},
					},
					success: false,
				},
			),
		},
		{
			name: "lastest error conditions",
			fields: fields{
				ReadyNodes: 2,
				State:      "",
				Conditions: []ClusterCondition{
					{
						Type:    ConditionError,
						Status:  corev1.ConditionTrue,
						Reason:  RaftNotReadyReason,
						Message: "Raft is not ready",
						LastTransitionTime: metav1.Time{
							Time: time.Now(),
						},
					},
				},
			},
			args: append(buildArgs(nextCondition{
				canInit:     false,
				canReady:    true,
				canScaleOut: false,
				canScaleIn:  false,
				canUpdate:   false,
				canClose:    true,
				canError:    true,
			}), testCase{
				condition: ClusterCondition{
					Type:    ConditionError,
					Status:  corev1.ConditionTrue,
					Reason:  RaftNotReadyReason,
					Message: "Raft is not ready",
					LastTransitionTime: metav1.Time{
						Time: time.Now(),
					},
				},
				success: false,
			},
				testCase{
					condition: ClusterCondition{
						Type:    ConditionError,
						Status:  corev1.ConditionTrue,
						Reason:  corev1.PodReasonUnschedulable,
						Message: "Pod Unschedulable",
						LastTransitionTime: metav1.Time{
							Time: time.Now(),
						},
					},
					// Only can append different error.
					success: true,
				}),
		},
		{
			name: "lastest closed conditions",
			fields: fields{
				ReadyNodes: 2,
				State:      "",
				Conditions: []ClusterCondition{
					{
						Type:   ConditionClose,
						Status: corev1.ConditionTrue,
						LastTransitionTime: metav1.Time{
							Time: time.Now(),
						},
					},
				},
				Nodes: []NodeStatus{},
			},
			args: buildArgs(nextCondition{
				canInit:     false,
				canReady:    false,
				canScaleOut: true,
				canScaleIn:  false,
				canUpdate:   true,
				canClose:    false,
				canError:    false,
			}),
		},
		{
			name: "lastest scaleout conditions",
			fields: fields{
				ReadyNodes: 2,
				State:      "",
				Conditions: []ClusterCondition{
					{
						Type:   ConditionScaleOut,
						Status: corev1.ConditionTrue,
						LastTransitionTime: metav1.Time{
							Time: time.Now(),
						},
					},
				},
				Nodes: []NodeStatus{},
			},
			args: buildArgs(nextCondition{
				canInit:     false,
				canReady:    true,
				canScaleOut: false,
				canScaleIn:  false,
				canUpdate:   false,
				canClose:    true,
				canError:    true,
			}),
		},
		{
			name: "lastest scalein conditions",
			fields: fields{
				ReadyNodes: 2,
				State:      "",
				Conditions: []ClusterCondition{
					{
						Type:   ConditionScaleIn,
						Status: corev1.ConditionTrue,
						LastTransitionTime: metav1.Time{
							Time: time.Now(),
						},
					},
				},
				Nodes: []NodeStatus{},
			},
			args: buildArgs(nextCondition{
				canInit:     false,
				canReady:    true,
				canScaleOut: false,
				canScaleIn:  false,
				canUpdate:   false,
				canClose:    true,
				canError:    true,
			}),
		},
		{
			name: "lastest update conditions",
			fields: fields{
				ReadyNodes: 2,
				State:      "",
				Conditions: []ClusterCondition{
					{
						Type:   ConditionUpdate,
						Status: corev1.ConditionTrue,
						LastTransitionTime: metav1.Time{
							Time: time.Now(),
						},
					},
				},
				Nodes: []NodeStatus{},
			},
			args: buildArgs(nextCondition{
				canInit:     false,
				canReady:    true,
				canScaleOut: false,
				canScaleIn:  false,
				canUpdate:   false,
				canClose:    true,
				canError:    true,
			}),
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
			for _, arg := range tt.args {
				cs.appendClusterCondition(arg.condition)
				lastCond := lastCondition(cs.Conditions)
				if arg.success {
					assert.Equal(t, lastCond, arg.condition)
				} else {
					assert.NotEqual(t, lastCond, arg.condition)
				}
				cs.Conditions = tt.fields.Conditions
			}
		})
	}
}

func TestMysqlClusterStatus_ClusterReady(t *testing.T) {
	type fields struct {
		ReadyNodes int
		State      ClusterState
		Conditions []ClusterCondition
		Nodes      []NodeStatus
	}
	type args struct {
		podReady  bool
		raftReady bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "state ready, pod ready, raft ready",
			fields: fields{
				ReadyNodes: 2,
				State:      ClusterReadyState,
			},
			args: args{
				podReady:  true,
				raftReady: true,
			},
			want: true,
		},
		{
			name: "state ready, pod not ready, raft ready",
			fields: fields{
				ReadyNodes: 2,
				State:      ClusterReadyState,
			},
			args: args{
				podReady:  false,
				raftReady: true,
			},
			want: true,
		},
		{
			name: "state ready, pod not ready, raft not ready",
			fields: fields{
				ReadyNodes: 2,
				State:      ClusterReadyState,
			},
			args: args{
				podReady:  false,
				raftReady: false,
			},
			want: false,
		},
		{
			name: "state ready, pod ready, raft not ready",
			fields: fields{
				ReadyNodes: 2,
				State:      ClusterReadyState,
			},
			args: args{
				podReady:  true,
				raftReady: false,
			},
			want: false,
		},
		{
			name: "state scale, pod not ready, raft ready",
			fields: fields{
				ReadyNodes: 2,
				State:      ClusterScaleOutState,
			},
			args: args{
				podReady:  false,
				raftReady: true,
			},
			want: false,
		},
		{
			name: "state scale, pod ready, raft ready",
			fields: fields{
				ReadyNodes: 2,
				State:      ClusterScaleOutState,
			},
			args: args{
				podReady:  true,
				raftReady: true,
			},
			want: true,
		},
		{
			name: "state update, pod not ready, raft ready",
			fields: fields{
				ReadyNodes: 2,
				State:      ClusterUpdateState,
			},
			args: args{
				podReady:  false,
				raftReady: true,
			},
			want: false,
		},
		{
			name: "state update, pod not ready, raft ready",
			fields: fields{
				ReadyNodes: 2,
				State:      ClusterUpdateState,
			},
			args: args{
				podReady:  true,
				raftReady: true,
			},
			want: true,
		},
		{
			name: "state error, pod not ready, raft ready",
			fields: fields{
				ReadyNodes: 2,
				State:      ClusterErrorState,
			},
			args: args{
				podReady:  false,
				raftReady: true,
			},
			want: true,
		},
		{
			name: "state closed",
			fields: fields{
				ReadyNodes: 2,
				State:      ClusterCloseState,
			},
			args: args{
				podReady:  true,
				raftReady: false,
			},
			want: false,
		},
	}

	raftReadyTest := func(cs *MysqlClusterStatus, podReady bool, want bool) {
		guard := PatchInstanceMethod(reflect.TypeOf(cs), "RaftReady", func(_ *MysqlClusterStatus) bool {
			return true
		})
		defer guard.Unpatch()
		assert.Equal(t, want, cs.ClusterReady(podReady))
	}

	raftNotReadyTest := func(cs *MysqlClusterStatus, podReady bool, want bool) {
		guard := PatchInstanceMethod(reflect.TypeOf(cs), "RaftReady", func(_ *MysqlClusterStatus) bool {
			return false
		})
		defer guard.Unpatch()
		assert.Equal(t, want, cs.ClusterReady(podReady))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &MysqlClusterStatus{
				ReadyNodes: tt.fields.ReadyNodes,
				State:      tt.fields.State,
				Conditions: tt.fields.Conditions,
				Nodes:      tt.fields.Nodes,
			}
			if tt.args.raftReady {
				raftReadyTest(cs, tt.args.podReady, tt.want)
			} else {
				raftNotReadyTest(cs, tt.args.podReady, tt.want)
			}
		})
	}
}
