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
	"context"
	"fmt"
	"net"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
//+kubebuilder:webhook:path=/validate-mysql-radondb-com-v1alpha1-mysqlcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=mysql.radondb.com,resources=mysqlclusters,verbs=create;update,versions=v1alpha1,name=vmysqlcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &MysqlCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MysqlCluster) ValidateCreate() error {
	mysqlclusterlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	if err := r.validateNFSServerAddress(r); err != nil {
		return err
	}

	if err := r.validBothS3NFS(); err != nil {
		return err
	}

	if err := r.validateMysqlVersionAndImage(); err != nil {
		return err
	}

	if err := r.ValidMySQLTemplate(); err != nil {
		return err
	}
	if err := r.validateMysqlVersion(); err != nil {
		return err
	}
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

	if err := r.validBothS3NFS(); err != nil {
		return err
	}

	if err := r.validateMysqlVersionAndImage(); err != nil {
		return err
	}

	if err := r.ValidMySQLTemplate(); err != nil {
		return err
	}
	if err := r.validateMysqlVersion(); err != nil {
		return err
	}
	if err := r.validateNFSServerAddress(oldCluster); err != nil {
		return err
	}
	return nil
}

func (r *MysqlCluster) ValidMySQLTemplate() error {

	tmplName := r.Spec.MysqlOpts.MysqlConfTemplate
	if len(tmplName) != 0 {
		// check whether the template is exist.
		if ok, err := getCofigMap(tmplName, r.Namespace); !ok {
			return apierrors.NewForbidden(schema.GroupResource{}, "",
				fmt.Errorf("configmap is not exist! %s", err.Error()))
		}
	}
	return nil
}

func getCofigMap(name string, namespace string) (bool, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return false, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return false, err
	}
	if _, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, v1.GetOptions{}); err != nil {
		return false, err
	}
	return true, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MysqlCluster) ValidateDelete() error {
	mysqlclusterlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

// Add NFSServerAddress webhook & backup schedule.
func (r *MysqlCluster) validateNFSServerAddress(oldCluster *MysqlCluster) error {
	//nfsaddress format is x.x.x.x:/yyy such as 10.233.55.172:/backup
	ipStr := strings.Split(r.Spec.NFSServerAddress, ":")
	if len(r.Spec.NFSServerAddress) != 0 && len(ipStr) == 0 {
		return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("nfsServerAddress should be set as IP:/PATH or IP"))
	}
	isIP := net.ParseIP(ipStr[0]) != nil
	if len(r.Spec.NFSServerAddress) != 0 && !isIP {
		return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("nfsServerAddress should be set as  IP:/PATH or IP"))
	}
	if len(r.Spec.BackupSchedule) != 0 && len(r.Spec.BackupSecretName) == 0 && !isIP {
		return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("backupSchedule is set without any backupSecretName or nfsServerAddress"))
	}
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
	if strings.Contains(r.Spec.MysqlOpts.Image, "8.0") &&
		oldmyconf["lower_case_table_names"] != newmyconf["lower_case_table_names"] {
		return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("lower_case_table_names cannot be changed in MySQL8.0+"))
	}
	return nil
}

// Validate MysqlVersion and spec.MysqlOpts.image are conflict.
func (r *MysqlCluster) validateMysqlVersionAndImage() error {
	if r.Spec.MysqlOpts.Image != "" && r.Spec.MysqlVersion != "" {
		if !strings.Contains(r.Spec.MysqlOpts.Image, r.Spec.MysqlVersion) {
			return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("spec.MysqlOpts.Image and spec.MysqlVersion are conflict"))
		}
	}
	return nil
}

// Validate MySQL version and related image.
func (r *MysqlCluster) validateMysqlVersion() error {
	switch r.Spec.MysqlVersion {
	case "5.7":
		if !strings.Contains(r.Spec.PodPolicy.SidecarImage, "57") {
			return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("spec.MysqlVersion is 5.7, but spec.PodPolicy.SidecarImage is not 5.7"))
		}
	case "8.0":
		if !strings.Contains(r.Spec.PodPolicy.SidecarImage, "80") {
			return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("spec.MysqlVersion is 8.0, but spec.PodPolicy.SidecarImage is not 8.0"))
		}
	default:
		return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("spec.MysqlVersion is not provided"))

	}
	return nil
}

// Validate BothS3NFS
func (r *MysqlCluster) validBothS3NFS() error {
	if r.Spec.BothS3NFS != nil &&
		(len(r.Spec.NFSServerAddress) == 0 ||
			len(r.Spec.BackupSecretName) == 0) {
		return apierrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("if BothS3NFS is set, backupSchedule/backupSecret/nfsAddress should not empty"))
	}
	return nil
}
