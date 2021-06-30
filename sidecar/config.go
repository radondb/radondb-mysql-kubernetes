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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/go-ini/ini"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// Config of the sidecar.
type Config struct {
	// The hostname of the pod.
	HostName string
	// The namespace where the pod is in.
	NameSpace string
	// The name of the headless service.
	ServiceName string

	// The password of the root user.
	RootPassword string

	// The name of replication user.
	ReplicationUser string
	// The password of the replication user.
	ReplicationPassword string

	// The name of metrics user.
	MetricsUser string
	// The password of metrics user.
	MetricsPassword string

	// The name of operator user.
	OperatorUser string
	// The password of operator user.
	OperatorPassword string

	// InitTokuDB represents if install tokudb engine.
	InitTokuDB bool

	// MySQLVersion represents the MySQL version that will be run.
	MySQLVersion semver.Version

	// The parameter in xenon means admit defeat count for hearbeat.
	AdmitDefeatHearbeatCount int32
	// The parameter in xenon means election timeout(ms)。
	ElectionTimeout int32
	//for mysql backup
	// backup user and password for http endpoint
	ClusterName    string
	BackupUser     string
	BackupPassword string
	// XbstreamExtraArgs is a list of extra command line arguments to pass to xbstream.
	XbstreamExtraArgs []string

	// XtrabackupExtraArgs is a list of extra command line arguments to pass to xtrabackup.
	XtrabackupExtraArgs []string

	// XtrabackupPrepareExtraArgs is a list of extra command line arguments to pass to xtrabackup
	// during --prepare.
	XtrabackupPrepareExtraArgs []string

	// XtrabackupTargetDir is a backup destination directory for xtrabackup.
	XtrabackupTargetDir string
	XCloudS3EndPoint    string
	XCloudS3AccessKey   string
	XCloudS3SecretKey   string
	XCloudS3Bucket      string
	XRestoreFrom        string
}

// NewConfig returns a pointer to Config.
func NewConfig() *Config {
	mysqlVersion, err := semver.Parse(getEnvValue("MYSQL_VERSION"))
	if err != nil {
		log.Info("MYSQL_VERSION is not a semver version")
		mysqlVersion, _ = semver.Parse(utils.MySQLDefaultVersion)
	}

	initTokuDB := false
	if len(getEnvValue("INIT_TOKUDB")) > 0 {
		initTokuDB = true
	}

	admitDefeatHearbeatCount, err := strconv.ParseInt(getEnvValue("ADMIT_DEFEAT_HEARBEAT_COUNT"), 10, 32)
	if err != nil {
		admitDefeatHearbeatCount = 5
	}
	electionTimeout, err := strconv.ParseInt(getEnvValue("ELECTION_TIMEOUT"), 10, 32)
	if err != nil {
		electionTimeout = 10000
	}

	return &Config{
		HostName:    getEnvValue("POD_HOSTNAME"),
		NameSpace:   getEnvValue("NAMESPACE"),
		ServiceName: getEnvValue("SERVICE_NAME"),

		RootPassword: getEnvValue("MYSQL_ROOT_PASSWORD"),

		ReplicationUser:     getEnvValue("MYSQL_REPL_USER"),
		ReplicationPassword: getEnvValue("MYSQL_REPL_PASSWORD"),

		MetricsUser:     getEnvValue("METRICS_USER"),
		MetricsPassword: getEnvValue("METRICS_PASSWORD"),

		OperatorUser:     getEnvValue("OPERATOR_USER"),
		OperatorPassword: getEnvValue("OPERATOR_PASSWORD"),

		InitTokuDB: initTokuDB,

		MySQLVersion: mysqlVersion,

		AdmitDefeatHearbeatCount: int32(admitDefeatHearbeatCount),
		ElectionTimeout:          int32(electionTimeout),

		ClusterName:                getEnvValue("SERVICE_NAME"),
		BackupUser:                 "sys_backups", //getEnvValue("BACKUP_USER"),
		BackupPassword:             "sys_backups", //getEnvValue("BACKUP_PASSWORD"),
		XbstreamExtraArgs:          strings.Fields(getEnvValue("XBSTREAM_EXTRA_ARGS")),
		XtrabackupExtraArgs:        strings.Fields(getEnvValue("XTRABACKUP_EXTRA_ARGS")),
		XtrabackupPrepareExtraArgs: strings.Fields(getEnvValue("XTRABACKUP_PREPARE_EXTRA_ARGS")),
		XtrabackupTargetDir:        getEnvValue("XTRABACKUP_TARGET_DIR"),
		//TODO
		XCloudS3EndPoint:  getEnvValue("S3_ENDPOINT"),
		XCloudS3AccessKey: getEnvValue("S3_ACCESSKEY"),
		XCloudS3SecretKey: getEnvValue("S3_SECRETKEY"),
		XCloudS3Bucket:    getEnvValue("S3_BUCKET"),
		XRestoreFrom:      getEnvValue("RESTORE_FROM"),
	}
}
func (cfg *Config) XtrabackupArgs() []string {
	// xtrabackup --backup <args> --target-dir=<backup-dir> <extra-args>
	user := "root"
	if len(cfg.ReplicationUser) != 0 {
		user = cfg.ReplicationUser
	}

	tmpdir := "/root/backup/"
	if len(cfg.XtrabackupTargetDir) != 0 {
		tmpdir = cfg.XtrabackupTargetDir
	}
	xtrabackupArgs := []string{
		"--backup",
		"--stream=xbstream",
		"--host=127.0.0.1",
		fmt.Sprintf("--user=%s", user),
		fmt.Sprintf("--target-dir=%s", tmpdir),
	}

	return append(xtrabackupArgs, cfg.XtrabackupExtraArgs...)
}
func (cfg *Config) XCloudArgs() []string {
	cur_time := time.Now()
	xcloudArgs := []string{
		"put",
		"--storage=S3",
		fmt.Sprintf("--s3-endpoint=%s", cfg.XCloudS3EndPoint),
		fmt.Sprintf("--s3-access-key=%s", cfg.XCloudS3AccessKey),
		fmt.Sprintf("--s3-secret-key=%s", cfg.XCloudS3SecretKey),
		fmt.Sprintf("--s3-bucket=%s", cfg.XCloudS3Bucket),
		"--parallel=10",
		fmt.Sprintf("backup_%v%v%v%v%v%v", cur_time.Year(), int(cur_time.Month()),
			cur_time.Day(), cur_time.Hour(), cur_time.Minute(), cur_time.Second()),
		"--insecure",
	}
	return xcloudArgs
}

// buildExtraConfig build a ini file for mysql.
func (cfg *Config) buildExtraConfig(filePath string) (*ini.File, error) {
	conf := ini.Empty()
	sec := conf.Section("mysqld")

	id, err := generateServerID(cfg.HostName)
	if err != nil {
		return nil, err
	}
	if _, err := sec.NewKey("server-id", strconv.Itoa(id)); err != nil {
		return nil, err
	}

	if _, err := sec.NewKey("init-file", filePath); err != nil {
		return nil, err
	}

	return conf, nil
}
