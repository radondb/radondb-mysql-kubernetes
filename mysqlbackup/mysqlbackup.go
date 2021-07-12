package mysqlbackup

import (
	"fmt"

	v1alhpa1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("mysqlbackup")

// MysqlBackup is a type wrapper over MysqlBackup that contains the Business logic
type MysqlBackup struct {
	*v1alhpa1.MysqlBackup
}

// New returns a wraper object over MysqlBackup
func New(backup *v1alhpa1.MysqlBackup) *MysqlBackup {
	return &MysqlBackup{
		MysqlBackup: backup,
	}
}

// Unwrap returns the api mysqlbackup object
func (b *MysqlBackup) Unwrap() *v1alhpa1.MysqlBackup {
	return b.MysqlBackup
}

// GetNameForJob returns the name of the job
func (b *MysqlBackup) GetNameForJob() string {
	return fmt.Sprintf("%s-backup", b.Name)
}

func (b *MysqlBackup) GetBackupURL(cluster_name string, hostname string) string {
	if len(hostname) != 0 {
		return fmt.Sprintf("%s.%s-mysql.%s:%v", hostname, cluster_name, b.Namespace, utils.XBackupPort)
	} else {
		return fmt.Sprintf("%s-leader.%s:%v", cluster_name, b.Namespace, utils.XBackupPort)
	}
}
