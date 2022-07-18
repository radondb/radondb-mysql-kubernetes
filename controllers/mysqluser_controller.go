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
	"time"

	"github.com/go-test/deep"
	"github.com/presslabs/controller-util/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/internal"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	mysqluser "github.com/radondb/radondb-mysql-kubernetes/mysqluser"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// MysqlUserReconciler reconciles a MysqlUser object.
type MysqlUserReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	// MySQL query runner.
	internal.SQLRunnerFactory
}

var (
	userLog       = log.Log.WithName("controller").WithName("mysqluser")
	userFinalizer = "mysqluser-finalizer"
)

//+kubebuilder:rbac:groups=mysql.radondb.com,resources=mysqlusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.radondb.com,resources=mysqlusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.radondb.com,resources=mysqlusers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the MysqlUser object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the MysqlUser.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *MysqlUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// your logic here.
	user := mysqluser.New(&apiv1alpha1.MysqlUser{})

	err := r.Get(ctx, req.NamespacedName, user.Unwrap())
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			userLog.Info("mysql user not found, maybe deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	oldStatus := user.Status.DeepCopy()

	// If mysql user has been deleted then delete it from mysql cluster.
	if !user.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.removeUser(ctx, user)
	}

	// Write the desired status into mysql cluster.
	ruErr := r.reconcileUserInCluster(ctx, user)
	if err := r.updateStatusAndErr(ctx, user, oldStatus, ruErr); err != nil {
		return ctrl.Result{}, err
	}

	// Enqueue the resource again after to keep the resource up to date in mysql
	// in case is changed directly into mysql.
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: 2 * time.Minute,
	}, nil
}

// removeUser deletes the corresponding user in mysql before mysql user cr is deleted.
func (r *MysqlUserReconciler) removeUser(ctx context.Context, mysqlUser *mysqluser.MysqlUser) error {
	// The resource has been deleted.
	if meta.HasFinalizer(&mysqlUser.ObjectMeta, userFinalizer) {
		// Drop the user if the finalizer is still present.
		if err := r.dropUserFromDB(ctx, mysqlUser); err != nil {
			return err
		}

		meta.RemoveFinalizer(&mysqlUser.ObjectMeta, userFinalizer)

		// Update resource so it will remove the finalizer.
		if err := r.Update(ctx, mysqlUser.Unwrap()); err != nil {
			return err
		}
	}
	return nil
}

// reconcileUserInCluster reconcileUserInCluster creates or updates users in mysql.
// Proceed as follows:
// 1. Create users and authorize according to the Spec.
// 2. Remove the host that does not exist in the spec from MySQL.
// 3. Make sure mysqluser has finalizer set.
// 4. Update status and condition.
func (r *MysqlUserReconciler) reconcileUserInCluster(ctx context.Context, mysqlUser *mysqluser.MysqlUser) (err error) {
	// Catch the error and set the failed status.
	defer setFailedStatus(&err, mysqlUser)

	// Reconcile the mysqlUser into mysql.
	if err = r.reconcileUserInDB(ctx, mysqlUser); err != nil {
		return
	}

	// Add finalizer if is not added on the resource.
	if !meta.HasFinalizer(&mysqlUser.ObjectMeta, userFinalizer) {
		meta.AddFinalizer(&mysqlUser.ObjectMeta, userFinalizer)
		if err = r.Update(ctx, mysqlUser.Unwrap()); err != nil {
			return
		}
	}

	// Update status for allowedHosts if needed, mark that status need to be updated.
	if !reflect.DeepEqual(mysqlUser.Status.AllowedHosts, mysqlUser.Spec.Hosts) {
		mysqlUser.Status.AllowedHosts = mysqlUser.Spec.Hosts
	}

	// Update the status according to the result.
	mysqlUser.UpdateStatusCondition(
		apiv1alpha1.MySQLUserReady, corev1.ConditionTrue,
		mysqluser.ProvisionSucceededReason, "The user provisioning has succeeded.",
	)

	return
}

// reconcileUserInDB creates and authorizes(If needed) users based on
// spec.Hosts, and then deletes users that do not exist in spec.Hosts.
func (r *MysqlUserReconciler) reconcileUserInDB(ctx context.Context, mysqlUser *mysqluser.MysqlUser) error {
	sqlRunner, closeConn, err := r.SQLRunnerFactory(internal.NewConfigFromClusterKey(
		r.Client, mysqlUser.GetClusterKey(), utils.RootUser, utils.LeaderHost))
	if err != nil {
		return err
	}
	defer closeConn()

	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{Name: mysqlUser.Spec.SecretSelector.SecretName, Namespace: mysqlUser.Namespace}

	if err := r.Get(ctx, secretKey, secret); err != nil {
		return err
	}

	password := string(secret.Data[mysqlUser.Spec.SecretSelector.SecretKey])
	if password == "" {
		return fmt.Errorf("the MySQL user's password must not be empty")
	}

	// Create/Update user in database.
	userLog.Info("creating mysql user", "key", mysqlUser.GetKey(), "username", mysqlUser.Spec.User, "cluster", mysqlUser.GetClusterKey())
	if err := internal.CreateUserIfNotExists(sqlRunner, mysqlUser.Unwrap(), password); err != nil {
		return err
	}

	// Remove allowed hosts for user.
	toRemove := utils.StringDiffIn(mysqlUser.Status.AllowedHosts, mysqlUser.Spec.Hosts)
	for _, host := range toRemove {
		if err := internal.DropUser(sqlRunner, mysqlUser.Spec.User, host); err != nil {
			return err
		}
	}

	return nil
}

func (r *MysqlUserReconciler) dropUserFromDB(ctx context.Context, mysqlUser *mysqluser.MysqlUser) error {
	sqlRunner, closeConn, err := r.SQLRunnerFactory(internal.NewConfigFromClusterKey(
		r.Client, mysqlUser.GetClusterKey(), utils.RootUser, utils.LeaderHost))
	if errors.IsNotFound(err) {
		// If the mysql cluster does not exists then we can safely assume that
		// the user is deleted so exist successfully.
		statusErr, ok := err.(*errors.StatusError)
		if ok && mysqlcluster.IsClusterKind(statusErr.Status().Details.Kind) {
			// It seems the cluster is not to be found, so we assume it has been deleted.
			return nil
		}
	}

	if err != nil {
		return err
	}
	defer closeConn()

	for _, host := range mysqlUser.Status.AllowedHosts {
		userLog.Info("removing user from mysql cluster", "key", mysqlUser.GetKey(), "username", mysqlUser.Spec.User, "cluster", mysqlUser.GetClusterKey())
		if err := internal.DropUser(sqlRunner, mysqlUser.Spec.User, host); err != nil {
			return err
		}
	}
	return nil
}

// updateStatusAndErr update the status and catch create/update error.
func (r *MysqlUserReconciler) updateStatusAndErr(ctx context.Context, mysqlUser *mysqluser.MysqlUser, oldStatus *apiv1alpha1.UserStatus, cuErr error) error {
	if !reflect.DeepEqual(oldStatus, &mysqlUser.Status) {
		userLog.Info("update mysql user status", "key", mysqlUser.GetKey(), "diff", deep.Equal(oldStatus, &mysqlUser.Status))
		if err := r.Status().Update(ctx, mysqlUser.Unwrap()); err != nil {
			if cuErr != nil {
				return fmt.Errorf("failed to update status: %s, previous error was: %s", err, cuErr)
			}
			return err
		}
	}

	return cuErr
}

func setFailedStatus(err *error, mysqlUser *mysqluser.MysqlUser) {
	if *err != nil {
		mysqlUser.UpdateStatusCondition(
			apiv1alpha1.MySQLUserReady, corev1.ConditionFalse,
			mysqluser.ProvisionFailedReason, fmt.Sprintf("The user provisioning has failed: %s", *err),
		)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.MysqlUser{}).
		Complete(r)
}
