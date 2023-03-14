package backup

import (
	"context"
	"fmt"

	v1beta1 "github.com/radondb/radondb-mysql-kubernetes/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BackupReconciler reconciles a Backup object.
type BackupReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	cluster  *v1beta1.MysqlCluster
}

//+kubebuilder:rbac:groups=mysql.radondb.com,resources=backups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mysql.radondb.com,resources=backups/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Backup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("controllers").WithName("mysqlbackup")
	// 1. 获取备份对象
	var backup v1beta1.Backup
	if err := r.Get(ctx, req.NamespacedName, &backup); err != nil {
		if errors.IsNotFound(err) {
			// 对象已被删除，清理相关资源并退出
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get backup")
		return ctrl.Result{}, err
	}
	// check cluster exists or not before backup
	var cluster v1beta1.MysqlCluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "Failed to get backup cluster")
			return ctrl.Result{}, err
		}
		// cache cluster sidecar image to backup
		r.cluster = &cluster
	}

	// 2. 检查备份是否已被删除，如果是则清理相关资源并退出
	if backup.ObjectMeta.DeletionTimestamp != nil {
		// 删除备份对应的资源并退出
		return ctrl.Result{}, nil
	}

	// 3. 检查备份是否需要创建
	if backup.Spec.BackupSchedule.CronExpression == "" {
		// 如果CronExpression为空，则创建 ManualBackup Job
		if err := r.createManualBackupJob(ctx, backup); err != nil {
			log.Error(err, "Failed to create manual backup job")
			return ctrl.Result{}, err
		}
	} else {
		// 如果CronExpression不为空，则创建 ScheduledBackup CronJob
		if err := r.createScheduledBackupCronJob(ctx, backup); err != nil {
			log.Error(err, "Failed to create scheduled backup cronjob")
			return ctrl.Result{}, err
		}
	}

	// 4. 更新备份状态并保存
	if err := r.Status().Update(ctx, &backup); err != nil {
		log.Error(err, "Failed to update backup status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *BackupReconciler) createManualBackupJob(ctx context.Context, backup v1beta1.Backup) error {
	jobName := getManualBackupJobName(&backup)

	// 构造ManualBackup Job对象
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: backup.Namespace,
			Labels: map[string]string{
				"app":      "backup-operator",
				"job-type": "manual-backup",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "backup-operator"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "backup",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Image:           r.cluster.Spec.Backup.Image,
							Args:            []string{"backup", "-type", string(backup.Spec.BackupType)},
							Env: []corev1.EnvVar{
								{Name: "BACKUP_NAME", Value: backup.Name},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	// 创建ManualBackup Job
	err := r.Create(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to create manual backup job: %w", err)
	}

	// 更新备份状态
	backup.Status.ManualBackup = &v1beta1.BackupJobStatus{
		BackupName:     backup.Name,
		BackupType:     backup.Spec.BackupType,
		StartTime:      metav1.Now(),
		CompletionTime: nil,
		Succeeded:      false,
		Reason:         "",
		Finished:       false,
		JobName:        jobName,
	}

	return nil
}

func (r *BackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Backup{}).
		Complete(r)
}

func (r *BackupReconciler) createScheduledBackupCronJob(ctx context.Context, backup v1beta1.Backup) error {
	cronJobName := getScheduledBackupCronJobName(&backup)

	// 构造ScheduledBackup CronJob对象
	cronJob := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronJobName,
			Namespace: backup.Namespace,
			Labels: map[string]string{
				"app":      "backup-operator",
				"job-type": "scheduled-backup",
			},
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:                   backup.Spec.BackupSchedule.CronExpression,
			JobTemplate:                batchv1beta1.JobTemplateSpec{},
			SuccessfulJobsHistoryLimit: &backup.Spec.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     &backup.Spec.FailedJobsHistoryLimit,
		},
	}
	cronJob.Spec.JobTemplate.Spec = batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": "backup-operator"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            "backup",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Image:           backup.Spec.BackupImage,
						Args:            []string{"backup", "-type", string(backup.Spec.BackupType)},
						Env: []corev1.EnvVar{
							{Name: "BACKUP_NAME", Value: backup.Name},
						},
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
			},
		},
	}

	// 创建ScheduledBackup CronJob
	err := r.Create(ctx, cronJob)
	if err != nil {
		return fmt.Errorf("failed to create scheduled backup cronjob: %w", err)
	}

	// 更新备份状态
	for _, schedule := range cronJob.Status.Active {
		backup.Status.ScheduledBackups = append(backup.Status.ScheduledBackups, v1alpha1.BackupJobStatus{
			BackupName:     backup.Name,
			BackupType:     backup.Spec.BackupType,
			StartTime:      metav1.Now(),
			CompletionTime: nil,
			Succeeded:      false,
			Reason:         "",
			Finished:       false,
			JobName:        schedule.Name,
		})
	}

	return nil
}

func getManualBackupJobName(backup *v1beta1.Backup) string {
	return fmt.Sprintf("%s-manual-backup", backup.Name)
}
