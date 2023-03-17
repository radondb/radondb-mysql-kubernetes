package sidecar

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type BackupClientConfig struct {
	BackupName        string `json:"backup_name"`
	NameSpace         string `json:"namespace"`
	ServiceName       string `json:"service_name"`
	BackupUser        string `json:"backup_user"`
	BackupPassword    string `json:"backup_password"`
	JobName           string `json:"job_name"`
	ClusterName       string `json:"cluster_name"`
	RootPassword      string `json:"root_password"`
	XCloudS3EndPoint  string `json:"xcloud_s3_endpoint"`
	XCloudS3AccessKey string `json:"xcloud_s3_access_key"`
	XCloudS3SecretKey string `json:"xcloud_s3_secret_key"`
	XCloudS3Bucket    string `json:"xcloud_s3_bucket"`
	// NFS server which Restore from
	XRestoreFromNFS string `json:"xrestore_from_nfs"`
	// XtrabackupExtraArgs is a list of extra command line arguments to pass to xtrabackup.
	XtrabackupExtraArgs []string `json:"xtrabackup_extra_args"`
	// XtrabackupTargetDir is a backup destination directory for xtrabackup.
	XtrabackupTargetDir string `json:"xtrabackup_target_dir"`
}

// NewReqBackupConfig returns the configuration file needed for backup job call /backup.
// The configuration file is obtained from the environment variables.
func NewReqBackupConfig() *BackupClientConfig {
	BackupName, _ := utils.BuildBackupName(getEnvValue("CLUSTER_NAME"))
	return &BackupClientConfig{
		NameSpace:         getEnvValue("NAMESPACE"),
		ServiceName:       getEnvValue("SERVICE_NAME"),
		BackupUser:        getEnvValue("BACKUP_USER"),
		BackupPassword:    getEnvValue("BACKUP_PASSWORD"),
		JobName:           getEnvValue("JOB_NAME"),
		ClusterName:       getEnvValue("CLUSTER_NAME"),
		RootPassword:      getEnvValue("MYSQL_ROOT_PASSWORD"),
		XCloudS3EndPoint:  getEnvValue("S3_ENDPOINT"),
		XCloudS3AccessKey: getEnvValue("S3_ACCESSKEY"),
		XCloudS3SecretKey: getEnvValue("S3_SECRETKEY"),
		XCloudS3Bucket:    getEnvValue("S3_BUCKET"),
		BackupName:        BackupName,
	}
}

// Build xbcloud arguments
func (cfg *BackupClientConfig) XCloudArgs(backupName string) []string {
	xcloudArgs := []string{
		"put",
		"--storage=S3",
		fmt.Sprintf("--s3-endpoint=%s", cfg.XCloudS3EndPoint),
		fmt.Sprintf("--s3-access-key=%s", cfg.XCloudS3AccessKey),
		fmt.Sprintf("--s3-secret-key=%s", cfg.XCloudS3SecretKey),
		fmt.Sprintf("--s3-bucket=%s", cfg.XCloudS3Bucket),
		"--parallel=10",
		// utils.BuildBackupName(cfg.ClusterName),
		cfg.BackupName,
		"--insecure",
	}
	return xcloudArgs
}

func RunTakeBackupCommand(cfg *BackupClientConfig) (string, string, error) {
	// cfg->XtrabackupArgs()
	xtrabackup := exec.Command(xtrabackupCommand, cfg.XtrabackupArgs()...)

	var err error
	backupName, DateTime := cfg.XBackupName()
	xcloud := exec.Command(xcloudCommand, cfg.XCloudArgs(backupName)...)
	log.Info("xargs ", "xargs", strings.Join(cfg.XCloudArgs(backupName), " "))
	if xcloud.Stdin, err = xtrabackup.StdoutPipe(); err != nil {
		log.Error(err, "failed to pipline")
		return "", "", err
	}
	xtrabackup.Stderr = os.Stderr
	xcloud.Stderr = os.Stderr

	if err := xtrabackup.Start(); err != nil {
		log.Error(err, "failed to start xtrabackup command")
		return "", "", err
	}
	if err := xcloud.Start(); err != nil {
		log.Error(err, "fail start xcloud ")
		return "", "", err
	}

	// pipe command fail one, whole things fail
	errorChannel := make(chan error, 2)
	go func() {
		errorChannel <- xcloud.Wait()
	}()
	go func() {
		errorChannel <- xtrabackup.Wait()
	}()

	for i := 0; i < 2; i++ {
		if err = <-errorChannel; err != nil {
			return "", "", err
		}
	}
	return backupName, DateTime, nil
}

func (cfg *BackupClientConfig) XtrabackupArgs() []string {
	// xtrabackup --backup <args> --target-dir=<backup-dir> <extra-args>
	tmpdir := "/root/backup/"
	if len(cfg.XtrabackupTargetDir) != 0 {
		tmpdir = cfg.XtrabackupTargetDir
	}
	xtrabackupArgs := []string{
		"--backup",
		"--stream=xbstream",
		"--host=127.0.0.1",
		fmt.Sprintf("--user=%s", utils.RootUser),
		fmt.Sprintf("--password=%s", cfg.RootPassword),
		fmt.Sprintf("--target-dir=%s", tmpdir),
	}

	return append(xtrabackupArgs, cfg.XtrabackupExtraArgs...)
}

func (cfg *BackupClientConfig) XBackupName() (string, string) {
	return utils.BuildBackupName(cfg.ClusterName)
}

func setAnnonations(cfg *BackupClientConfig, backname string, DateTime string, BackupType string, BackupSize int64) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	job, err := clientset.BatchV1().Jobs(cfg.NameSpace).Get(context.TODO(), cfg.JobName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if job.Annotations == nil {
		job.Annotations = make(map[string]string)
	}
	job.Annotations[utils.JobAnonationName] = backname
	job.Annotations[utils.JobAnonationDate] = DateTime
	job.Annotations[utils.JobAnonationType] = BackupType
	job.Annotations[utils.JobAnonationSize] = strconv.FormatInt(BackupSize, 10)
	_, err = clientset.BatchV1().Jobs(cfg.NameSpace).Update(context.TODO(), job, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func RunTakeS3BackupCommand(cfg *BackupClientConfig) (string, string, int64, error) {
	// cfg->XtrabackupArgs()
	xtrabackup := exec.Command(xtrabackupCommand, cfg.XtrabackupArgs()...)

	var err error
	backupName, DateTime := cfg.XBackupName()
	xcloud := exec.Command(xcloudCommand, cfg.XCloudArgs(backupName)...)
	log.Info("xargs ", "xargs", strings.Join(cfg.XCloudArgs(backupName), " "))

	// Create a pipe between xtrabackup and xcloud
	r, w := io.Pipe()
	defer r.Close()
	xcloud.Stdin = r

	// Start xtrabackup command with stdout directed to the pipe
	xtrabackupReader, err := xtrabackup.StdoutPipe()
	if err != nil {
		log.Error(err, "failed to create stdout pipe for xtrabackup")
		return "", "", 0, err
	}

	// set xtrabackup and xcloud stderr to os.Stderr
	xtrabackup.Stderr = os.Stderr
	xcloud.Stderr = os.Stderr

	// Start xtrabackup and xcloud in separate goroutines
	if err := xtrabackup.Start(); err != nil {
		log.Error(err, "failed to start xtrabackup command")
		return "", "", 0, err
	}
	if err := xcloud.Start(); err != nil {
		log.Error(err, "fail start xcloud ")
		return "", "", 0, err
	}

	// Use io.Copy to write xtrabackup output to the pipe while tracking the number of bytes written
	var n int64
	go func() {
		n, err = io.Copy(w, xtrabackupReader)
		if err != nil {
			log.Error(err, "failed to write xtrabackup output to pipe")
		}
		w.Close()
	}()

	// Wait for xtrabackup and xcloud to finish
	// pipe command fail one, whole things fail
	errorChannel := make(chan error, 2)
	go func() {
		errorChannel <- xcloud.Wait()
	}()
	go func() {
		errorChannel <- xtrabackup.Wait()
	}()

	for i := 0; i < 2; i++ {
		if err = <-errorChannel; err != nil {
			// If xtrabackup or xcloud failed, stop the pipe and kill the other command
			log.Error(err, "xtrabackup or xcloud failed closing the pipe...")
			xtrabackup.Process.Kill()
			xcloud.Process.Kill()
			return "", "", 0, err
		}
	}

	// Log backup size and upload speed
	backupSizeMB := float64(n) / (1024 * 1024)
	log.Info(fmt.Sprintf("Backup size: %.2f MB", backupSizeMB))

	return backupName, DateTime, n, nil
}
