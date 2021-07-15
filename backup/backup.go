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

package backup

import (
	"fmt"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1alhpa1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
)

var log = logf.Log.WithName("backup")

// Backup is a type wrapper over Backup that contains the Business logic
type Backup struct {
	*v1alhpa1.Backup
}

// New returns a wraper object over Backup
func New(backup *v1alhpa1.Backup) *Backup {
	return &Backup{
		Backup: backup,
	}
}

// Unwrap returns the api backup object
func (b *Backup) Unwrap() *v1alhpa1.Backup {
	return b.Backup
}

// GetNameForJob returns the name of the job
func (b *Backup) GetNameForJob() string {
	return fmt.Sprintf("%s-backup", b.Name)
}

func (b *Backup) GetBackupURL(cluster_name string, hostname string) string {
	return fmt.Sprintf("%s.%s-mysql.%s:%v", hostname, cluster_name, b.Namespace, utils.XBackupPort)

}
