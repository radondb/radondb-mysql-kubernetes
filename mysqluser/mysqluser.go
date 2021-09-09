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
	v1alhpa1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
)

const (
	// ConfigurationFailedReason is the condition reason when MysqlUser has failed to configure.
	ConfigurationFailedReason = "ConfigurationFailed"
	// ConfigurationInProgressReason is the reason when MysqlUser is being configured.
	ConfiguringReason = "Configuring"
	// ConfigurationSucceededReason the reason used when Configuration was successful.
	ConfigurationSucceededReason = "ConfigurationSucceeded"
)

// MysqlUser is a type wrapper over MysqlUser that contains the Business logic.
type MysqlUser struct {
	*v1alhpa1.MysqlUser
}

// New returns a wraper object over MysqlUser.
func New(mysqlUser *v1alhpa1.MysqlUser) *MysqlUser {
	return &MysqlUser{
		MysqlUser: mysqlUser,
	}
}

// Unwrap returns the api MysqlUser object.
func (u *MysqlUser) Unwrap() *v1alhpa1.MysqlUser {
	return u.MysqlUser
}
