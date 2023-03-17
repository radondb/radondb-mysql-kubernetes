package backup

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	v1beta1 "github.com/radondb/radondb-mysql-kubernetes/api/v1beta1"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// BackupReconciler reconciles a Backup object.
type BackupReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Owner    client.FieldOwner
}
type BackupResource struct {
	cronjobs     []*batchv1.CronJob
	jobs         []*batchv1.Job
	mysqlCluster *v1beta1.MysqlCluster
}

//+kubebuilder:rbac:groups=mysql.radondb.com,resources=backups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
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
	log := log.FromContext(ctx).WithName("controllers").WithName("backup")

	result := reconcile.Result{}
	backup := &v1beta1.Backup{}

	if err := r.Client.Get(ctx, req.NamespacedName, backup); err != nil {
		// NotFound cannot be fixed by requeuing so ignore it. During background
		// deletion, we receive delete events from backup's dependents after
		// backup is deleted.
		if err = client.IgnoreNotFound(err); err != nil {
			log.Error(err, "unable to fetch Backup")
		}
		return result, err
	}
	//set default value

	// if backup.Spec.ClusterName is empty, return error
	if backup.Spec.ClusterName == "" {
		return result, errors.New("backup.Spec.ClusterName is empty")
	}
	// get MySQLCluster object
	cluster := &v1beta1.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backup.Spec.ClusterName,
			Namespace: backup.Namespace,
		},
	}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(cluster), cluster); err != nil {
		// NotFound cannot be fixed by requeuing so ignore it. During background
		// deletion, we receive delete events from backup's dependents after
		// backup is deleted.
		if err = client.IgnoreNotFound(err); err != nil {
			log.Error(err, "unable to fetch MysqlCluster")
		}
	}

	var err error
	// Keep a copy of cluster prior to any manipulations.
	before := backup.DeepCopy()

	patchClusterStatus := func() (reconcile.Result, error) {
		if !equality.Semantic.DeepEqual(before.Status, backup.Status) {
			if err := errors.WithStack(r.Client.Status().Patch(
				ctx, backup, client.MergeFrom(before), r.Owner)); err != nil {
				log.Error(err, "patching cluster status")
				return result, err
			}
			log.V(1).Info("patched cluster status")
		}
		return result, err
	}

	// create the Result that will be updated while reconciling any/all backup resources

	backupResources, err := r.getBackupResources(ctx, backup)
	if err != nil {
		// exit early if can't get and clean existing resources as needed to reconcile
		return result, errors.WithStack(err)
	}
	backupResources.mysqlCluster = cluster
	if err := r.reconcileManualBackup(ctx, backup, backupResources.jobs, backupResources.mysqlCluster); err != nil {
		log.Error(err, "unable to reconcile manual backup")
	}
	if err := r.reconcileCronBackup(ctx, backup, backupResources.cronjobs, backupResources.jobs, cluster); err != nil {
		log.Error(err, "unable to reconcile cron backup")
	}
	return patchClusterStatus()
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Backup{}).
		Owns(&batchv1.Job{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}

func (r *BackupReconciler) getBackupResources(ctx context.Context,
	backup *v1beta1.Backup) (*BackupResource, error) {
	// get the cluster
	backupResource := &BackupResource{}
	gvks := []schema.GroupVersionKind{{
		Group:   batchv1.SchemeGroupVersion.Group,
		Version: batchv1.SchemeGroupVersion.Version,
		Kind:    "JobList",
	}, {
		Group:   batchv1.SchemeGroupVersion.Group,
		Version: batchv1.SchemeGroupVersion.Version,
		Kind:    "CronJobList",
	},
	}
	selector := BackupSelector(backup.Spec.ClusterName)
	for _, gvk := range gvks {
		uList := &unstructured.UnstructuredList{}
		uList.SetGroupVersionKind(gvk)
		if err := r.Client.List(ctx, uList,
			client.InNamespace(backup.GetNamespace()),
			client.MatchingLabelsSelector{Selector: selector}); err != nil {
			return nil, errors.WithStack(err)
		}
		if len(uList.Items) == 0 {
			continue
		}
		if err := unstructuredToBackupResources(gvk.Kind, backupResource,
			uList); err != nil {
			return nil, errors.WithStack(err)
		}

	}
	return backupResource, nil
}

func unstructuredToBackupResources(kind string, backupResource *BackupResource,
	uList *unstructured.UnstructuredList) error {
	for _, u := range uList.Items {
		switch kind {
		case "JobList":
			job := &batchv1.Job{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, job); err != nil {
				return errors.WithStack(err)
			}
			backupResource.jobs = append(backupResource.jobs, job)
		case "CronJobList":
			cronjob := &batchv1.CronJob{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cronjob); err != nil {
				return errors.WithStack(err)
			}
			backupResource.cronjobs = append(backupResource.cronjobs, cronjob)
		}
	}
	return nil
}

func (r *BackupReconciler) reconcileManualBackup(ctx context.Context,
	backup *v1beta1.Backup, manualBackupJobs []*batchv1.Job, cluster *v1beta1.MysqlCluster) error {
	manualStatus := backup.Status.ManualBackup
	var currentBackupJob *batchv1.Job

	if backup.Spec.BackupSchedule != nil {
		// if the backup is a scheduled backup, ignore manual backups
		return nil
	}
	if len(manualBackupJobs) > 0 {
		for _, job := range manualBackupJobs {
			if job.GetOwnerReferences()[0].Name == backup.GetName() {
				currentBackupJob = job
				break
			}
		}

		if manualStatus != nil && currentBackupJob != nil {
			completed := jobCompleted(currentBackupJob)
			failed := jobFailed(currentBackupJob)
			manualStatus.CompletionTime = currentBackupJob.Status.CompletionTime
			manualStatus.StartTime = currentBackupJob.Status.StartTime
			manualStatus.Failed = currentBackupJob.Status.Failed
			manualStatus.Succeeded = currentBackupJob.Status.Succeeded
			manualStatus.Active = currentBackupJob.Status.Active
			if completed {
				manualStatus.BackupName = currentBackupJob.GetAnnotations()["backupName"]
				manualStatus.BackupSize = currentBackupJob.GetAnnotations()["backupSize"]
				manualStatus.BackupType = currentBackupJob.GetAnnotations()["backupType"]

			}
			if completed || failed {
				manualStatus.Finished = true
			}

		}

	}

	// if there is an existing status, see if a new backup id has been provided, and if so reset
	// the status and proceed with reconciling a new backup
	if manualStatus == nil {
		manualStatus = &v1beta1.ManualBackupStatus{
			Finished: false,
		}
		backup.Status.ManualBackup = manualStatus
	}

	// if the status shows the Job is no longer in progress, then simply exit (which means a Job
	// that has reached a "completed" or "failed" status is no longer reconciled)
	if manualStatus != nil && manualStatus.Finished {
		return nil
	}

	backupJob := &batchv1.Job{}
	backupJob.ObjectMeta = ManualBackupJobMeta(cluster)
	if currentBackupJob != nil {
		backupJob.ObjectMeta.Name = currentBackupJob.ObjectMeta.Name
	}
	labels := ManualBackupLabels(cluster.Name)
	backupJob.ObjectMeta.Labels = labels

	jobName := backupJob.ObjectMeta.Name
	spec, err := generateBackupJobSpec(backup, cluster, labels, jobName)
	if err != nil {
		return errors.WithStack(err)
	}

	backupJob.Spec = *spec

	backupJob.SetGroupVersionKind(batchv1.SchemeGroupVersion.WithKind("Job"))
	if err := controllerutil.SetControllerReference(backup, backupJob,
		r.Client.Scheme()); err != nil {
		return errors.WithStack(err)
	}

	if err := r.apply(ctx, backupJob); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *BackupReconciler) apply(ctx context.Context, object client.Object) error {
	// Generate an apply-patch by comparing the object to its zero value.
	zero := reflect.New(reflect.TypeOf(object).Elem()).Interface()
	data, err := client.MergeFrom(zero.(client.Object)).Data(object)
	apply := client.RawPatch(client.Apply.Type(), data)

	// Keep a copy of the object before any API calls.
	patch := NewJSONPatch()

	// Send the apply-patch with force=true.
	if err == nil {
		err = r.patch(ctx, object, apply, client.ForceOwnership)
	}

	// Send the json-patch when necessary.
	if err == nil && !patch.IsEmpty() {
		err = r.patch(ctx, object, patch)
	}
	return err
}

func (r *BackupReconciler) patch(
	ctx context.Context, object client.Object,
	patch client.Patch, options ...client.PatchOption,
) error {
	options = append([]client.PatchOption{r.Owner}, options...)
	return r.Client.Patch(ctx, object, patch, options...)
}

func (r *BackupReconciler) reconcileCronBackup(ctx context.Context, backup *v1beta1.Backup,
	cronBackupJobs []*batchv1.CronJob, BackupJobs []*batchv1.Job, cluster *v1beta1.MysqlCluster) error {
	log := log.FromContext(ctx).WithValues("backip", "CronJob")

	if backup.Spec.BackupSchedule == nil {
		// if the backup is a manual backup, ignore scheduled backups
		return nil
	}
	// Update backup.Status.ScheduledBackups
	scheduledStatus := []v1beta1.ScheduledBackupStatus{}
	for _, job := range BackupJobs {
		sbs := v1beta1.ScheduledBackupStatus{}
		if job.GetLabels()[LableCronJob] != "" {
			if len(job.GetOwnerReferences()) > 0 {
				sbs.CronJobName = job.OwnerReferences[0].Name
			}
			sbs.BackupName = job.GetAnnotations()["backupName"]
			sbs.BackupSize = job.GetAnnotations()["backupSize"]
			sbs.CompletionTime = job.Status.CompletionTime
			sbs.Failed = job.Status.Failed
			sbs.Succeeded = job.Status.Succeeded
			scheduledStatus = append(scheduledStatus, sbs)
		}
	}
	// if nil ,create status
	if backup.Status.ScheduledBackups == nil {
		backup.Status.ScheduledBackups = scheduledStatus
	}

	labels := CronBackupLabels(cluster.Name)
	objectMeta := CronBackupJobMeta(cluster)
	for _, cronjob := range cronBackupJobs {
		if cronjob.GetDeletionTimestamp() != nil {
			continue
		}
		if cronjob.GetLabels()[LabelCluster] == cluster.Name &&
			cronjob.GetLabels()[LableCronJob] == "true" {
			objectMeta = metav1.ObjectMeta{
				Namespace: backup.GetNamespace(),
				Name:      cronjob.Name,
			}

		}

	}
	objectMeta.Labels = labels
	// objectmeta.Annotations = annotations
	jobSpec, err := generateBackupJobSpec(backup, cluster, labels, objectMeta.Name)
	if err != nil {
		return errors.WithStack(err)
	}
	suspend := (cluster.Status.State != v1beta1.ClusterReadyState) || (cluster.Spec.Standby != nil)
	cronJob := &batchv1.CronJob{
		ObjectMeta: objectMeta,
		Spec: batchv1.CronJobSpec{
			Schedule:          backup.Spec.BackupSchedule.CronExpression,
			Suspend:           &suspend,
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: *jobSpec,
			},
		},
	}
	cronJob.SetGroupVersionKind(batchv1.SchemeGroupVersion.WithKind("CronJob"))
	if err := controllerutil.SetControllerReference(backup, cronJob,
		r.Client.Scheme()); err != nil {
		return errors.WithStack(err)
	}
	if err := r.apply(ctx, cronJob); err != nil {
		log.Error(err, "error when attempting to create Backup CronJob")
	}

	return nil

}

func generateBackupJobSpec(backup *v1beta1.Backup, cluster *v1beta1.MysqlCluster, labels map[string]string, jobName string) (*batchv1.JobSpec, error) {
	backupHost := GetBackupHost(cluster)
	backupImage := cluster.Spec.Backup.Image
	serviceAccountName := backup.Spec.ClusterName
	clusterAuthsctName := fmt.Sprintf("%s-secret", cluster.GetName())
	s3SecretName := backup.Spec.BackupOpts.S3.BackupSecretName

	container := corev1.Container{
		Env: []corev1.EnvVar{
			{Name: "CONTAINER_TYPE", Value: utils.ContainerBackupJobName},
			{Name: "NAMESPACE", Value: cluster.Namespace},
			{Name: "CLUSTER_NAME", Value: cluster.GetName()},
			{Name: "SERVICE_NAME", Value: fmt.Sprintf("%s-mysql", cluster.GetName())},
			{Name: "HOST_NAME", Value: backupHost},
			{Name: "REPLICAS", Value: "1"},
			{Name: "JOB_NAME", Value: jobName},
		},
		Image:           backupImage,
		ImagePullPolicy: cluster.Spec.ImagePullPolicy,
		Name:            utils.ContainerBackupName,
	}
	container.Args = []string{
		"request_a_backup",
		GetXtrabackupURL(GetBackupHost(cluster)),
	}
	container.Env = append(container.Env,
		getEnvVarFromSecret(s3SecretName, "S3_ENDPOINT", "s3-endpoint", false),
		getEnvVarFromSecret(s3SecretName, "S3_ACCESSKEY", "s3-access-key", true),
		getEnvVarFromSecret(s3SecretName, "S3_SECRETKEY", "s3-secret-key", true),
		getEnvVarFromSecret(s3SecretName, "S3_BUCKET", "s3-bucket", true),
		getEnvVarFromSecret(clusterAuthsctName, "BACKUP_USER", "backup-user", true),
		getEnvVarFromSecret(clusterAuthsctName, "BACKUP_PASSWORD", "backup-password", true),
	)
	jobSpec := &batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: labels},
			Spec: corev1.PodSpec{
				Containers:         []corev1.Container{container},
				RestartPolicy:      corev1.RestartPolicyNever,
				ServiceAccountName: serviceAccountName,
			},
		},
	}
	var backoffLimit int32 = 1

	jobSpec.Template.Spec.Tolerations = cluster.Spec.Tolerations
	jobSpec.Template.Spec.Affinity = cluster.Spec.Affinity
	jobSpec.BackoffLimit = &backoffLimit
	return jobSpec, nil
}
