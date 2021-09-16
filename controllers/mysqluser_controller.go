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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/internal"
	"github.com/radondb/radondb-mysql-kubernetes/mysqluser"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// MysqlUserReconciler reconciles a MysqlUser object.
type MysqlUserReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

var userLog = log.Log.WithName("controller").WithName("mysqluser")

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

	sqlRunner, err := r.getSqlRunner(ctx, user)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("get sqlrunner faild, err: %s", err)
	}

	defer sqlRunner.Close()

	// If mysql user has been deleted then delete it from mysql cluster.
	if !user.ObjectMeta.DeletionTimestamp.IsZero() {
		queries := internal.GetDeleteQuery(ctx, user.Status.AllowedHosts, user.Spec.User)
		return ctrl.Result{}, r.deleteAllUser(ctx, queries, sqlRunner, user)
	}

	// If mysql user present, create or update.
	cuErr := r.createOrUpdate(ctx, sqlRunner, user)
	if err := r.updateStatusAndErr(ctx, user, oldStatus, cuErr); err != nil {
		setFailedStatus(&err, user)
		return ctrl.Result{}, err
	}

	// enqueue the resource again after to keep the resource up to date in mysql
	// in case is changed directly into mysql.
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: 2 * time.Minute,
	}, nil
}

// createOrUpdate create or update users in mysql.
// 1. Check whether the necessary parameters exist and comply with the rules.
// 2. Check if the spec.hosts exists in mysql.
// 3. Set the password and permissions for the existing host according to the spec.
// 4. Delete the host that exists in the mysql but does not exist in the spec.
// 5. Create unexist hosts.
// 6. Add finalizer.
// 7. Authorize newly created user hosts.
// 8. Update user`s status.
func (r *MysqlUserReconciler) createOrUpdate(ctx context.Context, sqlRunner *internal.SQLRunner, mysqlUser *mysqluser.MysqlUser) error {
	// 1. Check whether the necessary parameters exist and comply with the rules.
	userName := mysqlUser.Spec.User
	if userName == "" {
		return fmt.Errorf("mysql user name must not be empty")
	}

	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{Name: mysqlUser.Spec.SecretBinder.SecretName, Namespace: mysqlUser.Namespace}
	// Get the secret bound by mysql user.
	if err := r.Get(ctx, secretKey, secret); err != nil {
		return err
	}

	password := ""
	userSecretKey := mysqlUser.Spec.SecretBinder.SecretKey
	// Check if the key of the map exists.
	if pwdFromSecret, ok := secret.Data[userSecretKey]; !ok {
		return fmt.Errorf("secret key %s is empty", userSecretKey)
	} else {
		password = string(pwdFromSecret)
	}

	// Check if the password is empty.
	if password == "" {
		return fmt.Errorf("password must not be empty")
	}

	// 2. Check if the spec.hosts exists in mysql.
	// If mysql user host not exist, add to unexistHosts, waiting to be created.
	// If mysql user host exist, update password and permissions.
	unexistHosts := []string{}
	permissions := mysqlUser.Spec.Permissions
	for _, host := range mysqlUser.Spec.Hosts {
		if err := checkUserExist(ctx, internal.CheckUserQuery(ctx, userName, host), sqlRunner); err != nil {
			userLog.Info("mysql user not exist, create new mysql user", "need create host", host)
			// unexistHosts save the hosts that need create.
			unexistHosts = append(unexistHosts, host)
		} else {
			// 3. Set the password and permissions for the existing hosts according to the spec.
			// Regardless of whether the password is modified or not, reset it to keep it up to date.
			if err := r.setPassword(ctx, sqlRunner, userName, host, password); err != nil {
				return fmt.Errorf("set password faild, err: %s", err)
			}

			host := []string{host}
			// Regardless of whether the permissions is modified or not, reset it to keep it up to date.
			if err := r.setPermission(ctx, sqlRunner, permissions, userName, host); err != nil {
				return fmt.Errorf("set permissions faild, err: %s", err)
			}
		}
	}

	// 4. Delete the host that exists in the mysql but does not exist in the spec.
	if err := r.updateAllowedHosts(ctx, sqlRunner, mysqlUser); err != nil {
		return fmt.Errorf("update allowed hosts faild, err: %s", err)
	}
	if len(unexistHosts) == 0 {
		userLog.Info("users already exist", "user name:", mysqlUser.Spec.User, "hosts:", mysqlUser.Status.AllowedHosts)
		return nil
	}

	// 5. Create unexist hosts.
	query := internal.GetCreateQuery(ctx, unexistHosts, userName, password)
	if err := r.createUser(ctx, query, sqlRunner); err != nil {
		return fmt.Errorf("create user faild, err: %s", err)
	}
	userLog.Info("create mysql user success", "user name", userName, "hosts", unexistHosts)

	// 6. Add finalizer.
	if err := r.addUserFinalizer(ctx, mysqlUser.Unwrap()); err != nil {
		return fmt.Errorf("add finalizer faild")
	}

	// 7. Authorize newly created user hosts.
	// Ensure that the user has been successfully created before authorization.
	if err := r.setPermission(ctx, sqlRunner, permissions, userName, unexistHosts); err != nil {
		return fmt.Errorf("set permissions faild, err: %s", err)
	}

	// 8. Update user`s status.
	if !reflect.DeepEqual(mysqlUser.Status.AllowedHosts, mysqlUser.Spec.Hosts) {
		mysqlUser.Status.AllowedHosts = mysqlUser.Spec.Hosts
	}
	mysqlUser.UpdateStatusCondition(
		apiv1alpha1.MySQLUserReady, corev1.ConditionTrue,
		mysqluser.ConfigurationSucceededReason, "The mysql user has been configured successfully.",
	)

	return nil
}

// createUser create users in mysql based on cr information.
func (r *MysqlUserReconciler) createUser(ctx context.Context, query internal.Query, sqlRunner *internal.SQLRunner) error {
	if err := sqlRunner.RunQuery(query.String(), query.Args()...); err != nil {
		return fmt.Errorf("create mysql user faild, err: %s", err)
	}
	return nil
}

// updateAllowedHosts delete user hosts that do not exist in spec.
func (r *MysqlUserReconciler) updateAllowedHosts(ctx context.Context, sqlRunner *internal.SQLRunner, mysqlUser *mysqluser.MysqlUser) error {
	toRemove := utils.StringDiffIn(mysqlUser.Status.AllowedHosts, mysqlUser.Spec.Hosts)
	if len(toRemove) == 0 {
		return nil
	}
	query := internal.GetDeleteQuery(ctx, toRemove, mysqlUser.Spec.User)

	return r.deleteUser(ctx, query, sqlRunner, mysqlUser)
}

// setPassword set password to existed user.
func (r *MysqlUserReconciler) setPassword(ctx context.Context, sqlRunner *internal.SQLRunner, userName, host, password string) error {
	query := internal.SetPasswordQuery(ctx, userName, host, password)
	if err := sqlRunner.RunQuery(query.String(), query.Args()...); err != nil {
		return err
	}

	return nil
}

// setPermission set permissions to existed user.
func (r *MysqlUserReconciler) setPermission(ctx context.Context, sqlRunner *internal.SQLRunner, permissions []apiv1alpha1.UserPermission, userName string, hosts []string) error {
	if len(permissions) > 0 {
		query := internal.GetGrantQuery(permissions, userName, hosts)
		if err := sqlRunner.RunQuery(query.String(), query.Args()...); err != nil {
			return err
		}
	}

	return nil
}

// deleteUser delete users in mysql but not delete finalizer, its called when update allowed hosts.
func (r *MysqlUserReconciler) deleteUser(ctx context.Context, query internal.Query, sqlRunner *internal.SQLRunner, mysqlUser *mysqluser.MysqlUser) error {
	if err := sqlRunner.RunQuery(query.String(), query.Args()...); err != nil {
		return fmt.Errorf("delete mysql user faild, err: %s", err)
	}

	userLog.Info("update mysql user hosts success.")

	return nil
}

// deleteAllUser delete all users in mysql that has been created, its called when delete the MysqlUser cr.
func (r *MysqlUserReconciler) deleteAllUser(ctx context.Context, query internal.Query, sqlRunner *internal.SQLRunner, mysqlUser *mysqluser.MysqlUser) error {
	crdName := mysqlUser.Name
	userName := mysqlUser.Spec.User
	if controllerutil.ContainsFinalizer(mysqlUser.Unwrap(), string(utils.UserFinalizer)) {
		if err := sqlRunner.RunQuery(query.String(), query.Args()...); err != nil {
			return fmt.Errorf("delete mysql user faild, err: %s", err)
		}

		if err := r.deleteUserFinalizer(ctx, mysqlUser.Unwrap()); err != nil {
			return fmt.Errorf("delete mysql user finalizer faild, err: %s", err)
		}

		userLog.Info("delete mysql user success.", "mysql user crd name", crdName, "mysql user name", userName)
	}

	return nil
}

// checkUserExist check if the user exist.
func checkUserExist(ctx context.Context, query internal.Query, sqlRunner *internal.SQLRunner) error {
	if err := sqlRunner.CheckUserQuery(query.String(), query.Args()...); err != nil {
		return err
	}
	return nil
}

// updateStatusAndErr update the status and catch create/update error.
func (r *MysqlUserReconciler) updateStatusAndErr(ctx context.Context, mysqlUser *mysqluser.MysqlUser, oldStatus *apiv1alpha1.UserStatus, cuErr error) error {
	if !reflect.DeepEqual(oldStatus, &mysqlUser.Status) {
		userLog.Info("update mysql user status", "name", mysqlUser.Name, "oldHost", oldStatus.AllowedHosts, "newHost", mysqlUser.Spec.Hosts)

		if err := r.Status().Update(ctx, mysqlUser.Unwrap()); err != nil {
			if cuErr != nil {
				return fmt.Errorf("failed to update status: %s, previous error was: %s", err, cuErr)
			}
			return err
		}
	}

	return cuErr
}

// getSqlRunner return a sqlrunner that can execute sql in leader node.
func (r *MysqlUserReconciler) getSqlRunner(ctx context.Context, mysqlUser *mysqluser.MysqlUser) (*internal.SQLRunner, error) {
	releaseName := mysqlUser.Spec.ClusterBinder.ClusterName
	sctName := fmt.Sprintf("%s-secret", releaseName)
	svcName := fmt.Sprintf("%s-leader", releaseName)
	nameSpace := mysqlUser.Spec.ClusterBinder.NameSpace
	port := utils.MysqlPort

	// Get secrets.
	secret := &corev1.Secret{}
	if err := r.Get(context.TODO(),
		types.NamespacedName{
			Namespace: nameSpace,
			Name:      sctName,
		},
		secret,
	); err != nil {
		return nil, fmt.Errorf("failed to get the secret: %s", sctName)
	}

	password, ok := secret.Data["root-password"]
	if !ok {
		return nil, fmt.Errorf("failed to get the password: %s", password)
	}

	sqlRunner, err := internal.NewSQLRunner(utils.BytesToString([]byte(utils.RootUser)), utils.BytesToString(password), svcName, port)
	if err != nil {
		return nil, err
	}

	return sqlRunner, nil
}

// addUserFinalizer add the mysql user finalizer if it`s not exists, before delete mysql user cr , it will wait mysql user finalizer delete first.
// In this way, the logic of deleting the mysql mysql user can be executed before deleting cr.
func (r *MysqlUserReconciler) addUserFinalizer(ctx context.Context, mysqlUser *apiv1alpha1.MysqlUser) error {
	// add finalizer if not present.
	if !controllerutil.ContainsFinalizer(mysqlUser, string(utils.UserFinalizer)) {
		controllerutil.AddFinalizer(mysqlUser, string(utils.UserFinalizer))
	}

	if err := r.Update(ctx, mysqlUser); err != nil {
		userLog.Error(err, "failed to update mysql user status")
	}

	return nil
}

// deleteUserFinalizer delete the mysql user finalizer.
func (r *MysqlUserReconciler) deleteUserFinalizer(ctx context.Context, mysqlUser *apiv1alpha1.MysqlUser) error {
	controllerutil.RemoveFinalizer(mysqlUser, string(utils.UserFinalizer))
	if err := r.Update(ctx, mysqlUser); err != nil {
		userLog.Error(err, "failed to update mysql user status")
	}

	return nil
}

// setFailedStatus update msyql user`s status when operate user faild.
func setFailedStatus(err *error, user *mysqluser.MysqlUser) {
	if *err != nil {
		user.UpdateStatusCondition(
			apiv1alpha1.MySQLUserReady, corev1.ConditionFalse,
			mysqluser.ConfigurationFailedReason, fmt.Sprintf("The mysql user has been configured failed: %s", *err),
		)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.MysqlUser{}).
		Complete(r)
}
