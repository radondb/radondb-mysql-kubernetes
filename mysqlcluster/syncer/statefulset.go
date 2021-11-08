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
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/imdario/mergo"
	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/internal"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster/container"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// The wait time limit for pod upgrade.
const waitLimit = 2 * 60 * 60

// StatefulSetSyncer used to operate statefulset.
type StatefulSetSyncer struct {
	*mysqlcluster.MysqlCluster

	cli client.Client

	sfs *appsv1.StatefulSet

	// Configmap resourceVersion.
	cmRev string

	// Secret resourceVersion.
	sctRev string

	// Mysql query runner.
	internal.SQLRunnerFactory
	// ordinal
	ordinal int
}

// NewStatefulSetSyncer returns a pointer to StatefulSetSyncer.
func NewStatefulSetSyncer(cli client.Client, c *mysqlcluster.MysqlCluster,
	cmRev, sctRev string,
	ordinal int,
	sqlRunnerFactory internal.SQLRunnerFactory) *StatefulSetSyncer {
	//c.Ordinal = ordinal
	return &StatefulSetSyncer{
		MysqlCluster: c,
		cli:          cli,
		ordinal:      ordinal,
		sfs: &appsv1.StatefulSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "StatefulSet",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.GetNameForResource(utils.StatefulSet) + fmt.Sprintf("-%d", ordinal),
				Namespace: c.Namespace,
			},
		},
		cmRev:            cmRev,
		sctRev:           sctRev,
		SQLRunnerFactory: sqlRunnerFactory,
	}
}

// Object returns the object for which sync applies.
func (s *StatefulSetSyncer) Object() interface{} { return s.sfs }

// GetObject returns the object for which sync applies
func (s *StatefulSetSyncer) GetObject() interface{} { return s.sfs }

// Owner returns the object owner or nil if object does not have one.
func (s *StatefulSetSyncer) ObjectOwner() runtime.Object { return s.Unwrap() }

// GetOwner returns the object owner or nil if object does not have one.
func (s *StatefulSetSyncer) GetOwner() runtime.Object { return s.Unwrap() }

// Sync persists data into the external store.
// It's called by mysqlcluster controller, when return error, retry Reconcile(),when return nil, exit this cycle.
// See https://github.com/presslabs/controller-util/blob/master/syncer/object.go#L68
func (s *StatefulSetSyncer) Sync(ctx context.Context) (syncer.SyncResult, error) {
	var err error
	var kind string
	result := syncer.SyncResult{}

	result.Operation, err = s.createOrUpdate(ctx)

	// Get namespace and name.
	key := client.ObjectKeyFromObject(s.sfs)
	// Get groupVersionKind.
	gvk, gvkErr := apiutil.GVKForObject(s.sfs, s.cli.Scheme())
	if gvkErr != nil {
		kind = fmt.Sprintf("%T", s.sfs)
	} else {
		kind = gvk.String()
	}
	// Print log.
	// Info: owner is deleted or ignored error.
	// Warning: other errors.
	// Normal: no error.
	switch {
	case errors.Is(err, syncer.ErrOwnerDeleted):
		log.Info(string(result.Operation), "key", key, "kind", kind, "error", err)
		err = nil
	case errors.Is(err, syncer.ErrIgnore):
		log.Info("syncer skipped", "key", key, "kind", kind, "error", err)
		err = nil
	case err != nil:
		// When Invliad type error occur, and the PVC claims has changed
		// do the expand PVCs.
		if k8serrors.IsInvalid(err) && s.canExpandPVC(ctx) {
			result.Operation, err = s.expandPVCs(ctx)
		} else {
			result.SetEventData("Warning", basicEventReason(s.Name, err),
				fmt.Sprintf("%s %s failed syncing: %s", kind, key, err))
			log.Error(err, string(result.Operation), "key", key, "kind", kind)
		}

	default:
		result.SetEventData("Normal", basicEventReason(s.Name, err),
			fmt.Sprintf("%s %s %s successfully", kind, key, result.Operation))
		log.Info(string(result.Operation), "key", key, "kind", kind)
	}
	return result, err
}

// Check whether need to expand PVC.
func (s *StatefulSetSyncer) canExpandPVC(ctx context.Context) bool {
	// Get it again. Becaus it has been mutated in Sync.
	if err := s.cli.Get(ctx, client.ObjectKeyFromObject(s.sfs), s.sfs); err != nil {
		if k8serrors.IsNotFound(err) {
			return false
		}
	}
	oldRequest := s.sfs.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.DeepCopy()
	// Here do s.mutate again.
	if err := s.mutate(); err != nil {
		return false
	}
	newStorage := s.sfs.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage()
	// If newStorage is not greater than oldStorage, do not expand.
	if newStorage.Cmp(*oldRequest.Storage()) != 1 {
		log.Info("canExpandPVC", "result", "can not expand", "reason", "new pvc is not larger than old pvc")
		return false
	}
	return true
}

// expandPVCs by reCreate the statefulset and Expand pvcs.
func (s *StatefulSetSyncer) expandPVCs(ctx context.Context) (controllerutil.OperationResult, error) {
	// At first, delete the statefulset,for expand PVC.
	if err := s.cli.Delete(ctx, s.sfs); err != nil {
		return controllerutil.OperationResultNone, err
	}
	// Do expend the PVCs.
	if err := s.doExpandPVCs(ctx); err != nil {
		log.Error(err, "expandPVCs")
		return controllerutil.OperationResultNone, err
	}
	// Then Create sfs again.
	if err := s.cli.Create(ctx, s.sfs); err != nil {
		return controllerutil.OperationResultNone, err
	} else {
		return controllerutil.OperationResultCreated, nil
	}
}

// doExpandPVCs is used to extend PVC's size by refreshing PVC Resources.Requests in Spec.
func (s *StatefulSetSyncer) doExpandPVCs(ctx context.Context) error {
	pvcs := corev1.PersistentVolumeClaimList{}
	lables := s.GetLabels()
	s.AppendOrdinal(lables, s.ordinal)
	if err := s.cli.List(ctx,
		&pvcs,
		&client.ListOptions{
			Namespace:     s.sfs.Namespace,
			LabelSelector: lables.AsSelector(),
		},
	); err != nil {
		return err
	}

	for _, item := range pvcs.Items {
		// Notice: Only Resources.Requests can update, other field update will failure.
		// If storage Class's allowVolumeExpansion is false, update will failure.
		item.Spec.Resources.Requests = s.sfs.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests
		if err := s.cli.Update(ctx, &item); err != nil {
			return err
		}
		if err := retry(time.Second*2, time.Duration(waitLimit)*time.Second, func() (bool, error) {
			// Check the pvc status.
			var currentPVC corev1.PersistentVolumeClaim
			if err2 := s.cli.Get(ctx, client.ObjectKeyFromObject(&item), &currentPVC); err2 != nil {
				return true, err2
			}
			var conditons = currentPVC.Status.Conditions
			// Notice: When expanding not start, or been completed, conditons is nil
			if conditons == nil {
				// If change storage request when replicas are creating, should check the currentPVC.Status.Capacity.
				// for example:
				// Pod0 has created successful,but Pod1 is creating. then change PVC from 20Gi to 30Gi .
				// Pod0's PVC need to expand, but Pod1's PVC has created as 30Gi, so need to skip it.
				if equality.Semantic.DeepEqual(currentPVC.Status.Capacity, item.Spec.Resources.Requests) {
					return true, nil
				}
				return false, nil
			}
			status := conditons[0].Type
			if status == "FileSystemResizePending" {
				return true, nil
			}
			return false, nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// createOrUpdate creates or updates the statefulset in the Kubernetes cluster.
// See https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/controller/controllerutil?utm_source=gopls#CreateOrUpdate
func (s *StatefulSetSyncer) createOrUpdate(ctx context.Context) (controllerutil.OperationResult, error) {
	var err error
	// Check if statefulset exists
	if err = s.cli.Get(ctx, client.ObjectKeyFromObject(s.sfs), s.sfs); err != nil {
		if !k8serrors.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}

		if err = s.mutate(); err != nil {
			return controllerutil.OperationResultNone, err
		}

		if err = s.cli.Create(ctx, s.sfs); err != nil {
			return controllerutil.OperationResultNone, err
		} else {
			return controllerutil.OperationResultCreated, nil
		}
	}
	// Deep copy the old statefulset from StatefulSetSyncer.
	existing := s.sfs.DeepCopyObject()
	// Sync data from mysqlcluster.spec to statefulset.
	if err = s.mutate(); err != nil {
		return controllerutil.OperationResultNone, err
	}
	// Check if statefulset changed.
	if equality.Semantic.DeepEqual(existing, s.sfs) {
		return controllerutil.OperationResultNone, nil
	}
	// If changed, update statefulset.
	if err := s.cli.Update(ctx, s.sfs); err != nil {
		return controllerutil.OperationResultNone, err
	}
	// Update every pods of statefulset.
	// if err := s.updatePod(ctx); err != nil {
	// 	return controllerutil.OperationResultNone, err
	// }
	// // Update pvc.
	// if err := s.updatePVC(ctx); err != nil {
	// 	return controllerutil.OperationResultNone, err
	// }

	return controllerutil.OperationResultUpdated, nil
}

// updatePod update the pods, update follower nodes first.
// This can reduce the number of master-slave switching during the update process.
func (s *StatefulSetSyncer) updatePod(ctx context.Context) error {
	if s.sfs.Status.UpdatedReplicas >= s.sfs.Status.Replicas {
		return nil
	}

	log.Info("statefulSet was changed, run update")

	if s.sfs.Status.ReadyReplicas < s.sfs.Status.Replicas {
		log.Info("can't start/continue 'update': waiting for all replicas are ready")
		return nil
	}

	if backuping, _ := s.backupIsRunning(ctx); backuping {
		// return error, it will reconsile again
		return fmt.Errorf("can't start/continue 'update': waiting for all backup completed")
	}

	// Get all pods.
	pods := corev1.PodList{}
	if err := s.cli.List(ctx,
		&pods,
		&client.ListOptions{
			Namespace:     s.sfs.Namespace,
			LabelSelector: s.GetLabels().AsSelector(),
		},
	); err != nil {
		return err
	}
	var leaderPod corev1.Pod
	var followerPods []corev1.Pod
	for _, pod := range pods.Items {
		// Check if the pod is healthy.
		if pod.ObjectMeta.Labels["healthy"] != "yes" {
			return fmt.Errorf("can't start/continue 'update': pod[%s] is unhealthy", pod.Name)
		}
		// Skip if pod is leader.
		if pod.ObjectMeta.Labels["role"] == "leader" && leaderPod.Name == "" {
			leaderPod = pod
			continue
		}
		followerPods = append(followerPods, pod)
		// If pod is not leader, direct update.
		if err := s.applyNWait(ctx, &pod); err != nil {
			return err
		}
	}
	// All followers have been updated now, then update leader.
	if leaderPod.Name != "" {
		// When replicas is two (one leader and one follower).
		if *s.sfs.Spec.Replicas == 2 {
			if err := s.preUpdate(ctx, leaderPod.Name, followerPods[0].Name); err != nil {
				return err
			}
		}
		// Update the leader.
		if err := s.applyNWait(ctx, &leaderPod); err != nil {
			return err
		}
	}

	return nil
}

// preUpdate run before update the leader pod when replicas is 2.
// Its main function is manually switch the leader node.
// 1. Get secrets (operator-user, operator-password, root-password).
// 2. Connect leader mysql.
// 3. Set leader read only.
// 4. Make sure the leader has sent all binlog to follower.
// 5. Check followerHost current role.
// 6. If followerHost is not leader, switch it to leader through xenon.
func (s *StatefulSetSyncer) preUpdate(ctx context.Context, leader, follower string) error {
	leaderRunner, closeConn, err := s.SQLRunnerFactory(internal.NewConfigFromClusterKey(
		s.cli, s.MysqlCluster.GetClusterKey(), utils.OperatorUser, utils.LeaderHost))
	if err != nil {
		return err
	}
	defer closeConn()

	// Status.Replicas indicate the number of Pod has been created.
	// So sfs.Spec.Replicas is 2, May be sfs.Status.Replicas maybe are 3, 5 ,
	// because it do not update the pods, so it is still the last status.
	if *s.sfs.Spec.Replicas != 2 {
		return nil
	}

	// Touch a new preUpdate file ,indicate that preUpdate is going on
	// remove it when it is finished.
	// See https://github.com/radondb/radondb-mysql-kubernetes/issues/178
	utils.TouchUpdateFile()
	defer utils.RemoveUpdateFile()
	sctName := s.GetNameForResource(utils.Secret)
	svcName := s.GetNameForResource(utils.HeadlessSVC)
	nameSpace := s.Namespace

	// Get secrets.
	secret := &corev1.Secret{}
	if err := s.cli.Get(context.TODO(),
		types.NamespacedName{
			Namespace: nameSpace,
			Name:      sctName,
		},
		secret,
	); err != nil {
		return fmt.Errorf("failed to get the secret: %s", sctName)
	}

	rootPasswd, ok := secret.Data["root-password"]
	if !ok {
		return fmt.Errorf("failed to get the root password: %s", rootPasswd)
	}

	if err = retry(time.Second*2, time.Duration(waitLimit)*time.Second, func() (bool, error) {
		// Set leader read only.
		if err = leaderRunner.QueryExec(internal.NewQuery("SET GLOBAL super_read_only=on;")); err != nil {
			log.Error(err, "failed to set leader read only", "node", leader)
			return false, err
		}

		// Make sure the master has sent all binlog to slave.
		success, err := internal.CheckProcesslist(leaderRunner)
		if err != nil {
			return false, err
		}
		if success {
			return true, nil
		}
		return false, nil
	}); err != nil {
		return err
	}

	followerHost := fmt.Sprintf("%s.%s.%s", follower, svcName, nameSpace)
	if err = retry(time.Second*5, time.Second*60, func() (bool, error) {
		// Check whether is leader.
		status, err := checkRole(followerHost, rootPasswd)
		if err != nil {
			log.Error(err, "failed to check role", "pod", follower)
			return false, nil
		}
		if status == corev1.ConditionTrue {
			return true, nil
		}

		// If not leader, try to leader.
		xenonHttpRequest(followerHost, "POST", "/v1/raft/trytoleader", rootPasswd, nil)
		return false, nil
	}); err != nil {
		return err
	}

	return nil
}

// mutate set the statefulset.
func (s *StatefulSetSyncer) mutate() error {
	var replica int32 = 1
	s.sfs.Spec.ServiceName = s.GetNameForResource(utils.StatefulSet) + fmt.Sprintf("-%d", s.ordinal)
	s.sfs.Spec.Replicas = &replica
	s.sfs.Spec.Selector = metav1.SetAsLabelSelector(s.GetSelectorLabels())
	s.sfs.Spec.UpdateStrategy = appsv1.StatefulSetUpdateStrategy{
		Type: appsv1.OnDeleteStatefulSetStrategyType,
	}
	lables := s.GetLabels()
	lables = s.AppendOrdinal(lables, s.ordinal)
	s.sfs.Spec.Template.ObjectMeta.Labels = lables
	for k, v := range s.Spec.PodPolicy.Labels {
		s.sfs.Spec.Template.ObjectMeta.Labels[k] = v
	}
	s.sfs.Spec.Template.ObjectMeta.Labels["role"] = "candidate"
	s.sfs.Spec.Template.ObjectMeta.Labels["healthy"] = "no"

	s.sfs.Spec.Template.Annotations = s.Spec.PodPolicy.Annotations
	if len(s.sfs.Spec.Template.ObjectMeta.Annotations) == 0 {
		s.sfs.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	if s.Spec.MetricsOpts.Enabled {
		s.sfs.Spec.Template.ObjectMeta.Annotations["prometheus.io/scrape"] = "true"
		s.sfs.Spec.Template.ObjectMeta.Annotations["prometheus.io/port"] = fmt.Sprintf("%d", utils.MetricsPort)
	}
	s.sfs.Spec.Template.ObjectMeta.Annotations["config_rev"] = s.cmRev
	s.sfs.Spec.Template.ObjectMeta.Annotations["secret_rev"] = s.sctRev

	err := mergo.Merge(&s.sfs.Spec.Template.Spec, s.ensurePodSpec(s.ordinal), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}
	s.sfs.Spec.Template.Spec.Tolerations = s.Spec.PodPolicy.Tolerations

	if s.Spec.Persistence.Enabled {
		if s.sfs.Spec.VolumeClaimTemplates, err = s.EnsureVolumeClaimTemplates(s.cli.Scheme(), s.ordinal); err != nil {
			return err
		}
	}

	// Set owner reference only if owner resource is not being deleted, otherwise the owner
	// reference will be reset in case of deleting with cascade=false.
	if s.Unwrap().GetDeletionTimestamp().IsZero() {
		if err := controllerutil.SetControllerReference(s.Unwrap(), s.sfs, s.cli.Scheme()); err != nil {
			return err
		}
	} else if ctime := s.Unwrap().GetCreationTimestamp(); ctime.IsZero() {
		// The owner is deleted, don't recreate the resource if does not exist, because gc
		// will not delete it again because has no owner reference set.
		return fmt.Errorf("owner is deleted")
	}
	return nil
}

// ensurePodSpec used to ensure the podspec.
func (s *StatefulSetSyncer) ensurePodSpec(ordinal int) corev1.PodSpec {
	initSidecar := container.EnsureContainer(utils.ContainerInitSidecarName, s.MysqlCluster, ordinal)
	initMysql := container.EnsureContainer(utils.ContainerInitMysqlName, s.MysqlCluster, ordinal)
	initContainers := []corev1.Container{initSidecar, initMysql}

	mysql := container.EnsureContainer(utils.ContainerMysqlName, s.MysqlCluster, ordinal)
	xenon := container.EnsureContainer(utils.ContainerXenonName, s.MysqlCluster, ordinal)
	backup := container.EnsureContainer(utils.ContainerBackupName, s.MysqlCluster, ordinal)
	containers := []corev1.Container{mysql, xenon, backup}
	if s.Spec.MetricsOpts.Enabled {
		containers = append(containers, container.EnsureContainer(utils.ContainerMetricsName, s.MysqlCluster, ordinal))
	}
	if s.Spec.PodPolicy.SlowLogTail {
		containers = append(containers, container.EnsureContainer(utils.ContainerSlowLogName, s.MysqlCluster, ordinal))
	}
	if s.Spec.PodPolicy.AuditLogTail {
		containers = append(containers, container.EnsureContainer(utils.ContainerAuditLogName, s.MysqlCluster, ordinal))
	}

	return corev1.PodSpec{
		InitContainers:     initContainers,
		Containers:         containers,
		Volumes:            s.EnsureVolumes(),
		SchedulerName:      s.Spec.PodPolicy.SchedulerName,
		ServiceAccountName: s.GetNameForResource(utils.ServiceAccount),
		Affinity:           s.Spec.PodPolicy.Affinity,
		PriorityClassName:  s.Spec.PodPolicy.PriorityClassName,
		Tolerations:        s.Spec.PodPolicy.Tolerations,
	}
}

// updatePVC used to update the pvc, check and remove the extra pvc.
func (s *StatefulSetSyncer) updatePVC(ctx context.Context) error {
	pvcs := corev1.PersistentVolumeClaimList{}
	if err := s.cli.List(ctx,
		&pvcs,
		&client.ListOptions{
			Namespace:     s.sfs.Namespace,
			LabelSelector: s.GetLabels().AsSelector(),
		},
	); err != nil {
		return err
	}

	for _, item := range pvcs.Items {
		if item.DeletionTimestamp != nil {
			log.Info("pvc is being deleted", "pvc", item.Name, "key", s.Unwrap())
			continue
		}

		ordinal, err := utils.GetOrdinal(item.Name)
		if err != nil {
			log.Error(err, "pvc deletion error", "key", s.Unwrap())
			continue
		}

		if ordinal >= int(*s.Spec.Replicas) {
			log.Info("cleaning up pvc", "pvc", item.Name, "key", s.Unwrap())
			if err := s.cli.Delete(ctx, &item); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *StatefulSetSyncer) applyNWait(ctx context.Context, pod *corev1.Pod) error {
	// Check version, if not latest, delete node.
	if pod.ObjectMeta.Labels["controller-revision-hash"] == s.sfs.Status.UpdateRevision {
		log.Info("pod is already updated", "pod name", pod.Name)
	} else {
		s.Status.State = apiv1alpha1.ClusterUpdateState
		log.Info("updating pod", "pod", pod.Name, "key", s.Unwrap())
		if pod.DeletionTimestamp != nil {
			log.Info("pod is being deleted", "pod", pod.Name, "key", s.Unwrap())
		} else {
			if err := s.cli.Delete(ctx, pod); err != nil {
				return err
			}
		}
	}

	// Wait the pod restart and healthy.
	return retry(time.Second*10, time.Duration(waitLimit)*time.Second, func() (bool, error) {
		err := s.cli.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pod)
		if err != nil && !k8serrors.IsNotFound(err) {
			return false, err
		}

		ordinal, err := utils.GetOrdinal(pod.Name)
		if err != nil {
			return false, err
		}
		if ordinal >= int(*s.Spec.Replicas) {
			log.Info("replicas were changed,  should skip", "pod", pod.Name)
			return true, nil
		}

		if pod.Status.Phase == corev1.PodFailed {
			return false, fmt.Errorf("pod %s is in failed phase", pod.Name)
		}

		if pod.ObjectMeta.Labels["healthy"] == "yes" &&
			pod.ObjectMeta.Labels["controller-revision-hash"] == s.sfs.Status.UpdateRevision {
			return true, nil
		}

		// fix issue#219. When 2->5 rolling update, Because of PDB, minAvaliable 50%, if Spec Replicas is 5, sfs Spec first be set to 3, then to be set 5
		// pod healthy is yes,but controller-revision-hash will never correct, it must return,otherwise wait for 2 hours.
		// https://kubernetes.io/zh/docs/tasks/run-application/configure-pdb/
		if pod.ObjectMeta.Labels["healthy"] == "yes" &&
			pod.ObjectMeta.Labels["controller-revision-hash"] != s.sfs.Status.UpdateRevision {
			return false, fmt.Errorf("pod %s is ready, wait next schedule", pod.Name)
		}
		return false, nil
	})
}

// retry runs func "f" every "in" time until "limit" is reached.
// it also doesn't have an extra tail wait after the limit is reached
// and f func runs first time instantly
func retry(in, limit time.Duration, f func() (bool, error)) error {
	fdone, err := f()
	if err != nil {
		return err
	}
	if fdone {
		return nil
	}

	done := time.NewTimer(limit)
	defer done.Stop()
	tk := time.NewTicker(in)
	defer tk.Stop()

	for {
		select {
		case <-done.C:
			return fmt.Errorf("reach pod wait limit")
		case <-tk.C:
			fdone, err := f()
			if err != nil {
				return err
			}
			if fdone {
				return nil
			}
		}
	}
}

func basicEventReason(objKindName string, err error) string {
	if err != nil {
		return fmt.Sprintf("%sSyncFailed", strcase.ToCamel(objKindName))
	}

	return fmt.Sprintf("%sSyncSuccessfull", strcase.ToCamel(objKindName))
}

func xenonHttpRequest(host, method, url string, rootPasswd []byte, body io.Reader) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("http://%s:%d%s", host, utils.XenonPeerPort, url), body)
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(append([]byte("root:"), rootPasswd...))
	req.Header.Set("Authorization", "Basic "+encoded)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get raft status failed, status code is %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// check the backup is exist and running
func (s *StatefulSetSyncer) backupIsRunning(ctx context.Context) (bool, error) {
	backuplist := apiv1alpha1.BackupList{}
	if err := s.cli.List(ctx,
		&backuplist,
		&client.ListOptions{
			Namespace: s.sfs.Namespace,
		},
	); err != nil {
		return false, err
	}
	for _, bcp := range backuplist.Items {
		if bcp.ClusterName != s.ClusterName {
			continue
		}
		if !bcp.Status.Completed {
			return true, nil
		}
	}
	return false, nil
}
