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

package user

import (
	"fmt"

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

// User is a type wrapper over User that contains the Business logic
type User struct {
	*v1alhpa1.User
}

// New returns a wraper object over User
func New(user *v1alhpa1.User) *User {
	return &User{
		User: user,
	}
}

// Unwrap returns the api user object
func (u *User) Unwrap() *v1alhpa1.User {
	return u.User
}

// GetNameForJob returns the name of the job
func (u *User) GetNameForJob() string {
	return fmt.Sprintf("%s-user", u.Name)
}
