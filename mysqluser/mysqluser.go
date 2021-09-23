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

package mysqluser

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alhpa1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
)

const (
	// ProvisionFailedReason is the condition reason when MysqlUser provisioning
	// has failed.
	ProvisionFailedReason = "ProvisionFailed"
	// ProvisionInProgressReason is the reason when MysqlUser provisioning has
	// started.
	ProvisionInProgressReason = "ProvisionInProgress"

	// ProvisionSucceededReason the reason used when provision was successful.
	ProvisionSucceededReason = "ProvisionSucceeded"
)

// MysqlUser is a type wrapper over MysqlUser that contains the Business logic.
type MysqlUser struct {
	*apiv1alhpa1.MysqlUser
}

// New returns a wraper object over MysqlUser.
func New(mysqlUser *apiv1alhpa1.MysqlUser) *MysqlUser {
	return &MysqlUser{
		MysqlUser: mysqlUser,
	}
}

// Unwrap returns the api MysqlUser object.
func (u *MysqlUser) Unwrap() *apiv1alhpa1.MysqlUser {
	return u.MysqlUser
}

// GetClusterKey returns the MysqlUser's MySQLCluster key.
func (u *MysqlUser) GetClusterKey() client.ObjectKey {
	ns := u.Spec.UserOwner.NameSpace
	if ns == "" {
		ns = u.Namespace
	}

	return client.ObjectKey{
		Name:      u.Spec.UserOwner.ClusterName,
		Namespace: ns,
	}
}

// GetKey return the user key. Usually used for logging or for runtime.Client.Get as key.
func (u *MysqlUser) GetKey() client.ObjectKey {
	return types.NamespacedName{
		Namespace: u.Namespace,
		Name:      u.Name,
	}
}
