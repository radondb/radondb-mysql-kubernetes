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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/presslabs/controller-util/syncer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/internal"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// The max quantity of the statuses.
const maxStatusesQuantity = 10

// The retry time for check node status.
const checkNodeStatusRetry = 3

// StatusSyncer used to update the status.
type StatusSyncer struct {
	*mysqlcluster.MysqlCluster

	cli client.Client

	// Mysql query runner.
	internal.SQLRunnerFactory
}

// NewStatusSyncer returns a pointer to StatusSyncer.
func NewStatusSyncer(c *mysqlcluster.MysqlCluster, cli client.Client, sqlRunnerFactory internal.SQLRunnerFactory) *StatusSyncer {
	return &StatusSyncer{
		MysqlCluster:     c,
		cli:              cli,
		SQLRunnerFactory: sqlRunnerFactory,
	}
}

// Object returns the object for which sync applies.
func (s *StatusSyncer) Object() interface{} { return nil }

// GetObject returns the object for which sync applies
// Deprecated: use github.com/presslabs/controller-util/syncer.Object() instead.
func (s *StatusSyncer) GetObject() interface{} { return nil }

// Owner returns the object owner or nil if object does not have one.
func (s *StatusSyncer) ObjectOwner() runtime.Object { return s.MysqlCluster }

// GetOwner returns the object owner or nil if object does not have one.
// Deprecated: use github.com/presslabs/controller-util/syncer.ObjectOwner() instead.
func (s *StatusSyncer) GetOwner() runtime.Object { return s.MysqlCluster }

// Sync persists data into the external store.
func (s *StatusSyncer) Sync(ctx context.Context) (syncer.SyncResult, error) {
	clusterCondition := s.updateClusterStatus()

	list := corev1.PodList{}
	err := s.cli.List(
		ctx,
		&list,
		&client.ListOptions{
			Namespace:     s.Namespace,
			LabelSelector: s.GetLabels().AsSelector(),
		},
	)
	if err != nil {
		return syncer.SyncResult{}, err
	}

	// get ready nodes.
	var readyNodes []corev1.Pod
	for _, pod := range list.Items {
		for _, cond := range pod.Status.Conditions {
			switch cond.Type {
			case corev1.ContainersReady:
				if cond.Status == corev1.ConditionTrue {
					readyNodes = append(readyNodes, pod)
				}
			case corev1.PodScheduled:
				if cond.Reason == corev1.PodReasonUnschedulable {
					// When an error occurs, it is first recorded in the condition,
					// but the cluster status is not updated immediately.
					clusterCondition = apiv1alpha1.ClusterCondition{
						Type:               apiv1alpha1.ConditionError,
						Status:             corev1.ConditionTrue,
						LastTransitionTime: metav1.NewTime(time.Now()),
						Reason:             corev1.PodReasonUnschedulable,
						Message:            cond.Message,
					}
				}
			}
		}
	}

	s.Status.ReadyNodes = len(readyNodes)
	if s.Status.ReadyNodes == int(*s.Spec.Replicas) && int(*s.Spec.Replicas) != 0 {
		s.Status.State = apiv1alpha1.ClusterReadyState
		clusterCondition.Type = apiv1alpha1.ConditionReady
	}

	if len(s.Status.Conditions) == 0 {
		s.Status.Conditions = append(s.Status.Conditions, clusterCondition)
	} else {
		lastCond := s.Status.Conditions[len(s.Status.Conditions)-1]
		if lastCond.Type != clusterCondition.Type {
			s.Status.Conditions = append(s.Status.Conditions, clusterCondition)
		}
	}
	if len(s.Status.Conditions) > maxStatusesQuantity {
		s.Status.Conditions = s.Status.Conditions[len(s.Status.Conditions)-maxStatusesQuantity:]
	}

	// Update ready nodes' status.
	return syncer.SyncResult{}, s.updateNodeStatus(ctx, s.cli, readyNodes)
}

// updateClusterStatus update the cluster status and returns condition.
func (s *StatusSyncer) updateClusterStatus() apiv1alpha1.ClusterCondition {
	clusterCondition := apiv1alpha1.ClusterCondition{
		Type:               apiv1alpha1.ConditionInit,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	oldState := s.Status.State
	// If the state does not exist, the cluster is being initialized.
	if oldState == "" {
		s.Status.State = apiv1alpha1.ClusterInitState
		return clusterCondition
	}
	// If the expected number of replicas and the actual number
	// of replicas are both 0, the cluster has been closed.
	if int(*s.Spec.Replicas) == 0 && s.Status.ReadyNodes == 0 {
		clusterCondition.Type = apiv1alpha1.ConditionClose
		s.Status.State = apiv1alpha1.ClusterCloseState
		return clusterCondition
	}
	// When the cluster is ready or closed, the number of replicas changes,
	// indicating that the cluster is updating nodes.
	if oldState == apiv1alpha1.ClusterReadyState || oldState == apiv1alpha1.ClusterCloseState {
		if int(*s.Spec.Replicas) != s.Status.ReadyNodes {
			clusterCondition.Type = apiv1alpha1.ConditionUpdate
			s.Status.State = apiv1alpha1.ClusterUpdateState
			return clusterCondition
		}
	}

	clusterCondition.Type = apiv1alpha1.ClusterConditionType(oldState)
	return clusterCondition
}

// updateNodeStatus update the node status.
func (s *StatusSyncer) updateNodeStatus(ctx context.Context, cli client.Client, pods []corev1.Pod) error {
	sctName := s.GetNameForResource(utils.Secret)
	svcName := s.GetNameForResource(utils.HeadlessSVC)
	nameSpace := s.Namespace

	secret := &corev1.Secret{}
	if err := cli.Get(context.TODO(),
		types.NamespacedName{
			Namespace: nameSpace,
			Name:      sctName,
		},
		secret,
	); err != nil {
		log.V(1).Info("secret not found", "name", sctName)
		return nil
	}

	rootPasswd, ok := secret.Data["root-password"]
	if !ok {
		return fmt.Errorf("failed to get the root password: %s", rootPasswd)
	}

	for _, pod := range pods {
		podName := pod.Name
		ordinal, _ := utils.GetOrdinal(podName)
		host := fmt.Sprintf("%s.%s.%s", podName, fmt.Sprintf("%s-%d", svcName, ordinal), nameSpace)
		index := s.getNodeStatusIndex(host)
		node := &s.Status.Nodes[index]
		node.Message = ""

		isLeader, err := checkRole(host, rootPasswd)
		if err != nil {
			log.Error(err, "failed to check the node role", "node", node.Name)
			node.Message = err.Error()
		}
		// update apiv1alpha1.NodeConditionLeader.
		s.updateNodeCondition(node, int(apiv1alpha1.IndexLeader), isLeader)

		isLagged, isReplicating, isReadOnly := corev1.ConditionUnknown, corev1.ConditionUnknown, corev1.ConditionUnknown
		sqlRunner, closeConn, err := s.SQLRunnerFactory(internal.NewConfigFromClusterKey(
			s.cli, s.MysqlCluster.GetClusterKey(), utils.OperatorUser, host))
		defer closeConn()
		if err != nil {
			log.Error(err, "failed to connect the mysql", "node", node.Name)
			node.Message = err.Error()
		} else {
			isLagged, isReplicating, err = internal.CheckSlaveStatusWithRetry(sqlRunner, checkNodeStatusRetry)
			if err != nil {
				log.Error(err, "failed to check slave status", "node", node.Name)
				node.Message = err.Error()
			}

			isReadOnly, err = internal.CheckReadOnly(sqlRunner)
			if err != nil {
				log.Error(err, "failed to check read only", "node", node.Name)
				node.Message = err.Error()
			}

			if !utils.ExistUpdateFile() &&
				isLeader == corev1.ConditionTrue &&
				isReadOnly != corev1.ConditionFalse {
				log.V(1).Info("try to correct the leader writeable", "node", node.Name)
				sqlRunner.QueryExec(internal.NewQuery("SET GLOBAL read_only=off"))
				sqlRunner.QueryExec(internal.NewQuery("SET GLOBAL super_read_only=off"))
			}
		}

		// update apiv1alpha1.NodeConditionLagged.
		s.updateNodeCondition(node, int(apiv1alpha1.IndexLagged), isLagged)
		// update apiv1alpha1.NodeConditionReplicating.
		s.updateNodeCondition(node, int(apiv1alpha1.IndexReplicating), isReplicating)
		// update apiv1alpha1.NodeConditionReadOnly.
		s.updateNodeCondition(node, int(apiv1alpha1.IndexReadOnly), isReadOnly)

		if err = s.setPodHealthy(ctx, &pod, node); err != nil {
			log.Error(err, "cannot update pod", "name", podName, "namespace", pod.Namespace)
		}
	}

	return nil
}

// getNodeStatusIndex get the node index in the status.
func (s *StatusSyncer) getNodeStatusIndex(name string) int {
	len := len(s.Status.Nodes)
	for i := 0; i < len; i++ {
		if s.Status.Nodes[i].Name == name {
			return i
		}
	}

	lastTransitionTime := metav1.NewTime(time.Now())
	status := apiv1alpha1.NodeStatus{
		Name: name,
		Conditions: []apiv1alpha1.NodeCondition{
			{
				Type:               apiv1alpha1.NodeConditionLagged,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               apiv1alpha1.NodeConditionLeader,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               apiv1alpha1.NodeConditionReadOnly,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               apiv1alpha1.NodeConditionReplicating,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
		},
	}
	s.Status.Nodes = append(s.Status.Nodes, status)
	return len
}

// updateNodeCondition update the node condition.
func (s *StatusSyncer) updateNodeCondition(node *apiv1alpha1.NodeStatus, idx int, status corev1.ConditionStatus) {
	if node.Conditions[idx].Status != status {
		t := time.Now()
		log.V(3).Info(fmt.Sprintf("Found status change for node %q condition %q: %q -> %q; setting lastTransitionTime to %v",
			node.Name, node.Conditions[idx].Type, node.Conditions[idx].Status, status, t))
		node.Conditions[idx].Status = status
		node.Conditions[idx].LastTransitionTime = metav1.NewTime(t)
	}
}

// checkRole used to check whether the mysql role is leader.
func checkRole(host string, rootPasswd []byte) (corev1.ConditionStatus, error) {
	body, err := xenonHttpRequest(host, "GET", "/v1/raft/status", rootPasswd, nil)
	if err != nil {
		return corev1.ConditionUnknown, err
	}

	var out map[string]interface{}
	if err = unmarshalJSON(body, &out); err != nil {
		return corev1.ConditionUnknown, err
	}

	if out["state"] == "LEADER" {
		return corev1.ConditionTrue, nil
	}

	if out["state"] == "FOLLOWER" {
		return corev1.ConditionFalse, nil
	}

	return corev1.ConditionUnknown, nil
}

// setPodHealthy set the pod lable healthy.
func (s *StatusSyncer) setPodHealthy(ctx context.Context, pod *corev1.Pod, node *apiv1alpha1.NodeStatus) error {
	healthy := "no"
	if node.Conditions[apiv1alpha1.IndexLagged].Status == corev1.ConditionFalse {
		if node.Conditions[apiv1alpha1.IndexLeader].Status == corev1.ConditionFalse &&
			node.Conditions[apiv1alpha1.IndexReadOnly].Status == corev1.ConditionTrue &&
			node.Conditions[apiv1alpha1.IndexReplicating].Status == corev1.ConditionTrue {
			healthy = "yes"
		} else if node.Conditions[apiv1alpha1.IndexLeader].Status == corev1.ConditionTrue &&
			node.Conditions[apiv1alpha1.IndexReplicating].Status == corev1.ConditionFalse &&
			node.Conditions[apiv1alpha1.IndexReadOnly].Status == corev1.ConditionFalse {
			healthy = "yes"
		}
	}

	if pod.Labels["healthy"] != healthy {
		pod.Labels["healthy"] = healthy
		if err := s.cli.Update(ctx, pod); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func unmarshalJSON(in io.Reader, obj interface{}) error {
	body, err := ioutil.ReadAll(in)
	if err != nil {
		return fmt.Errorf("io read error: %s", err)
	}

	if err = json.Unmarshal(body, obj); err != nil {
		log.V(1).Info("error unmarshal data", "body", string(body))
		return err
	}

	return nil
}
