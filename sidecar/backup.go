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

package sidecar

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type BackupClientConfig struct {
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
	// BackupType is a backup type for xtrabackup. s3 or disk
	BackupType BkType `json:"backup_type"`
}

type BkType string

const (
	// BackupTypeS3 is a backup type for xtrabackup. s3
	S3 BkType = "s3"
	// BackupTypeDisk is a backup type for xtrabackup. disk
	NFS BkType = "nfs"
)

// NewReqBackupConfig returns the configuration file needed for backup job call /backup.
// The configuration file is obtained from the environment variables.
func NewReqBackupConfig() *BackupClientConfig {

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
		BackupType:        BkType(getEnvValue("BACKUP_TYPE")),
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
		backupName,
		"--insecure",
	}
	return xcloudArgs
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

func setAnnonations(cfg *BackupClientConfig, backname string, DateTime string, BackupType string, BackupSize int64, Gtid string) error {
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
	job.Annotations[utils.JobAnonationGtid] = Gtid
	_, err = clientset.BatchV1().Jobs(cfg.NameSpace).Update(context.TODO(), job, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func RunTakeS3BackupCommand(cfg *BackupClientConfig) (string, string, int64, string, error) {
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
		return "", "", 0, "", err
	}
	var wg sync.WaitGroup
	// set xtrabackup and xcloud stderr to os.Stderr
	Gtid := ""
	xcloud.Stderr = os.Stderr
	Stderr, err := xtrabackup.StderrPipe()
	if err != nil {
		return "", "", 0, "", err
	}
	// Start xtrabackup and xcloud in separate goroutines
	if err := xtrabackup.Start(); err != nil {
		log.Error(err, "failed to start xtrabackup command")
		return "", "", 0, "", err
	}
	if err := xcloud.Start(); err != nil {
		log.Error(err, "fail start xcloud ")
		return "", "", 0, "", err
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
	//xtrabackup.Stderr = os.Stderr

	scanner := bufio.NewScanner(Stderr)
	//scanner.Split(ScanLinesR)
	wg.Add(1)
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			fmt.Println(text)
			if index := strings.Index(text, "GTID"); index != -1 {
				// Mysql5.7 examples: MySQL binlog position: filename 'mysql-bin.000002', position '588', GTID of the last change '319bd6eb-2ea2-11ed-bf40-7e1ef582b427:1-2'
				// MySQL8.0 no gtid:  MySQL binlog position: filename 'mysql-bin.000025', position '156'
				length := len("GTID of the last change")
				Gtid = strings.Trim(text[index+length:], " '") // trim space and \'
				if len(Gtid) != 0 {
					log.Info("Catch gtid: " + Gtid)
				}

			}
		}
		wg.Done()
	}()

	wg.Wait()
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
			return "", "", 0, "", err
		}
	}

	// Log backup size and upload speed
	backupSizeMB := float64(n) / (1024 * 1024)
	log.Info(fmt.Sprintf("Backup size: %.2f MB", backupSizeMB))

	return backupName, DateTime, n, Gtid, nil
}
