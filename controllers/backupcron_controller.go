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
	"fmt"
	"reflect"
	"sync"

	"github.com/go-logr/logr"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/wgliang/cron"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/backup"
)

// BackupCronReconciler reconciles a BackupCron object
type BackupCronReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	Cron            *cron.Cron
	LockJobRegister *sync.Mutex
}

type startStopCron struct {
	Cron *cron.Cron
}

func (c startStopCron) Start(ctx context.Context) error {
	c.Cron.Start()
	<-ctx.Done()
	c.Cron.Stop()

	return nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BackupCron object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *BackupCronReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("controllers").WithName("backupCronJob")

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

	if err = instance.Validate(); err != nil {
		return ctrl.Result{}, err
	}
	// if spec.backupScheduler is not set then don't do anything
	if len(instance.Spec.BackupSchedule) == 0 {
		return reconcile.Result{}, nil
	}

	schedule, err := cron.Parse(instance.Spec.BackupSchedule)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to parse schedule: %s", err)
	}

	log.V(1).Info("register cluster in cronjob", "key", instance, "schedule", schedule)

	return ctrl.Result{}, r.updateClusterSchedule(ctx, instance.Unwrap(), schedule, log)
}

// updateClusterSchedule creates/updates a cron job for specified cluster.
func (r *BackupCronReconciler) updateClusterSchedule(ctx context.Context, cluster *apiv1alpha1.MysqlCluster, schedule cron.Schedule, log logr.Logger) error {

	r.LockJobRegister.Lock()
	defer r.LockJobRegister.Unlock()

	for _, entry := range r.Cron.Entries() {
		j, ok := entry.Job.(*backup.CronJob)
		if ok && j.ClusterName == cluster.Name && j.Namespace == cluster.Namespace {
			log.V(1).Info("cluster already added to cron.", "key", cluster)

			// change scheduler for already added crons
			if !reflect.DeepEqual(entry.Schedule, schedule) {
				log.Info("update cluster scheduler", "key", cluster,
					"scheduler", cluster.Spec.BackupSchedule)

				if err := r.Cron.Remove(cluster.Name); err != nil {
					return err
				}
				break
			}
			if j.Image != cluster.Spec.PodPolicy.SidecarImage {
				log.Info("update cluster image", "key", cluster, "image", cluster.Spec.PodPolicy.SidecarImage)
				j.Image = cluster.Spec.PodPolicy.SidecarImage
			}
			return nil
		}
	}

	r.Cron.Schedule(schedule, &backup.CronJob{
		ClusterName:                    cluster.Name,
		Namespace:                      cluster.Namespace,
		Client:                         r.Client,
		Image:                          cluster.Spec.PodPolicy.SidecarImage,
		BackupScheduleJobsHistoryLimit: cluster.Spec.BackupScheduleJobsHistoryLimit,
		//BackupRemoteDeletePolicy:       cluster.Spec.BackupRemoteDeletePolicy,

		NFSServerAddress: cluster.Spec.NFSServerAddress,
		Log:              log,
	}, cluster.Name)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackupCronReconciler) SetupWithManager(mgr ctrl.Manager) error {
	sscron := startStopCron{
		Cron: r.Cron,
	}
	mgr.Add(sscron)
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&apiv1alpha1.MysqlCluster{}).
		Complete(r)
}
