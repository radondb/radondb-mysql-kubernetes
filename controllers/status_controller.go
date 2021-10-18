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

package controllers

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/presslabs/controller-util/syncer"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/internal"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	clustersyncer "github.com/radondb/radondb-mysql-kubernetes/mysqlcluster/syncer"
)

// reconcileTimePeriod represents the time in which a cluster should be reconciled
var reconcileTimePeriod = time.Second * 5

// StatusReconciler reconciles a Status object
type StatusReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	// Mysql query runner.
	internal.SQLRunnerFactory
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Status object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *StatusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("controllers").WithName("Status")

	// your logic here
	instance := mysqlcluster.New(&apiv1alpha1.MysqlCluster{})

	err := r.Get(ctx, req.NamespacedName, instance.Unwrap())
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			log.Info("instance not found, maybe removed")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	status := *instance.Status.DeepCopy()
	defer func() {
		if !reflect.DeepEqual(status, instance.Status) {
			sErr := r.Status().Update(ctx, instance.Unwrap())
			if sErr != nil {
				log.Error(sErr, "failed to update cluster status")
			}
		}
	}()

	statusSyncer := clustersyncer.NewStatusSyncer(instance, r.Client, r.SQLRunnerFactory)
	if err := syncer.Sync(ctx, statusSyncer, r.Recorder); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := log.Log.WithName("controllers").WithName("Status")
	clusters := &sync.Map{}
	events := make(chan event.GenericEvent, 1024)
	bld := ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.MysqlCluster{}).
		Watches(&source.Kind{Type: &apiv1alpha1.MysqlCluster{}}, &handler.Funcs{
			CreateFunc: func(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
				if evt.Object == nil {
					log.Error(nil, "CreateEvent received with no metadata", "CreateEvent", evt)
					return
				}

				log.V(1).Info("register cluster in clusters list", "obj", evt.Object)
				clusters.Store(getKey(evt.Object), event.GenericEvent{
					Object: evt.Object,
				})
			},
			DeleteFunc: func(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
				if evt.Object == nil {
					log.Error(nil, "DeleteEvent received with no metadata", "DeleteEvent", evt)
					return
				}

				log.V(1).Info("remove cluster from clusters list", "obj", evt.Object)
				clusters.Delete(getKey(evt.Object))
			},
		}).
		Watches(&source.Channel{Source: events}, &handler.EnqueueRequestForObject{})

	// create a runnable function that dispatches events to events channel
	// this runnableFunc is passed to the manager that starts it.
	var f manager.RunnableFunc = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(reconcileTimePeriod):
				// write all clusters to events chan to be processed
				clusters.Range(func(key, value interface{}) bool {
					events <- value.(event.GenericEvent)
					log.V(1).Info("Schedule new cluster for reconciliation", "key", key)
					return true
				})
			}
		}
	}

	mgr.Add(f)
	return bld.Complete(r)
}

// getKey returns a string that represents the key under which cluster is registered
func getKey(obj klog.KMetadata) string {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}.String()
}
