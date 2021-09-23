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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1alhpa1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
)

// UpdateStatusCondition sets the condition to a status.
// for example Ready condition to True, or False.
func (u *MysqlUser) UpdateStatusCondition(
	condType apiv1alhpa1.MysqlUserConditionType,
	status corev1.ConditionStatus, reason, message string,
) (
	cond *apiv1alhpa1.MySQLUserCondition, changed bool,
) {
	t := metav1.NewTime(time.Now())

	existingCondition, exists := u.ConditionExists(condType)
	if !exists {
		newCondition := apiv1alhpa1.MySQLUserCondition{
			Type:               condType,
			Status:             status,
			Reason:             reason,
			Message:            message,
			LastTransitionTime: t,
			LastUpdateTime:     t,
		}
		u.Status.Conditions = append(u.Status.Conditions, newCondition)

		return &newCondition, true
	}

	if status != existingCondition.Status {
		existingCondition.LastTransitionTime = t
		changed = true
	}

	if message != existingCondition.Message || reason != existingCondition.Reason {
		existingCondition.LastUpdateTime = t
		changed = true
	}

	existingCondition.Status = status
	existingCondition.Message = message
	existingCondition.Reason = reason

	return existingCondition, changed
}

// ConditionExists returns a condition and whether it exists.
func (u *MysqlUser) ConditionExists(
	ct apiv1alhpa1.MysqlUserConditionType,
) (
	*apiv1alhpa1.MySQLUserCondition, bool,
) {
	for i := range u.Status.Conditions {
		cond := &u.Status.Conditions[i]
		if cond.Type == ct {
			return cond, true
		}
	}

	return nil, false
}
