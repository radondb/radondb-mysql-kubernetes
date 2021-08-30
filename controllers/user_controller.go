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

	mysqlv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/internal"
	mysqlUser "github.com/radondb/radondb-mysql-kubernetes/user"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=mysql.radondb.com,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.radondb.com,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mysql.radondb.com,resources=users/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the User object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("controllers").WithName("User")

	// your logic here
	user := mysqlUser.New(&mysqlv1alpha1.User{})

	err := r.Get(ctx, req.NamespacedName, user.Unwrap())
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			log.Info("user not found, maybe removed")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	oldStatus := user.Status.DeepCopy()

	sqlRunner, err := r.getSqlRunner(ctx, user.Unwrap())

	if err != nil {
		return ctrl.Result{}, fmt.Errorf("get sqlrunner faild, err: %s", err)
	}

	defer sqlRunner.Close()

	// If the user has been deleted then remove it from mysql cluster.
	if !user.ObjectMeta.DeletionTimestamp.IsZero() {
		queries := r.getDropQuery(ctx, user.Status.AllowedHosts, user.Spec.User)
		return ctrl.Result{}, r.removeUser(ctx, queries, sqlRunner, user.Unwrap())
	}
	// If user present, create or update.
	cuErr := r.createOrUpdate(ctx, sqlRunner, user)
	if cuErr != nil {
		return ctrl.Result{}, err
	}
	if r.updateStatusAndErr(ctx, user.Unwrap(), oldStatus, cuErr); err != nil {
		return ctrl.Result{}, err
	}

	user.UpdateStatusCondition(
		mysqlv1alpha1.MySQLUserReady, corev1.ConditionTrue,
		mysqlUser.ConfigurationSucceededReason, "The user has been configured successfully.",
	)

	// enqueue the resource again after to keep the resource up to date in mysql
	// in case is changed directly into mysql
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: 2 * time.Minute,
	}, nil
}

// createOrUpdate determine the type of operation and call the corresponding function
func (r *UserReconciler) createOrUpdate(ctx context.Context, sqlRunner *internal.SQLRunner, user *mysqlUser.User) error {
	log := log.FromContext(ctx).WithName("controllers").WithName("User")

	if user.Spec.User == "" {
		return fmt.Errorf("user name must not be empty")
	}

	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{Name: user.Spec.SecretBinder.SecretName, Namespace: user.Namespace}

	// Get secret object of user.
	if err := r.Get(ctx, secretKey, secret); err != nil {
		return err
	}

	// When apply the user crd, the corresponding secret is created.
	password := string(secret.Data[user.Spec.SecretBinder.SecretKey])
	// Must check password is not empty first, if not, create user.
	if password == "" {
		return fmt.Errorf("password must not be empty")
	}

	// TODO: need update user logic

	// Try create normal user with Permissions.
	// TODO: merge this function to createUser
	if len(user.Spec.Permissions) > 0 {
		return r.createUserWithGrants(ctx, user.Unwrap())
	}

	unexistHost := []string{}
	for _, host := range user.Spec.Hosts {
		// if user not exist, create
		if err := r.checkUserExist(ctx, r.checkUserQuery(ctx, user.Spec.User, host), sqlRunner); err != nil {
			log.Info("user not exist, create new user", "need create host", host)
			// Try create user.
			unexistHost = append(unexistHost, host)
		}
		// TODO: check if user need update.
	}

	if len(unexistHost) == 0 {
		log.Info("Users already exist", "userName:", user.Spec.User, "hosts:", user.Status.AllowedHosts)
		return nil
	}

	queries := r.getCreateQuery(ctx, unexistHost, user.Spec.User, password)

	if err := r.createUser(ctx, queries, sqlRunner); err != nil {
		return err
	}

	user.UpdateStatusCondition(
		mysqlv1alpha1.MySQLUserReady, corev1.ConditionFalse,
		mysqlUser.ConfigurationFailedReason, "User configuration failed.",
	)

	// update status for allowedHosts if needed, mark that status need to be updated
	if !reflect.DeepEqual(user.Status.AllowedHosts, user.Spec.Hosts) {
		user.Status.AllowedHosts = user.Spec.Hosts
	}

	// add finalizer for user.
	r.addUserFinalizer(ctx, user.Unwrap())

	return nil
}

// createUser create users in mysql based on cr information.
func (r *UserReconciler) createUser(ctx context.Context, queries []internal.Query, sqlRunner *internal.SQLRunner) error {
	log := log.FromContext(ctx).WithName("controllers").WithName("User")

	query := internal.BuildAtomicQuery(queries...)

	// TODO: add check for different password. 其他信息相同，密码不同的情况 IF NOT EXISTS 会正常执行，需要针对这种情况校验。
	// TODO: if user exist, print 'user exist'. now is 'create user success'
	// TODO: check if the user exist in db, when mysql reload, user need recreate.
	if err := sqlRunner.RunQuery(query.String(), query.Args()...); err != nil {
		return fmt.Errorf("create user faild, err: %s", err)
	}

	log.Info("create user success")

	return nil
}

// removeUser called when remove the user cr, it will remove all users in mysql that has been created.
func (r *UserReconciler) removeUser(ctx context.Context, queries []internal.Query, sqlRunner *internal.SQLRunner, user *mysqlv1alpha1.User) error {
	log := log.FromContext(ctx).WithName("controllers").WithName("User")

	query := internal.BuildAtomicQuery(queries...)

	if controllerutil.ContainsFinalizer(user, string(utils.UserFinalizer)) {
		if err := sqlRunner.RunQuery(query.String(), query.Args()...); err != nil {
			return fmt.Errorf("drop user faild, err: %s", err)
		}

		if err := r.removeUserFinalizer(ctx, user); err != nil {
			return fmt.Errorf("remove user finalizer faild, err: %s", err)
		}

		log.Info("drop user success")
	}

	return nil
}

// checkUserExist check if the user exist.
func (r *UserReconciler) checkUserExist(ctx context.Context, query internal.Query, sqlRunner *internal.SQLRunner) error {
	if err := sqlRunner.CheckUserQuery(query.String(), query.Args()...); err != nil {
		return err
	}
	return nil
}

// getCreateQuery get the queries of create users.
func (r *UserReconciler) getCreateQuery(ctx context.Context, hosts []string, userName, password string) []internal.Query {
	queries := []internal.Query{}

	for _, host := range hosts {
		queries = append(queries, internal.NewQuery("CREATE USER IF NOT EXISTS ?@? IDENTIFIED BY ?;", userName, host, password))
	}

	return queries
}

// getDropQuery get the queries of drop users.
func (r *UserReconciler) getDropQuery(ctx context.Context, hosts []string, userName string) []internal.Query {
	queries := []internal.Query{}

	for _, host := range hosts {
		queries = append(queries, internal.NewQuery("DROP USER IF EXISTS ?@?;", userName, host))
	}

	return queries
}

// checkUserQuery get the query of check if the user exist.
func (r *UserReconciler) checkUserQuery(ctx context.Context, name, host string) internal.Query {
	return internal.NewQuery("SELECT USER FROM mysql.user WHERE user = ? AND host = ?;", name, host)
}

// createUserWithGrants is a demo.
// TODO: it may be merged with createUser in the future.
func (r *UserReconciler) createUserWithGrants(ctx context.Context, user *mysqlv1alpha1.User) error {
	return fmt.Errorf("create NomalUser With Grants")
}

// updateStatusAndErr update the status and catch create/update error.
func (r *UserReconciler) updateStatusAndErr(ctx context.Context, user *mysqlv1alpha1.User, oldStatus *mysqlv1alpha1.UserStatus, cuErr error) error {
	log := log.FromContext(ctx).WithName("controllers").WithName("User")

	if !reflect.DeepEqual(oldStatus, &user.Status) {
		log.Info("update mysql user status", "name", user.Name, "oldHost", oldStatus.AllowedHosts, "newHost", user.Spec.Hosts)

		if err := r.Status().Update(ctx, user); err != nil {
			if cuErr != nil {
				return fmt.Errorf("failed to update status: %s, previous error was: %s", err, cuErr)
			}

			return err
		}
	}

	return cuErr
}

// getsqlRunner return a sqlrunner that can execute sql in leader node.
func (r *UserReconciler) getSqlRunner(ctx context.Context, user *mysqlv1alpha1.User) (*internal.SQLRunner, error) {
	releaseName := user.Spec.ClusterBinder.ClusterName
	sctName := fmt.Sprintf("%s-secret", releaseName)
	svcName := fmt.Sprintf("%s-leader", releaseName)
	nameSpace := user.Spec.ClusterBinder.NameSpace
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
	operatorUser, ok := secret.Data["operator-user"]
	if !ok {
		return nil, fmt.Errorf("failed to get the user: %s", operatorUser)
	}
	password, ok := secret.Data["operator-password"]
	if !ok {
		return nil, fmt.Errorf("failed to get the password: %s", password)
	}

	sqlRunner, err := internal.NewSQLRunner(utils.BytesToString(operatorUser), utils.BytesToString(password), svcName, port)
	if err != nil {
		return nil, err
	}

	return sqlRunner, nil
}

// addUserFinalizer add the user finalizer if it`s not exists, before delete user cr , it will wait user finalizer delete first.
// In this way, the logic of deleting the mysql user can be executed before deleting cr.
func (r *UserReconciler) addUserFinalizer(ctx context.Context, user *mysqlv1alpha1.User) error {
	// add finalizer if not present.

	if !controllerutil.ContainsFinalizer(user, string(utils.UserFinalizer)) {
		controllerutil.AddFinalizer(user, string(utils.UserFinalizer))
	}

	if err := r.Update(ctx, user); err != nil {
		log.Log.Error(err, "failed to update user status")
	}

	return nil
}

// removeUserFinalizer remove the user finalizer.
func (r *UserReconciler) removeUserFinalizer(ctx context.Context, user *mysqlv1alpha1.User) error {
	controllerutil.RemoveFinalizer(user, string(utils.UserFinalizer))

	if err := r.Update(ctx, user); err != nil {
		log.Log.Error(err, "failed to update user status")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlv1alpha1.User{}).
		Complete(r)
}
