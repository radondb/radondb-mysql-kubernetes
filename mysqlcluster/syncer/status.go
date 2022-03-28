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
	"fmt"
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

// The retry time for check node status.
const checkNodeStatusRetry = 3

// StatusSyncer used to update the status.
type StatusSyncer struct {
	*mysqlcluster.MysqlCluster

	cli client.Client

	// Mysql query runner.
	internal.SQLRunnerFactory
	// XenonExecutor is used to execute Xenon HTTP instructions.
	internal.XenonExecutor
}

// NewStatusSyncer returns a pointer to StatusSyncer.
func NewStatusSyncer(c *mysqlcluster.MysqlCluster, cli client.Client, sqlRunnerFactory internal.SQLRunnerFactory, xenonExecutor internal.XenonExecutor) *StatusSyncer {
	return &StatusSyncer{
		MysqlCluster:     c,
		cli:              cli,
		SQLRunnerFactory: sqlRunnerFactory,
		XenonExecutor:    xenonExecutor,
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
	s.updateClusterStatus()

	readyNodes, error := s.getReadyPods(ctx)
	if error != nil {
		return syncer.SyncResult{}, error
	}

	s.Status.ReadyNodes = len(readyNodes)
	if !s.Status.ClusterClosed(*s.Spec.Replicas) {
		// Only reconcile node which is ready.
		if err := s.reconcileXenon(s.expectXenonNodes(readyNodes)); err != nil {
			s.Status.AppendErrorCondition("", fmt.Sprintf("%s", err))
		}
	}

	if s.Status.ClusterReady(s.PodReady(s.Status.ReadyNodes)) {
		s.SetClusterReady()
	}

	// Update ready nodes' status.
	return syncer.SyncResult{}, s.updateNodeStatus(ctx, s.cli, readyNodes)
}

// Ready node must meet the following conditions:
// 1. pod phase is running.
// 2. podReady condition is true.
// 3. pod`s xenon can be pinged.
// Notice: worker crash is still satisfied conditions 1 and 2.
func (s *StatusSyncer) getReadyPods(ctx context.Context) ([]corev1.Pod, error) {
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
		return nil, err
	}

	// get ready nodes.
	var readyNodes []corev1.Pod
	for _, pod := range list.Items {
		if !s.podInSpec(pod.Name) {
			continue
		}
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}
		if pod.ObjectMeta.Labels[utils.LableRebuild] == "true" {
			if err := s.AutoRebuild(ctx, &pod); err != nil {
				log.Error(err, "failed to AutoRebuild", "pod", pod.Name, "namespace", pod.Namespace)
			}
			continue
		}
		if err := s.XenonExecutor.XenonPing(s.GetPodHostName(pod.Name)); err != nil {
			s.removeInvalidNodeStatus(s.GetPodHostName(pod.Name))
			log.Error(err, "failed to ping", "pod", pod.Name, "namespace", pod.Namespace)
			continue
		}

		for _, cond := range pod.Status.Conditions {
			switch cond.Type {
			case corev1.PodReady:
				if cond.Status == corev1.ConditionTrue {
					readyNodes = append(readyNodes, pod)
				}
			case corev1.PodScheduled:
				// Temporary unschedulable occurs whenever a new pod is created.
				if cond.Reason == corev1.PodReasonUnschedulable &&
					cond.LastTransitionTime.Time.Before(time.Now().Add(-1*time.Minute)) {
					s.Status.AppendErrorCondition(corev1.PodReasonUnschedulable, cond.Message)
				}
			}
		}
	}
	return readyNodes, nil
}

// podInSpec returns true if the pod ordinal is meets the expectations.
func (s *StatusSyncer) podInSpec(podName string) bool {
	ordinal, err := utils.GetOrdinal(podName)
	if err != nil {
		return false
	}
	if int32(ordinal) >= *s.Spec.Replicas {
		return false
	}
	return true
}

func (s *StatusSyncer) updateClusterStatus() {
	replicas := s.Spec.Replicas

	switch {
	case s.Status.ClusterInitializing():
		s.SetClusterInitializing()
	case s.Status.ClusterClosed(*replicas):
		s.SetClusterClosed()
	case s.Status.ClusterError():
		s.SetClusterError()
	case s.Status.ClusterUpdating():
		s.SetClusterUpdating()
	case s.Status.ClusterScaling():
		s.SetClusterScaling(*replicas)
	default:
	}
}

// Rebuild Pod by deleting and creating it.
// Notice: This function just delete Pod and PVC,
// then after k8s recreate pod, it will clone and initial it.
func (s *StatusSyncer) AutoRebuild(ctx context.Context, pod *corev1.Pod) error {
	ordinal, err := utils.GetOrdinal(pod.Name)
	if err != nil {
		return err

	}
	// Set Pod UnHealthy.
	pod.Labels["healthy"] = "no"
	if err := s.cli.Update(ctx, pod); err != nil {
		return err
	}
	// Delete the Pod.
	if err := s.cli.Delete(ctx, pod); err != nil {
		return err
	}
	// Delete the pvc.
	pvcName := fmt.Sprintf("%s-%s-%d", utils.DataVolumeName,
		s.GetNameForResource(utils.StatefulSet), ordinal)
	pvc := corev1.PersistentVolumeClaim{}

	if err := s.cli.Get(ctx,
		types.NamespacedName{Name: pvcName, Namespace: s.Namespace},
		&pvc); err != nil {
		return err
	}
	if err := s.cli.Delete(ctx, &pvc); err != nil {
		return err
	}
	return nil
}

// updateNodeStatus update the node status.
func (s *StatusSyncer) updateNodeStatus(ctx context.Context, cli client.Client, pods []corev1.Pod) error {
	for _, pod := range pods {
		host := s.GetPodHostName(pod.Name)
		index := s.getNodeStatusIndex(host)
		node := &s.Status.Nodes[index]
		node.Message = ""

		if err := s.updateNodeRaftStatus(node); err != nil {
			log.Error(err, "failed to get/update node raft status", "node", node.Name)
			node.Message = err.Error()
		}

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
				node.RaftStatus.Role == string(utils.Leader) &&
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

		if err = s.updatePodLabel(ctx, &pod, node); err != nil {
			log.Error(err, "failed to update labels", "pod", pod.Name, "namespace", pod.Namespace)
		}
	}

	// Delete node status of nodes that have been deleted.
	s.truncateNodeStatus()

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

func (s *StatusSyncer) truncateNodeStatus() {
	for index, nodeStatus := range s.Status.Nodes {
		if !utils.StringExistIn(s.expectNodeStatus(), nodeStatus.Name) {
			s.Status.Nodes = append(s.Status.Nodes[:index], s.Status.Nodes[index+1:]...)
		}
	}
}

func (s *StatusSyncer) expectNodeStatus() []string {
	expectXenonNodes := []string{}
	for i := 0; i < int(*s.Spec.Replicas); i++ {
		expectXenonNodes = append(expectXenonNodes, s.GetPodHostName(fmt.Sprintf("%s-%d", s.GetNameForResource(utils.HeadlessSVC), i)))
	}
	return expectXenonNodes
}

func (s *StatusSyncer) removeInvalidNodeStatus(name string) {
	for i, node := range s.Status.Nodes {
		if node.Name == name {
			s.Status.Nodes = append(s.Status.Nodes[:i], s.Status.Nodes[i+1:]...)
			return
		}
	}
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

// updateNodeRaftStatus Update Node RaftStatus.
func (s *StatusSyncer) updateNodeRaftStatus(node *apiv1alpha1.NodeStatus) error {
	isLeader := corev1.ConditionFalse
	node.RaftStatus = apiv1alpha1.RaftStatus{
		Role:   string(utils.Unknown),
		Leader: "UNKNOWN",
		Nodes:  nil,
	}

	raftStatus, err := s.XenonExecutor.RaftStatus(node.Name)
	if err == nil && raftStatus != nil {
		node.RaftStatus = *raftStatus
		if raftStatus.Role == string(utils.Leader) {
			isLeader = corev1.ConditionTrue
		}
	}

	// update apiv1alpha1.NodeConditionLeader.
	s.updateNodeCondition(node, int(apiv1alpha1.IndexLeader), isLeader)
	return err
}

func (s *StatusSyncer) reconcileXenon(readyNodes []string) error {
	for _, nodeStatus := range s.Status.Nodes {
		if !utils.StringExistIn(readyNodes, fmt.Sprintf("%s:%d", nodeStatus.Name, utils.XenonPort)) {
			continue
		}
		toRemove := utils.StringDiffIn(nodeStatus.RaftStatus.Nodes, readyNodes)
		if err := s.removeNodesFromXenon(nodeStatus.Name, toRemove); err != nil {
			return err
		}
		toAdd := utils.StringDiffIn(readyNodes, nodeStatus.RaftStatus.Nodes)
		if err := s.addNodesInXenon(nodeStatus.Name, toAdd); err != nil {
			return err
		}
	}
	return nil
}

func (s *StatusSyncer) expectXenonNodes(pods []corev1.Pod) []string {
	expectXenonNodes := []string{}
	for _, pod := range pods {
		expectXenonNodes = append(expectXenonNodes, fmt.Sprintf("%s:%d", s.GetPodHostName(pod.Name), utils.XenonPort))
	}
	return expectXenonNodes
}

func (s *StatusSyncer) removeNodesFromXenon(host string, toRemove []string) error {
	if err := s.XenonExecutor.XenonPing(host); err != nil {
		return err
	}
	for _, removeHost := range toRemove {
		if err := s.XenonExecutor.ClusterRemove(host, removeHost); err != nil {
			return err
		}
	}
	return nil
}

func (s *StatusSyncer) addNodesInXenon(host string, toAdd []string) error {
	if err := s.XenonExecutor.XenonPing(host); err != nil {
		return err
	}
	for _, addHost := range toAdd {
		if err := s.XenonExecutor.ClusterAdd(host, addHost); err != nil {
			return err
		}
	}
	return nil
}

// updatePodLabel update the pod lables.
func (s *StatusSyncer) updatePodLabel(ctx context.Context, pod *corev1.Pod, node *apiv1alpha1.NodeStatus) error {
	healthy := "no"
	isPodLabelsUpdated := false
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
		isPodLabelsUpdated = true
	}
	if pod.Labels["role"] != node.RaftStatus.Role {
		pod.Labels["role"] = node.RaftStatus.Role
		isPodLabelsUpdated = true
	}
	if isPodLabelsUpdated {
		if err := s.cli.Update(ctx, pod); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func (s *StatusSyncer) GetPodHostName(podName string) string {
	return fmt.Sprintf("%s.%s.%s", podName, s.GetNameForResource(utils.HeadlessSVC), s.Namespace)
}

func (s *StatusSyncer) SetClusterReady() {
	s.Status.AppendReadyCondition()
	s.Status.State = apiv1alpha1.ClusterReadyState
}

func (s *StatusSyncer) SetClusterInitializing() {
	s.Status.AppendInitCondition(*s.Spec.Replicas)
	s.Status.State = apiv1alpha1.ClusterInitState
}

func (s *StatusSyncer) SetClusterClosed() {
	s.Status.AppendClosedCondition()
	s.Status.State = apiv1alpha1.ClusterCloseState
}

func (s *StatusSyncer) SetClusterError() {
	s.Status.State = apiv1alpha1.ClusterErrorState
}

func (s *StatusSyncer) SetClusterUpdating() {
	s.Status.State = apiv1alpha1.ClusterUpdateState
}

func (s *StatusSyncer) SetClusterScaling(replicas int32) {
	if s.Status.ReadyNodes > int(replicas) {
		s.Status.State = apiv1alpha1.ClusterScaleInState
	}
	if s.Status.ReadyNodes < int(replicas) {
		s.Status.State = apiv1alpha1.ClusterScaleOutState
	}
}
