package backup

import (
	"context"
	"fmt"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
)

// The job structure contains the context to schedule a backup
type CronJob struct {
	ClusterName string
	Namespace   string

	// kubernetes client
	Client client.Client

	BackupScheduleJobsHistoryLimit *int
	Image                          string
	NFSServerAddress               string
	Log                            logr.Logger
}

func (j *CronJob) Run() {
	// nolint: govet
	log := j.Log
	log.Info("scheduled backup job started")

	// run garbage collector if needed
	if j.BackupScheduleJobsHistoryLimit != nil {
		defer j.backupGC()
	}
	if !j.backupNeedRun() {
		log.Info("the cron is deleting when cronjob is starting")
		return
	}
	// check if a backup is running
	if j.scheduledBackupsRunningCount() > 0 {
		log.Info("at least a backup is running", "running_backups_count", j.scheduledBackupsRunningCount())
		return
	}

	// create the backup
	if _, err := j.createBackup(); err != nil {
		log.Error(err, "failed to create backup")
	}
}

func (j *CronJob) scheduledBackupsRunningCount() int {
	log := j.Log
	backupsList := &apiv1alpha1.BackupList{}
	// select all backups with labels recurrent=true and and not completed of the cluster
	selector := j.backupSelector()
	client.MatchingFields{"status.completed": "false"}.ApplyToList(selector)

	if err := j.Client.List(context.TODO(), backupsList, selector); err != nil {
		log.Error(err, "failed getting backups", "selector", selector)
		return 0
	}

	return len(backupsList.Items)
}

func (j *CronJob) backupSelector() *client.ListOptions {
	selector := &client.ListOptions{}

	client.InNamespace(j.Namespace).ApplyToList(selector)
	client.MatchingLabels(j.recurrentBackupLabels()).ApplyToList(selector)

	return selector
}

func (j *CronJob) backupNeedRun() bool {
	// When remove the entries, it may has cron task is running.
	cluster := &apiv1alpha1.MysqlCluster{}
	if err := j.Client.Get(context.TODO(), client.ObjectKey{
		Name:      j.ClusterName,
		Namespace: j.Namespace,
	}, cluster); err != nil {
		return false
	}
	return *cluster.Spec.Replicas != 0
}

func (j *CronJob) recurrentBackupLabels() map[string]string {
	return map[string]string{
		"recurrent": "true",
		"cluster":   j.ClusterName,
	}
}

func (j *CronJob) backupGC() {
	var err error
	log := j.Log
	backupsList := &apiv1alpha1.BackupList{}
	if err = j.Client.List(context.TODO(), backupsList, j.backupSelector()); err != nil {
		log.Error(err, "failed getting backups", "selector", j.backupSelector())
		return
	}

	// sort backups by creation time before removing extra backups
	sort.Sort(byTimestamp(backupsList.Items))

	for i, backup := range backupsList.Items {
		if i >= *j.BackupScheduleJobsHistoryLimit {
			// delete the backup
			if err = j.Client.Delete(context.TODO(), &backup); err != nil {
				log.Error(err, "failed to delete a backup", "backup", backup)
			}
		}
	}
}

func (j *CronJob) createBackup() (*apiv1alpha1.Backup, error) {
	backupName := fmt.Sprintf("%s-auto-%s", j.ClusterName, time.Now().Format("2006-01-02t15-04-05"))

	backup := &apiv1alpha1.Backup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupName,
			Namespace: j.Namespace,
			Labels:    j.recurrentBackupLabels(),
		},
		Spec: apiv1alpha1.BackupSpec{
			ClusterName: j.ClusterName,
			//TODO modify to cluster sidecar image
			Image: j.Image,
			//RemoteDeletePolicy: j.BackupRemoteDeletePolicy,
			HostName: fmt.Sprintf("%s-mysql-0", j.ClusterName),
		},
	}
	if len(j.NFSServerAddress) > 0 {
		backup.Spec.NFSServerAddress = j.NFSServerAddress
	}
	return backup, j.Client.Create(context.TODO(), backup)
}

type byTimestamp []apiv1alpha1.Backup

func (a byTimestamp) Len() int      { return len(a) }
func (a byTimestamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byTimestamp) Less(i, j int) bool {
	return a[j].ObjectMeta.CreationTimestamp.Before(&a[i].ObjectMeta.CreationTimestamp)
}
