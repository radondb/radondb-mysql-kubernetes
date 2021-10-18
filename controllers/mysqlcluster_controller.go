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

	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/internal"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	clustersyncer "github.com/radondb/radondb-mysql-kubernetes/mysqlcluster/syncer"
)

// MysqlClusterReconciler reconciles a MysqlCluster object
type MysqlClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	// Mysql query runner.
	internal.SQLRunnerFactory
}

// +kubebuilder:rbac:groups=mysql.radondb.com,resources=mysqlclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql.radondb.com,resources=mysqlclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mysql.radondb.com,resources=mysqlclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps;secrets;services;pods;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;create;patch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MysqlCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *MysqlClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("controllers").WithName("MysqlCluster")
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

	status := *instance.Status.DeepCopy()
	defer func() {
		if !reflect.DeepEqual(status, instance.Status) {
			sErr := r.Status().Update(ctx, instance.Unwrap())
			if sErr != nil {
				log.Error(sErr, "failed to update cluster status")
			}
		}
	}()

	configMapSyncer := clustersyncer.NewConfigMapSyncer(r.Client, instance)
	if err = syncer.Sync(ctx, configMapSyncer, r.Recorder); err != nil {
		return ctrl.Result{}, err
	}

	secretSyncer := clustersyncer.NewSecretSyncer(r.Client, instance)
	if err = syncer.Sync(ctx, secretSyncer, r.Recorder); err != nil {
		return ctrl.Result{}, err
	}

	cmRev := configMapSyncer.Object().(*corev1.ConfigMap).ResourceVersion
	sctRev := secretSyncer.Object().(*corev1.Secret).ResourceVersion

	// run the syncers for services, pdb and statefulset
	syncers := []syncer.Interface{
		clustersyncer.NewRoleSyncer(r.Client, instance),
		clustersyncer.NewRoleBindingSyncer(r.Client, instance),
		clustersyncer.NewServiceAccountSyncer(r.Client, instance),
		clustersyncer.NewHeadlessSVCSyncer(r.Client, instance),
		clustersyncer.NewLeaderSVCSyncer(r.Client, instance),
		clustersyncer.NewFollowerSVCSyncer(r.Client, instance),
		clustersyncer.NewStatefulSetSyncer(r.Client, instance, cmRev, sctRev, r.SQLRunnerFactory),
		clustersyncer.NewPDBSyncer(r.Client, instance),
	}

	if instance.Spec.MetricsOpts.Enabled {
		syncers = append(syncers, clustersyncer.NewMetricsSVCSyncer(r.Client, instance))
	}

	// run the syncers
	for _, sync := range syncers {
		if err = syncer.Sync(ctx, sync, r.Recorder); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.MysqlCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.Secret{}).
		Owns(&policyv1beta1.PodDisruptionBudget{}).
		Complete(r)
}
