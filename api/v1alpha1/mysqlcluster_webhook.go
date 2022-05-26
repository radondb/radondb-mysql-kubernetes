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

package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var mysqlclusterlog = logf.Log.WithName("mysqlcluster-resource")

func (r *MysqlCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/convert,mutating=false,failurePolicy=fail,sideEffects=None,groups=mysql.radondb.com,resources=mysqlclusters,verbs=create;update,versions=v1alpha1,name=vmysqlcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &MysqlCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MysqlCluster) ValidateCreate() error {
	mysqlclusterlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MysqlCluster) ValidateUpdate(old runtime.Object) error {
	mysqlclusterlog.Info("validate update", "name", r.Name)

	oldCluster, ok := old.(*MysqlCluster)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected an MysqlCluster but got a %T", old))
	}
	if err := r.validateVolumeSize(oldCluster); err != nil {
		return err
	}
	if err := r.validateLowTableCase(oldCluster); err != nil {
		return err
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MysqlCluster) ValidateDelete() error {
	mysqlclusterlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

// Validate volume size, forbidden shrink storage size.
func (r *MysqlCluster) validateVolumeSize(oldCluster *MysqlCluster) error {
	oldStorageSize, err := resource.ParseQuantity(oldCluster.Spec.Persistence.Size)
	if err != nil {
		return err
	}
	newStorageSize, err := resource.ParseQuantity(r.Spec.Persistence.Size)
	if err != nil {
		return err
	}
	// =1 means that old storage size is greater than new.
	if oldStorageSize.Cmp(newStorageSize) == 1 {
		return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("volesize can not be decreased"))
	}
	return nil
}

// Validate low table case for mysqlcluster.
func (r *MysqlCluster) validateLowTableCase(oldCluster *MysqlCluster) error {
	oldmyconf := oldCluster.Spec.MysqlOpts.MysqlConf
	newmyconf := r.Spec.MysqlOpts.MysqlConf
	if r.Spec.MysqlVersion == "8.0" &&
		oldmyconf["lower_case_table_names"] != newmyconf["lower_case_table_names"] {
		return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("lower_case_table_names cannot be changed in MySQL8.0+"))
	}
	return nil
}
