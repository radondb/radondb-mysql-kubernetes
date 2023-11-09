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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/presslabs/controller-util/pkg/syncer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"

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
	// XenonExecutor is used to execute Xenon HTTP instructions.
	internal.XenonExecutor
	// Logger
	log logr.Logger
}

// NewStatusSyncer returns a pointer to StatusSyncer.
func NewStatusSyncer(c *mysqlcluster.MysqlCluster, cli client.Client, sqlRunnerFactory internal.SQLRunnerFactory, xenonExecutor internal.XenonExecutor) *StatusSyncer {
	return &StatusSyncer{
		MysqlCluster:     c,
		cli:              cli,
		SQLRunnerFactory: sqlRunnerFactory,
		XenonExecutor:    xenonExecutor,
		log:              logf.Log.WithName("syncer.StatusSyncer"),
	}
}

// Object returns the object for which sync applies.
func (s *StatusSyncer) Object() interface{} { return nil }

// GetObject returns the object for which sync applies
// Deprecated: use github.com/presslabs/controller-util/pkg/syncer.Object() instead.
func (s *StatusSyncer) GetObject() interface{} { return nil }

// Owner returns the object owner or nil if object does not have one.
func (s *StatusSyncer) ObjectOwner() runtime.Object { return s.MysqlCluster }

// GetOwner returns the object owner or nil if object does not have one.
// Deprecated: use github.com/presslabs/controller-util/pkg/syncer.ObjectOwner() instead.
func (s *StatusSyncer) GetOwner() runtime.Object { return s.MysqlCluster }

// Sync persists data into the external store.
func (s *StatusSyncer) Sync(ctx context.Context) (syncer.SyncResult, error) {
	clusterCondition := s.updateClusterStatus()
	labelSelector := s.GetLabels().AsSelector()
	// Find the pods that revision is old.
	r, err := labels.NewRequirement("readonly", selection.DoesNotExist, []string{})
	if err != nil {
		s.log.V(1).Info("failed to create label requirement", "error", err)
		return syncer.SyncResult{}, err
	}
	labelSelector = labelSelector.Add(*r)

	list := corev1.PodList{}
	err = s.cli.List(
		ctx,
		&list,
		&client.ListOptions{
			Namespace:     s.Namespace,
			LabelSelector: labelSelector,
		},
	)
	if err != nil {
		return syncer.SyncResult{}, err
	}

	// get ready nodes.
	var readyNodes []corev1.Pod
	var PodLeader, PodTryLeader *corev1.Pod
	PodLeader, PodTryLeader = nil, nil
	for _, pod := range list.Items {
		if len(pod.ObjectMeta.Labels[utils.LableRebuild]) > 0 {
			if err := s.AutoRebuild(ctx, &pod, list.Items); err != nil {
				s.log.Error(err, "failed to AutoRebuild", "pod", pod.Name, "namespace", pod.Namespace)
			}
			continue
		}
		if pod.ObjectMeta.Labels != nil {
			if len(pod.ObjectMeta.Labels[utils.LabelTryLeader]) != 0 {
				PodTryLeader = &pod
			}
			if pod.ObjectMeta.Labels["role"] == string(utils.Leader) {
				PodLeader = &pod
			}
		}

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
	// try leader
	if PodTryLeader != nil {
		if PodLeader != nil {
			if err := s.SetLeaderReadOnly(PodLeader); err != nil {
				s.log.Info("set leader readonly", "error", err.Error())
			}
		}
		if err := s.TryLeader(ctx, PodTryLeader); err != nil {
			s.log.Error(err, "failed to Try leader", "pod", PodTryLeader.Name, "namespace", PodTryLeader.Namespace)
		}
	}
	s.Status.ReadyNodes = len(readyNodes)
	if s.Status.ReadyNodes == int(*s.Spec.Replicas) && int(*s.Spec.Replicas) != 0 {
		if err := s.reconcileXenon(s.Status.ReadyNodes); err != nil {
			clusterCondition.Message = fmt.Sprintf("%s", err)
			clusterCondition.Type = apiv1alpha1.ConditionError
		} else {
			s.Status.State = apiv1alpha1.ClusterReadyState
			clusterCondition.Type = apiv1alpha1.ConditionReady
		}
	}

	if len(s.Status.Conditions) == 0 {
		s.Status.Conditions = append(s.Status.Conditions, clusterCondition)
	} else {
		lastCond := s.Status.Conditions[len(s.Status.Conditions)-1]
		if lastCond.Type != clusterCondition.Type {
			s.Status.Conditions = append(s.Status.Conditions, clusterCondition)
		}
	}

	//update read slave status for remote cluster
	if err := s.clusterSlaveCheck(); err != nil {
		//Notice!!! remote Cluster slave  node fail, just show the error log, do not return here!
		s.log.Error(err, "slace for Remote cluster fail", "namespace", s.Namespace)
	}
	//(RO) because the ReadOnly Pods create after the cluster ready, so the ReadOnly pods are always
	// the last part of node status
	if err := s.updateReadOnlyNodeStatus(ctx, s.cli, list.Items); err != nil {
		//Notice!!! ReadOnly node fail, just show the error log, do not return here!
		s.log.Error(err, "ReadOnly pod fail", "namespace", s.Namespace)
	}
	// conditions := s.Status.Conditions
	// sort.Slice(conditions, func(i, j int) bool {
	// 	if conditions[i].Type < conditions[j].Type {
	// 		return true
	// 	} else if conditions[i].Type == conditions[j].Type &&
	// 		conditions[i].LastTransitionTime.Before(&conditions[j].LastTransitionTime) {
	// 		return true
	// 	} else {
	// 		return false
	// 	}
	// })
	// s.Status.Conditions = conditions
	if len(s.Status.Conditions) > maxStatusesQuantity {
		s.Status.Conditions = s.Status.Conditions[len(s.Status.Conditions)-maxStatusesQuantity:]
	}
	// update backup Status
	s.updateLastBackup()

	// Update all nodes' status.
	return syncer.SyncResult{}, s.updateNodeStatus(ctx, s.cli, list.Items)
}

func (s *StatusSyncer) updateLastBackup() error {
	// 1. fetch all finished backup cr
	backupsList := &apiv1alpha1.BackupList{}
	labelSet := labels.Set{"cluster": s.Name}
	if err := s.cli.List(context.TODO(), backupsList, &client.ListOptions{
		Namespace: s.Namespace, LabelSelector: labelSet.AsSelector(),
	}); err != nil {
		return err
	}
	var finisheds []apiv1alpha1.Backup
	for _, b := range backupsList.Items {
		if b.Status.Completed {
			finisheds = append(finisheds, b)
		}
	}
	// 2. sort descent
	sort.Slice(finisheds, func(i, j int) bool {
		return finisheds[j].ObjectMeta.CreationTimestamp.Before(&finisheds[i].ObjectMeta.CreationTimestamp)
	})
	// 3. get first backup which has backup Name
	for _, b := range finisheds {
		if len(b.Status.BackupName) != 0 &&
			s.CreationTimestamp.Before(&b.ObjectMeta.CreationTimestamp) {
			s.Status.LastBackup = b.Status.BackupName
			s.Status.LastBackupGtid = b.Status.Gtid
			s.Status.LastBackupTime = b.ObjectMeta.CreationTimestamp
			break
		}

	}

	return nil
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
		if int(*s.Spec.Replicas) > s.Status.ReadyNodes {
			clusterCondition.Type = apiv1alpha1.ConditionScaleOut
			s.Status.State = apiv1alpha1.ClusterScaleOutState
			return clusterCondition
		} else if int(*s.Spec.Replicas) < s.Status.ReadyNodes {
			clusterCondition.Type = apiv1alpha1.ConditionScaleIn
			s.Status.State = apiv1alpha1.ClusterScaleInState
			return clusterCondition
		}
	}

	clusterCondition.Type = apiv1alpha1.ClusterConditionType(oldState)
	return clusterCondition
}

// Rebuild Pod by deleting and creating it.
// Notice: This function just delete Pod and PVC,
// then after k8s recreate pod, it will clone and initial it.
func (s *StatusSyncer) AutoRebuild(ctx context.Context, pod *corev1.Pod, items []corev1.Pod) error {
	ordinal, err := utils.GetOrdinal(pod.Name)
	if err != nil {
		return err

	}
	if pod.ObjectMeta.Labels[utils.LableRebuild] != "true" {
		podNumber, err := strconv.Atoi(pod.ObjectMeta.Labels[utils.LableRebuild])
		if err != nil {
			return fmt.Errorf("rebuild label should be true, or number")
		}
		for _, other := range items {
			ord, err2 := utils.GetOrdinal(other.Name)
			if err2 != nil {
				return err

			}
			if ord == podNumber {
				other.Labels[utils.LabelRebuildFrom] = "true"
				if err := s.cli.Update(ctx, &other); err != nil {
					return err
				}
				break
			}
		}
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
	if s.Spec.MysqlVersion == "8.0" {
		s.log.Info("Do not delete pvc")
		return nil
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
	closeCh := make(chan func())
	for _, pod := range pods {
		podName := pod.Name
		host := fmt.Sprintf("%s.%s.%s", podName, s.GetNameForResource(utils.HeadlessSVC), s.Namespace)
		index := s.getNodeStatusIndex(host)
		node := &s.Status.Nodes[index]
		node.Message = ""

		if err := s.updateNodeRaftStatus(node); err != nil {
			s.log.V(1).Info("failed to get/update node raft status", "node", node.Name, "error", err)
			node.Message = err.Error()
		}

		isLagged, isReplicating, isReadOnly := corev1.ConditionUnknown, corev1.ConditionUnknown, corev1.ConditionUnknown
		var sqlRunner internal.SQLRunner
		var closeConn func()
		errCh := make(chan error)
		go func(sqlRunner *internal.SQLRunner, errCh chan error, closeCh chan func()) {
			var err error
			*sqlRunner, closeConn, err = s.SQLRunnerFactory(internal.NewConfigFromClusterKey(
				s.cli, s.MysqlCluster.GetClusterKey(), utils.OperatorUser, host))
			if err != nil {
				s.log.V(1).Info("failed to get sql runner", "node", node.Name, "error", err)
				errCh <- err
				return
			}
			if closeConn != nil {
				closeCh <- closeConn
				return
			}
			errCh <- nil
		}(&sqlRunner, errCh, closeCh)

		var err error
		select {
		case <-errCh:
		case closeConn := <-closeCh:
			defer closeConn()
		case <-time.After(time.Second * 5):
		}
		if sqlRunner != nil {
			isLagged, isReplicating, err = internal.CheckSlaveStatusWithRetry(sqlRunner, checkNodeStatusRetry, s.Spec.ReplicaLag)
			if err != nil {
				s.log.V(1).Info("failed to check slave status", "node", node.Name, "error", err)
				node.Message = err.Error()
			}

			isReadOnly, err = internal.CheckReadOnly(sqlRunner)
			if err != nil {
				s.log.V(1).Info("failed to check read only", "node", node.Name, "error", err)
				node.Message = err.Error()
			}
			// move it to mysql readiness
			// if !utils.ExistUpdateFile() &&
			// 	node.RaftStatus.Role == string(utils.Leader) &&
			// 	isReadOnly != corev1.ConditionFalse {
			// 	s.log.V(1).Info("try to correct the leader writeable", "node", node.Name)
			// 	sqlRunner.QueryExec(internal.NewQuery("SET GLOBAL read_only=off"))
			// 	sqlRunner.QueryExec(internal.NewQuery("SET GLOBAL super_read_only=off"))
			// }
		}

		// update apiv1alpha1.NodeConditionLagged.
		s.updateNodeCondition(node, int(apiv1alpha1.IndexLagged), isLagged)
		// update apiv1alpha1.NodeConditionReplicating.
		s.updateNodeCondition(node, int(apiv1alpha1.IndexReplicating), isReplicating)
		// update apiv1alpha1.NodeConditionReadOnly.
		s.updateNodeCondition(node, int(apiv1alpha1.IndexReadOnly), isReadOnly)

		if err = s.updatePodLabel(ctx, &pod, node); err != nil {
			s.log.V(1).Info("failed to update labels", "pod", pod.Name, "error", err)
		}
	}

	// Delete node status of nodes that have been deleted.
	if len(s.Status.Nodes) > len(pods) {
		trimNodes := s.Status.Nodes[:len(pods)]
		if s.Spec.ReadOnlys != nil {
			// get the last parts of ReadOnly Nodes.
			roNodes := s.Status.Nodes[len(s.Status.Nodes)-int(s.Spec.ReadOnlys.Num) : len(s.Status.Nodes)]
			trimNodes = append(trimNodes, roNodes...)
		}
		s.Status.Nodes = trimNodes

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
		s.log.V(3).Info(fmt.Sprintf("Found status change for node %q condition %q: %q -> %q; setting lastTransitionTime to %v",
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

func (s *StatusSyncer) reconcileXenon(readyNodes int) error {
	expectXenonNodes := s.getExpectXenonNodes(readyNodes)
	// filter nodes with out ro
	for _, nodeStatus := range s.Status.Nodes {
		if strings.HasPrefix(nodeStatus.Name, s.GetNameForResource(utils.ReadOnlyHeadlessSVC)) {
			continue
		}

		toRemove := utils.StringDiffIn(nodeStatus.RaftStatus.Nodes, expectXenonNodes)
		if err := s.removeNodesFromXenon(nodeStatus.Name, toRemove); err != nil {
			return err
		}
		toAdd := utils.StringDiffIn(expectXenonNodes, nodeStatus.RaftStatus.Nodes)
		if err := s.addNodesInXenon(nodeStatus.Name, toAdd); err != nil {
			return err
		}
	}
	return nil
}

func (s *StatusSyncer) getExpectXenonNodes(readyNodes int) []string {
	expectXenonNodes := []string{}
	for i := 0; i < readyNodes; i++ {
		expectXenonNodes = append(expectXenonNodes, fmt.Sprintf("%s:%d", s.GetPodHostName(i), utils.XenonPort))
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
	oldPod := pod.DeepCopy()
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
	if pod.DeletionTimestamp != nil || pod.Status.Phase != corev1.PodRunning {
		healthy = "no"
		node.RaftStatus.Role = string(utils.Unknown)
	}
	// update healthy no if container is not ready.
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == utils.ContainerMysqlName && !containerStatus.Ready {
			healthy = "no"
			isPodLabelsUpdated = true
			break
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
		if err := s.cli.Patch(ctx, pod, client.MergeFrom(oldPod)); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

// Update the Readonly node status
func (s *StatusSyncer) updateReadOnlyNodeStatus(ctx context.Context, cli client.Client, cluster_pods []corev1.Pod) error {
	labels := s.GetLabels()
	labels["app.kubernetes.io/name"] = "mysql-readonly"
	labels["readonly"] = "true"

	list := corev1.PodList{}
	err := s.cli.List(
		ctx,
		&list,
		&client.ListOptions{
			Namespace:     s.Namespace,
			LabelSelector: labels.AsSelector(),
		},
	)
	if err != nil {
		return err
	}
	// (RO) 1. status update
	if err := s.RoCheckStatus(ctx, cli, list.Items); err != nil {
		return err
	}
	for _, pod := range list.Items {
		if err := s.DoRoRebuild(ctx, &pod, cluster_pods); err != nil {
			// Cannot return ,just print log.
			s.log.Error(err, "failed to AutoRebuild", "pod", pod.Name, "namespace", pod.Namespace)
		}
	}

	return nil
}

// (RO) check readonly and check semi replication
func (s *StatusSyncer) RoCheckStatus(ctx context.Context, cli client.Client, pods []corev1.Pod) error {
	closeCh := make(chan func())
	for _, pod := range pods {
		podName := pod.Name
		host := fmt.Sprintf("%s.%s.%s", podName, s.GetNameForResource(utils.ReadOnlyHeadlessSVC), s.Namespace)
		index := s.getRoStatusIndex(host)
		node := &s.Status.Nodes[index]

		isInitial, isReadonly, isCloseSemi, isReplicating := corev1.ConditionUnknown, corev1.ConditionUnknown, corev1.ConditionUnknown, corev1.ConditionUnknown
		isSupperReadOnly := corev1.ConditionUnknown
		if pod.Status.Phase == corev1.PodRunning {
			isInitial = corev1.ConditionTrue
		}
		var sqlRunner internal.SQLRunner
		var closeConn func()
		errCh := make(chan error)
		go func(sqlRunner *internal.SQLRunner, errCh chan error, closeCh chan func()) {
			var err error
			*sqlRunner, closeConn, err = s.SQLRunnerFactory(internal.NewConfigFromClusterKey(
				s.cli, s.MysqlCluster.GetClusterKey(), utils.OperatorUser, host))
			if err != nil {
				s.log.V(1).Info("failed to get sql runner", "node", node.Name, "error", err)
				errCh <- err
				return
			}
			if closeConn != nil {
				closeCh <- closeConn
				return
			}
			errCh <- nil
		}(&sqlRunner, errCh, closeCh)

		select {
		case <-errCh:
		case closeConn := <-closeCh:
			defer closeConn()
		case <-time.After(time.Second * 5):
		}
		var err error

		if sqlRunner != nil {
			// (RO) 1. add check readonly
			if isReadonly, err = internal.CheckReadOnly(sqlRunner); err != nil {
				node.Message = err.Error()
			}
			if isSupperReadOnly, err = internal.CheckSuperReadOnly(sqlRunner); err != nil {
				node.Message = err.Error()
			}
			// 2. set rpl_semi_sync_slave_enabled off
			if status, err := internal.CheckSemSync(sqlRunner); err != nil {
				node.Message = err.Error()

			} else if status == corev1.ConditionFalse {
				isCloseSemi = corev1.ConditionTrue
			}
			// 3. change master
			if _, isReplicating, err = internal.CheckSlaveStatus(sqlRunner, s.Spec.ReplicaLag); err != nil {
				parts := strings.Split(node.Message, " ")
				now := time.Now().Unix()
				var last int64
				var err2 error
				if len(parts[0]) == 0 {
					node.Message = fmt.Sprintf("%d unixtime error:", now) + err.Error()
				} else if last, err2 = strconv.ParseInt(parts[0], 10, 64); err2 != nil {
					s.log.Info("get times from message", "times", last)
				}
				if now-last > 10 {
					node.Message = fmt.Sprintf("%d unixseconds error:", now) + err.Error()
				}

			} else {
				node.Message = ""
			}
		}
		//update node Rostatus
		node.RoStatus = &apiv1alpha1.RoStatus{
			ReadOnly:    isReadonly == corev1.ConditionTrue && isSupperReadOnly == corev1.ConditionTrue,
			Replication: isReplicating == corev1.ConditionTrue && isCloseSemi == corev1.ConditionTrue,
			Master: func(cr *StatusSyncer) string {
				if *cr.Spec.Replicas == 1 {
					return fmt.Sprintf("%s-0.%s.%s", cr.GetNameForResource(utils.StatefulSet),
						s.GetNameForResource(utils.StatefulSet), cr.Namespace)
				}
				if cr.Spec.ReadOnlys != nil {
					if len(cr.Spec.ReadOnlys.Host) == 0 {
						return fmt.Sprintf("%s-follower", cr.Name)
					} else {
						return fmt.Sprintf("%s.%s.%s", cr.Spec.ReadOnlys.Host, cr.GetNameForResource(utils.StatefulSet), cr.Namespace)
					}
				} else {
					return ""
				}
			}(s),
		}
		// update apiv1alpha1.NodeConditionLagged.
		s.updateNodeCondition(node, 0, isInitial)
		// readonly
		s.updateNodeCondition(node, int(apiv1alpha1.IndexRoReadOnly-apiv1alpha1.IndexRoInit), isReadonly)
		// close semi check
		s.updateNodeCondition(node, int(apiv1alpha1.IndexRoSemiClose-apiv1alpha1.IndexRoInit), isCloseSemi)
		// update apiv1alpha1.NodeConditionReplicating.
		s.updateNodeCondition(node, int(apiv1alpha1.IndexRoReplicating-apiv1alpha1.IndexRoInit), isReplicating)

	}
	return nil
}

// (RO) 3. get status index and set conditon
func (s *StatusSyncer) getRoStatusIndex(name string) int {
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
				Type:               apiv1alpha1.NodeConditionRoInitial,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               apiv1alpha1.NodeConditionRoReadOnly,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               apiv1alpha1.NodeConditionRoSemiClose,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
			{
				Type:               apiv1alpha1.NodeConditionRoReplicating,
				Status:             corev1.ConditionUnknown,
				LastTransitionTime: lastTransitionTime,
			},
		},
	}
	s.Status.Nodes = append(s.Status.Nodes, status)
	return len
}

func (s *StatusSyncer) DoRoRebuild(ctx context.Context, pod *corev1.Pod, items []corev1.Pod) error {
	if len(pod.ObjectMeta.Labels[utils.LableRebuild]) == 0 {
		return nil
	}
	ordinal, err := utils.GetOrdinal(pod.Name)
	if err != nil {
		return err

	}
	if pod.ObjectMeta.Labels[utils.LableRebuild] != "true" {
		podNumber, err := strconv.Atoi(pod.ObjectMeta.Labels[utils.LableRebuild])
		if err != nil {
			return fmt.Errorf("rebuild label should be true, or number")
		}
		for _, other := range items {
			ord, err2 := utils.GetOrdinal(other.Name)
			if err2 != nil {
				return err

			}
			if ord == podNumber {
				other.Labels[utils.LabelRebuildFrom] = "true"
				if err := s.cli.Update(ctx, &other); err != nil {
					return err
				}
				break
			}
		}
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
		s.GetNameForResource(utils.ReadOnlyHeadlessSVC), ordinal)
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

func (s *StatusSyncer) TryLeader(ctx context.Context, pod *corev1.Pod) error {

	// 1. close the xenon's SemiCheck.
	executor, err := internal.NewPodExecutor()
	if err != nil {
		return err
	}

	err = executor.XenonTryLeader(s.Namespace, pod.Name)
	s.log.Info("the xenon's tryleader", "pod", pod.Name)
	delete(pod.ObjectMeta.Labels, utils.LabelTryLeader)
	if err != nil {
		return err
	}

	if err := s.cli.Update(ctx, pod); err != nil {
		return err
	}
	return nil
}

func (s *StatusSyncer) SetLeaderReadOnly(pod *corev1.Pod) error {
	var sqlRunner internal.SQLRunner
	closeCh := make(chan func())

	var closeConn func()
	errCh := make(chan error)
	host := fmt.Sprintf("%s.%s-mysql.%s", pod.Name, s.Name, s.Namespace)
	cfg, errOut := internal.NewConfigFromClusterKey(
		s.cli, s.MysqlCluster.GetClusterKey(), utils.RootUser, host)
	go func(sqlRunner *internal.SQLRunner, errCh chan error, closeCh chan func()) {
		var err error
		*sqlRunner, closeConn, err = s.SQLRunnerFactory(cfg, errOut)
		if err != nil {
			s.log.V(1).Info("failed to get sql runner", "error", err)
			errCh <- err
			return
		}
		if closeConn != nil {
			closeCh <- closeConn
			return
		}
		errCh <- nil
	}(&sqlRunner, errCh, closeCh)

	select {
	case errOut = <-errCh:
		return errOut
	case closeConn := <-closeCh:
		defer closeConn()
	case <-time.After(time.Second * 5):
	}
	if sqlRunner != nil {
		return sqlRunner.QueryExec(internal.NewQuery("SET GLOBAL super_read_only=on"))
	}
	return nil
}

func (s *StatusSyncer) clusterSlaveCheck() error {
	if s.Spec.RemoteCluster == nil {
		return nil
	}
	//check remote cluster slave status
	//1. connect the leader pod.
	host := fmt.Sprintf("%s.%s", s.GetNameForResource(utils.LeaderService), s.Namespace)
	cfg, err := internal.NewConfigFromClusterKey(
		s.cli, s.MysqlCluster.GetClusterKey(), utils.RootUser, host)
	if err != nil {
		return err
	}
	sqlRunner, closeConn, err := s.SQLRunnerFactory(cfg)
	if err != nil {
		return err
	}
	defer closeConn()
	//if sqlRunner.
	if sqlRunner != nil {
		// check remote cluster slave status.
		s.log.Info("check remote cluster slave status", "namespace", s.Namespace)
		var isReplicating corev1.ConditionStatus
		var err error
		if _, isReplicating, err = internal.CheckSlaveStatus(sqlRunner, s.Spec.ReplicaLag); err != nil {
			//Notice!!! this has error, just show error message, can not return.
			s.log.V(1).Info("slave status has gotten error", "error", err)
		}
		clusterCondition := apiv1alpha1.ClusterCondition{
			Type:               apiv1alpha1.ConditionRemoteSlave,
			Status:             corev1.ConditionFalse,
			Message:            "the leader is not slave",
			Reason:             "need change master",
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
		if isReplicating == corev1.ConditionFalse {
			s.log.Info("remote cluster has no slave, change the status")
		} else {
			clusterCondition.Status = corev1.ConditionTrue
			clusterCondition.Message = ""
			clusterCondition.Reason = ""
		}
		s.Status.Conditions = append(s.Status.Conditions, clusterCondition)
	}
	return nil
}

func (s *StatusSyncer) updateClusterCondition(idx int, status corev1.ConditionStatus) {
	if s.Status.Conditions[idx].Status != status {
		t := time.Now()
		s.log.V(3).Info(fmt.Sprintf("Found status change for cluster %q condition %q: %q -> %q; setting lastTransitionTime to %v",
			s.Name, s.Status.Conditions[idx].Type, s.Status.Conditions[idx].Status, status, t))
		s.Status.Conditions[idx].Status = status
		s.Status.Conditions[idx].LastTransitionTime = metav1.NewTime(t)
	}
}
