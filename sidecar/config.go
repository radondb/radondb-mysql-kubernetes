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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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
	// The name of the statefulset.
	StatefulSetName string

	// The password of the root user.
	RootPassword string

	// Username of new user to create.
	User string
	// Password for the new user.
	Password string
	// Name for new database to create.
	Database string
	// Name for donor clone
	DonorClone string
	// Name for donor clone
	DonorClonePassword string
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
	// The password of the mysql root user, for operator use only.
	InternalRootPassword string

	// InitTokuDB represents if install tokudb engine.
	InitTokuDB bool

	// MySQLVersion represents the MySQL version that will be run.
	MySQLVersion semver.Version

	// The parameter in xenon means admit defeat count for hearbeat.
	AdmitDefeatHearbeatCount int32
	// The parameter in xenon means election timeout(ms).
	ElectionTimeout int32

	// Whether the MySQL data exists.
	existMySQLData bool
	// for mysql backup
	// backup user and password for http endpoint
	ClusterName string
	// Job name if is backup Job
	JobName string
	// Backup user name to http Server
	BackupUser string

	// Backup Password to htpp Server
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

	// Time Point for Restore
	RestorePoint string
	// Clone flag
	CloneFlag bool

	// GtidPurged is the gtid set of the slave cluster to purged.
	GtidPurged string

	// XRestoreFromNFS string

	// User customized initsql.
	InitSQL string

	// directory in S3 bucket for cluster restore from
	XRestoreFrom      string
	XRestoreFromNFS   string
	XCloudS3EndPoint  string
	XCloudS3AccessKey string
	XCloudS3SecretKey string
	XCloudS3Bucket    string
	XRemoteDateSource string
	// Need Upgrade
	NeedUpgrade bool

	RemoteClusterName      string
	RemoteClusterNamespace string
	// add it in env
	ServerIDStartOffset string
}

// NewInitConfig returns a pointer to Config.
func NewInitConfig() *Config {
	// check mysql version is supported or not and then get parse mysql semver version
	var mysqlSemVer semver.Version

	if ver := getEnvValue("MYSQL_VERSION"); ver == utils.InvalidMySQLVersion {
		panic("invalid mysql version, currently we only support 5.7 or 8.0")
	} else {
		var err error
		// Do not use := here, it will alloc a new semver.Version every time.
		mysqlSemVer, err = semver.Parse(ver)
		if err != nil {
			log.Info("semver get from MYSQL_VERSION is invalid", "semver: ", mysqlSemVer)
			panic(err)
		}
	}

	initTokuDB := false
	if len(getEnvValue("INIT_TOKUDB")) > 0 {
		initTokuDB = true
	}
	needUpgrade := false
	if len(getEnvValue("NEED_UPGRADE")) > 0 {
		needUpgrade = true
	}

	admitDefeatHearbeatCount, err := strconv.ParseInt(getEnvValue("ADMIT_DEFEAT_HEARBEAT_COUNT"), 10, 32)
	if err != nil {
		admitDefeatHearbeatCount = 5
	}
	electionTimeout, err := strconv.ParseInt(getEnvValue("ELECTION_TIMEOUT"), 10, 32)
	if err != nil {
		electionTimeout = 10000
	}

	existMySQLData, _ := checkIfPathExists(fmt.Sprintf("%s/mysql", dataPath))

	return &Config{
		HostName:        getEnvValue("POD_HOSTNAME"),
		NameSpace:       getEnvValue("NAMESPACE"),
		ServiceName:     getEnvValue("SERVICE_NAME"),
		StatefulSetName: getEnvValue("STATEFULSET_NAME"),

		RootPassword:         getEnvValue("MYSQL_ROOT_PASSWORD"),
		InternalRootPassword: getEnvValue("INTERNAL_ROOT_PASSWORD"),

		Database: getEnvValue("MYSQL_DATABASE"),
		User:     getEnvValue("MYSQL_USER"),
		Password: getEnvValue("MYSQL_PASSWORD"),

		ReplicationUser:     getEnvValue("MYSQL_REPL_USER"),
		ReplicationPassword: getEnvValue("MYSQL_REPL_PASSWORD"),

		MetricsUser:     getEnvValue("METRICS_USER"),
		MetricsPassword: getEnvValue("METRICS_PASSWORD"),

		OperatorUser:     getEnvValue("OPERATOR_USER"),
		OperatorPassword: getEnvValue("OPERATOR_PASSWORD"),

		DonorClone:         getEnvValue("DONOR_USER"),
		DonorClonePassword: getEnvValue("DONOR_PASSWORD"),

		InitTokuDB: initTokuDB,

		MySQLVersion: mysqlSemVer,

		AdmitDefeatHearbeatCount: int32(admitDefeatHearbeatCount),
		ElectionTimeout:          int32(electionTimeout),

		existMySQLData:    existMySQLData,
		XRestoreFrom:      getEnvValue("RESTORE_FROM"),
		RestorePoint:      getEnvValue("RESTORE_POINT"),
		XRestoreFromNFS:   getEnvValue("RESTORE_FROM_NFS"),
		XCloudS3EndPoint:  getEnvValue("S3_ENDPOINT"),
		XCloudS3AccessKey: getEnvValue("S3_ACCESSKEY"),
		XCloudS3SecretKey: getEnvValue("S3_SECRETKEY"),
		XCloudS3Bucket:    getEnvValue("S3_BUCKET"),
		XRemoteDateSource: getEnvValue("REMOTESRC"),
		ClusterName:       getEnvValue("CLUSTER_NAME"),
		CloneFlag:         false,
		GtidPurged:        "",
		// need upgrade
		NeedUpgrade:            needUpgrade,
		RemoteClusterName:      getEnvValue("REMOTE_CLUSTER_NAME"),
		RemoteClusterNamespace: getEnvValue("REMOTE_CLUSTER_NAMESPACE"),
		// SERVER_ID_OFFSET
		ServerIDStartOffset: getEnvValue("SERVER_ID_OFFSET"),
	}
}

// NewBackupConfig returns the configuration file needed for backup container.
func NewBackupConfig() *Config {
	return &Config{
		NameSpace:    getEnvValue("NAMESPACE"),
		ServiceName:  getEnvValue("SERVICE_NAME"),
		ClusterName:  getEnvValue("CLUSTER_NAME"),
		RootPassword: getEnvValue("MYSQL_ROOT_PASSWORD"),

		BackupUser:     getEnvValue("BACKUP_USER"),
		BackupPassword: getEnvValue("BACKUP_PASSWORD"),

		// XCloudS3EndPoint:  getEnvValue("S3_ENDPOINT"),
		// XCloudS3AccessKey: getEnvValue("S3_ACCESSKEY"),
		// XCloudS3SecretKey: getEnvValue("S3_SECRETKEY"),
		// XCloudS3Bucket:    getEnvValue("S3_BUCKET"),
	}
}

// NewReqBackupConfig returns the configuration file needed for backup job.
// func NewReqBackupConfig() *Config {
// 	return &Config{
// 		NameSpace:   getEnvValue("NAMESPACE"),
// 		ServiceName: getEnvValue("SERVICE_NAME"),

// 		BackupUser:     getEnvValue("BACKUP_USER"),
// 		BackupPassword: getEnvValue("BACKUP_PASSWORD"),
// 		JobName:        getEnvValue("JOB_NAME"),
// 	}
// }

// GetContainerType returns the CONTAINER_TYPE of the currently running container.
// CONTAINER_TYPE used to mark the container type.
func GetContainerType() string {
	return getEnvValue("CONTAINER_TYPE")
}

// build Xtrabackup arguments
func (cfg *Config) XtrabackupArgs() []string {
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

// Build xbcloud arguments
func (cfg *Config) XCloudArgs(backupName string) []string {
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

func (cfg *Config) XBackupName() (string, string) {
	return utils.BuildBackupName(cfg.ClusterName)
}

// buildExtraConfig build a ini file for mysql.
func (cfg *Config) buildExtraConfig(filePath string) (*ini.File, error) {
	conf := ini.Empty()
	sec := conf.Section("mysqld")
	// convert cfg.SERVER_ID_OFFSET to int
	var offset int
	var err error
	if offset, err = strconv.Atoi(cfg.ServerIDStartOffset); err != nil {
		offset = 0
	}
	startIndex := func() int {
		if offset <= 0 {
			return mysqlServerIDOffset
		}
		return offset
	}()

	ordinal, err := utils.GetOrdinal(cfg.HostName)
	arr := strings.Split(cfg.HostName, "-")
	if len(cfg.RemoteClusterName) > 0 && offset <= 0 {
		log.Info("It has remote cluster server-id start offset  +100")
		startIndex += mysqlServerIDOffsetInc
	}
	if len(arr) >= 3 && arr[len(arr)-2] == "ro" {
		log.Info("It is readonly pod,  server-id start offset  +100")
		startIndex += mysqlServerIDOffsetInc
	}

	if err != nil {
		return nil, err
	}
	if _, err := sec.NewKey("server-id", strconv.Itoa(startIndex+ordinal)); err != nil {
		return nil, err
	}

	if _, err := sec.NewKey("init-file", filePath); err != nil {
		return nil, err
	}

	return conf, nil
}

// buildXenonConf build a config file for xenon.
func (cfg *Config) buildXenonConf() []byte {
	pingTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount
	heartbeatTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount
	requestTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount

	version := "mysql80"
	if cfg.MySQLVersion.Major == 5 {
		version = "mysql57"
	}

	var srcSysVars, replicaSysVars string
	if cfg.InitTokuDB {
		srcSysVars = "tokudb_fsync_log_period=default;sync_binlog=default;innodb_flush_log_at_trx_commit=default"
		replicaSysVars = "tokudb_fsync_log_period=1000;sync_binlog=1000;innodb_flush_log_at_trx_commit=1"
	} else {
		srcSysVars = "sync_binlog=default;innodb_flush_log_at_trx_commit=default"
		replicaSysVars = "sync_binlog=1000;innodb_flush_log_at_trx_commit=1"
	}

	hostName := fmt.Sprintf("%s.%s.%s", cfg.HostName, cfg.ServiceName, cfg.NameSpace)
	// Because go-sql-driver will translate localhost to 127.0.0.1 or ::1, but never set the hostname
	// so the host is set to "127.0.0.1" in config file.
	str := fmt.Sprintf(`{
		"log": {
			"level": "INFO"
		},
		"server": {
			"endpoint": "%s:%d",
			"peer-address": "%s:%d",
			"enable-apis": true
		},
		"replication": {
			"passwd": "%s",
			"user": "%s",
			"gtid-purged": "%s"
		},
		"rpc": {
			"request-timeout": %d
		},
		"mysql": {
			"admit-defeat-ping-count": 3,
			"admin": "root",
			"ping-timeout": %d,
			"passwd": "%s",
			"host": "127.0.0.1",
			"version": "%s",
			"master-sysvars": "%s",
			"slave-sysvars": "%s",
			"port": 3306,
			"monitor-disabled": true
		},
		"raft": {
			"election-timeout": %d,
			"admit-defeat-hearbeat-count": %d,
			"heartbeat-timeout": %d,
			"meta-datadir": "%s",
			"semi-sync-degrade": true,
			"purge-binlog-disabled": true,
			"super-idle": false,
			"leader-start-command": "/xenonchecker leaderStart",
			"leader-stop-command": "/xenonchecker leaderStop"
		}
	}
	`, hostName, utils.XenonPort, hostName, utils.XenonPeerPort, cfg.ReplicationPassword, cfg.ReplicationUser,
		cfg.GtidPurged, requestTimeout,
		pingTimeout, cfg.RootPassword, version, srcSysVars, replicaSysVars, cfg.ElectionTimeout,
		cfg.AdmitDefeatHearbeatCount, heartbeatTimeout, xenonConfigPath)

	return utils.StringToBytes(str)
}

// buildInitSql used to build init.sql. The file run after the mysql init.
func (cfg *Config) buildInitSql(hasInit bool) []byte {
	initSQL, err := os.ReadFile(path.Join(mysqlCMPath, "init.sql"))
	if err != nil {
		log.Info("failed to read /mnt/mysql-cm/init.sql")
	}
	var sql string
	if cfg.MySQLVersion.Major == 5 {
		log.Info("version is 5.7, need not donor userq")
		sql = fmt.Sprintf(`SET @@SESSION.SQL_LOG_BIN=0;
CREATE DATABASE IF NOT EXISTS %s;
DROP user IF EXISTS 'root'@'127.0.0.1';
CREATE USER 'root'@'127.0.0.1' IDENTIFIED BY '%s';
GRANT ALL ON *.* TO 'root'@'127.0.0.1'  with grant option;
DROP user IF EXISTS 'root'@'%%';
CREATE USER 'root'@'%%' IDENTIFIED BY '%s';
GRANT ALL ON *.* TO 'root'@'%%' with grant option;
DROP user IF EXISTS '%s'@'%%';
CREATE USER '%s'@'%%' IDENTIFIED BY '%s';
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO '%s'@'%%';
DROP user IF EXISTS '%s'@'%%';
CREATE USER '%s'@'%%' IDENTIFIED BY '%s';
GRANT SELECT, PROCESS, REPLICATION CLIENT ON *.* TO '%s'@'%%';
DROP user IF EXISTS '%s'@'%%';
CREATE USER '%s'@'%%' IDENTIFIED BY '%s';
GRANT SUPER, PROCESS, RELOAD, CREATE, SELECT ON *.* TO '%s'@'%%';
DROP user IF EXISTS '%s'@'%%';
CREATE USER '%s'@'%%' IDENTIFIED BY '%s';
GRANT ALL ON %s.* TO '%s'@'%%' ;
FLUSH PRIVILEGES;

%s
`,
			cfg.Database, //database
			cfg.RootPassword,
			cfg.InternalRootPassword,
			cfg.ReplicationUser,                          //drop user
			cfg.ReplicationUser, cfg.ReplicationPassword, //create user
			cfg.ReplicationUser, //grant REPLICATION

			cfg.MetricsUser,                      //drop user MetricsUser
			cfg.MetricsUser, cfg.MetricsPassword, //create user
			cfg.MetricsUser, //grant

			cfg.OperatorUser,                       //drop user
			cfg.OperatorUser, cfg.OperatorPassword, //create
			cfg.OperatorUser, //grant

			cfg.User,               //drop user
			cfg.User, cfg.Password, //create user
			cfg.Database, cfg.User, //grant
			initSQL,
		)
	} else {
		sql = fmt.Sprintf(`SET @@SESSION.SQL_LOG_BIN=0;
		CREATE DATABASE IF NOT EXISTS %s;
		DROP user IF EXISTS 'root'@'127.0.0.1';
		CREATE USER 'root'@'127.0.0.1' IDENTIFIED BY '%s';
		GRANT ALL ON *.* TO 'root'@'127.0.0.1'  with grant option;
		DROP user IF EXISTS 'root'@'%%';
		CREATE USER 'root'@'%%' IDENTIFIED BY '%s';
		GRANT ALL ON *.* TO 'root'@'%%' with grant option;
		DROP user IF EXISTS '%s'@'%%';
		CREATE USER '%s'@'%%' IDENTIFIED BY '%s';
		GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO '%s'@'%%';
		DROP user IF EXISTS '%s'@'%%';
		CREATE USER '%s'@'%%' IDENTIFIED BY '%s';
		GRANT SELECT, PROCESS, REPLICATION CLIENT ON *.* TO '%s'@'%%';
		DROP user IF EXISTS '%s'@'%%';
		CREATE USER '%s'@'%%' IDENTIFIED BY '%s';
		GRANT SUPER, PROCESS, RELOAD, CREATE, SELECT ON *.* TO '%s'@'%%';
		DROP user IF EXISTS '%s'@'%%';
		CREATE USER '%s'@'%%' IDENTIFIED BY '%s';
		GRANT ALL ON %s.* TO '%s'@'%%' ;
		DROP user IF EXISTS '%s'@'%%';
		CREATE USER '%s'@'%%' IDENTIFIED BY '%s';
		GRANT BACKUP_ADMIN ON *.* TO '%s'@'%%' ;
		FLUSH PRIVILEGES;
		
		%s
		`,
			cfg.Database, //database
			cfg.RootPassword,
			cfg.InternalRootPassword,
			cfg.ReplicationUser,                          //drop user
			cfg.ReplicationUser, cfg.ReplicationPassword, //create user
			cfg.ReplicationUser, //grant REPLICATION

			cfg.MetricsUser,                      //drop user MetricsUser
			cfg.MetricsUser, cfg.MetricsPassword, //create user
			cfg.MetricsUser, //grant

			cfg.OperatorUser,                       //drop user
			cfg.OperatorUser, cfg.OperatorPassword, //create
			cfg.OperatorUser, //grant

			cfg.User,               //drop user
			cfg.User, cfg.Password, //create user
			cfg.Database, cfg.User, //grant
			cfg.DonorClone,
			cfg.DonorClone, cfg.DonorClonePassword,
			cfg.DonorClone, // grant
			initSQL,
		)
	}

	if hasInit {
		sql += "\nRESET SLAVE ALL;\n"
	}
	return utils.StringToBytes(sql)
}

// buildClientConfig used to build client.conf.
func (cfg *Config) buildClientConfig() (*ini.File, error) {
	conf := ini.Empty()
	sec := conf.Section("client")

	if _, err := sec.NewKey("host", "127.0.0.1"); err != nil {
		return nil, err
	}

	if _, err := sec.NewKey("port", fmt.Sprintf("%d", utils.MysqlPort)); err != nil {
		return nil, err
	}

	if _, err := sec.NewKey("user", cfg.OperatorUser); err != nil {
		return nil, err
	}

	if _, err := sec.NewKey("password", cfg.OperatorPassword); err != nil {
		return nil, err
	}
	if cfg.MySQLVersion.Major == 8 {
		if _, err := sec.NewKey("donor", cfg.DonorClone); err != nil {
			return nil, err
		}
	}

	return conf, nil
}

// // buildLeaderStart build the leader-start.sh.
// func (cfg *Config) buildLeaderStart() []byte {
// 	str := fmt.Sprintf(`#!/usr/bin/env bash
// curl -X PATCH -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" -H "Content-Type: application/json-patch+json" \
// --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/namespaces/%s/pods/$HOSTNAME \
// -d '[{"op": "replace", "path": "/metadata/labels/role", "value": "leader"}]'
// `, cfg.NameSpace)
// 	return utils.StringToBytes(str)
// }

// // buildLeaderStop build the leader-stop.sh.
// func (cfg *Config) buildLeaderStop() []byte {
// 	str := fmt.Sprintf(`#!/usr/bin/env bash
// curl -X PATCH -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" -H "Content-Type: application/json-patch+json" \
// --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/namespaces/%s/pods/$HOSTNAME \
// -d '[{"op": "replace", "path": "/metadata/labels/role", "value": "follower"}]'
// `, cfg.NameSpace)
// 	return utils.StringToBytes(str)
// }

/*
	The function is equivalent to the following shell script template:

#!/bin/sh
if [ ! -d {{.DataDir}} ] ; then

	echo "is not exist the var lib mysql"
	mkdir {{.DataDir}}
	chown -R mysql.mysql {{.DataDir}}

fi
mkdir /root/backup
xbcloud get --storage=S3 \
--s3-endpoint="{{.XCloudS3EndPoint}}" \
--s3-access-key="{{.XCloudS3AccessKey}}" \
--s3-secret-key="{{.XCloudS3SecretKey}}" \
--s3-bucket="{{.XCloudS3Bucket}}" \
--parallel=10 {{.XRestoreFrom}} \
--insecure |xbstream -xv -C /root/backup
# prepare redolog
xtrabackup --defaults-file={{.MyCnfMountPath}} --use-memory=3072M --prepare --apply-log-only --target-dir=/root/backup
# prepare data
xtrabackup --defaults-file={{.MyCnfMountPath}} --use-memory=3072M --prepare --target-dir=/root/backup
chown -R mysql.mysql /root/backup
xtrabackup --defaults-file={{.MyCnfMountPath}} --datadir={{.DataDir}} --copy-back --target-dir=/root/backup
chown -R mysql.mysql {{.DataDir}}
rm -rf /root/backup
*/
func (cfg *Config) executeS3Restore(pathArg string) error {
	if len(cfg.XRestoreFrom) == 0 {
		return fmt.Errorf("do not have restore from")
	}
	if len(cfg.XCloudS3EndPoint) == 0 ||
		len(cfg.XCloudS3AccessKey) == 0 ||
		len(cfg.XCloudS3SecretKey) == 0 ||
		len(cfg.XCloudS3Bucket) == 0 {
		return fmt.Errorf("do not have S3 information")
	}
	// Check has directory, and create it.
	if _, err := os.Stat(utils.DataVolumeMountPath); os.IsNotExist(err) {
		if err := os.MkdirAll(utils.DataVolumeMountPath, 0755); err != nil {
			return fmt.Errorf("failed to create data directory : %s", err)
		}
	}
	// Execute xbcloud get.
	args := []string{
		"get",
		"--storage=S3",
		"--s3-endpoint=" + cfg.XCloudS3EndPoint,
		"--s3-access-key=" + cfg.XCloudS3AccessKey,
		"--s3-secret-key=" + cfg.XCloudS3SecretKey,
		"--s3-bucket=" + cfg.XCloudS3Bucket,
		"--parallel=10",
		cfg.XRestoreFrom,
		"--insecure",
	}
	log.Info(fmt.Sprintf("run args %v", args))
	xcloud := exec.Command(xcloudCommand, args...)                               //nolint
	xbstream := exec.Command("xbstream", "-xv", "-C", utils.DataVolumeMountPath) //nolint
	var err error
	if xbstream.Stdin, err = xcloud.StdoutPipe(); err != nil {
		return fmt.Errorf("failed to xbstream and xcloud piped")
	}
	xbstream.Stderr = os.Stderr
	xcloud.Stderr = os.Stderr
	if err := xcloud.Start(); err != nil {
		return fmt.Errorf("failed to xcloud start : %s", err)
	}
	if err := xbstream.Start(); err != nil {
		return fmt.Errorf("failed to xbstream start : %s", err)
	}
	// Make error channels.
	errCh := make(chan error, 2)
	go func() {
		errCh <- xcloud.Wait()
	}()
	go func() {
		errCh <- xbstream.Wait()
	}()
	// Wait for error.
	for i := 0; i < 2; i++ {
		if err = <-errCh; err != nil {
			return err
		}
	}
	// Get backup gtid
	gtid, _ := GetXtrabackupGTIDPurged("/root/backup/")

	log.Info("get restore gtid:", "gtid", gtid)
	// Xtrabackup prepare and apply-log-only.
	log.Info("Xtrabackup prepare and apply-log-only")
	cmd := exec.Command(xtrabackupCommand, "--defaults-file="+utils.MysqlConfVolumeMountPath+"/my.cnf", "--prepare", "--apply-log-only", "--target-dir="+utils.DataVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to xtrabackup prepare and apply-log-only : %s", err)
	}
	// Xtrabackup prepare.
	log.Info("Xtrabackup prepare")
	cmd = exec.Command(xtrabackupCommand, "--defaults-file="+utils.MysqlConfVolumeMountPath+"/my.cnf", "--prepare", "--target-dir="+utils.DataVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to xtrabackup prepare : %s", err)
	}
	// Do not need to Xtrabackup copy-back to /var/lib/mysql.
	// Execute chown -R mysql.mysql /var/lib/mysql.
	log.Info("chown -R mysql.mysql /var/lib/mysql")
	if err := exec.Command("chown", "-R", "mysql.mysql", utils.DataVolumeMountPath).Run(); err != nil {
		return fmt.Errorf("failed to chown mysql.mysql %s  : %s", utils.DataVolumeMountPath, err)
	}
	// Do not need Xtrabackup copy-back.
	// S3 download

	log.Info("downloading S3 binlog")
	if s3, err := NewS3(strings.TrimPrefix(strings.TrimPrefix(cfg.XCloudS3EndPoint, "https://"), "http://"),
		cfg.XCloudS3AccessKey, cfg.XCloudS3SecretKey, cfg.XCloudS3Bucket,
		strings.HasPrefix(cfg.XCloudS3EndPoint, "https")); err != nil {
		return fmt.Errorf("failed to new s3 : %s", err)
	} else {
		if err = os.Mkdir(path.Join(utils.InitFileVolumeMountPath, buildBinlogDir(cfg.XRestoreFrom)), os.FileMode(0755)); err != nil {
			if !os.IsExist(err) {
				return fmt.Errorf("error mkdir %s: %s", buildBinlogDir(cfg.XRestoreFrom), err)
			}
		}
		if len(cfg.RestorePoint) != 0 {
			restorePoint, err := time.Parse("2006-01-02 15:04:05", cfg.RestorePoint)
			if err != nil {
				log.Info("restore point parse error : %s", err)
			} else {
				s3.S3Download(cfg, buildBinlogDir(cfg.XRestoreFrom))
				cfg.BuildScript(gtid, restorePoint)
			}

		}

	}
	return nil
}

// Build Script in InitFileVolumeMountPath
func (cfg *Config) BuildScript(skipGtid string, restorePoint time.Time) error {
	//restorePoint, err := time.Parse("20060102-150405", string)
	binlogPathDir := path.Join(utils.InitFileVolumeMountPath, buildBinlogDir(cfg.XRestoreFrom))
	script := fmt.Sprintf(`#!/bin/bash
mysqld&
pid="$!"
echo "now to restore binlog"
mysql=( mysql -uroot )
for i in {120..0}; do
			if echo 'SELECT 1' | "${mysql[@]}" &> /dev/null; then
				break
			fi
			echo 'MySQL init process in progress...'
			sleep 1
done
echo "now mysqlstart!"
if [ "$i" = 0 ]; then
			echo >&2 'MySQL init process failed.'
			exit 1
fi
ls %s/mysql-bin*|awk '{print "cat   "$1  "|mysqlbinlog --exclude-gtids=\"%s\" --stop-datetime=\"%s\" -|mysql -uroot"}'|bash
if ! kill -s TERM "$pid" || ! wait "$pid"; then
	echo >&2 'MySQL init process failed.'
	exit 1
fi
`, binlogPathDir, skipGtid, restorePoint.Format(RestoreTimeSample))
	fmt.Println(script)
	if err := ioutil.WriteFile(utils.RadonDBBinDir+"/restore.sh", []byte(script), 0755); err != nil {
		return err
	}

	return nil
}

// Do Restore after clone.
func (cfg *Config) executeCloneRestore() error {
	// Check directory exist, create if not exist.
	if _, err := os.Stat(utils.DataVolumeMountPath); os.IsNotExist(err) {
		os.Mkdir(utils.DataVolumeMountPath, 0755)
		// Empty the directory. just for lost+found.
		dir, err := ioutil.ReadDir(utils.DataVolumeMountPath)
		if err != nil {
			return fmt.Errorf("failed to read datadir %s", err)
		}
		for _, d := range dir {
			os.RemoveAll(path.Join([]string{utils.DataVolumeMountPath, d.Name()}...))
		}
	}

	// Xtrabackup prepare and apply-log-only.
	log.Info("xtrabackup prepare apply-log only")
	cmd := exec.Command(xtrabackupCommand, "--defaults-file="+utils.MysqlConfVolumeMountPath+"/my.cnf", "--use-memory=3072M", "--prepare", "--apply-log-only", "--target-dir="+utils.DataVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to xtrabackup prepare apply-log-only : %s", err)
	}
	// Xtrabackup Prepare.
	log.Info("xtrabackup prepare")
	cmd = exec.Command(xtrabackupCommand, "--defaults-file="+utils.MysqlConfVolumeMountPath+"/my.cnf", "--use-memory=3072M", "--prepare", "--target-dir="+utils.DataVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to xtrabackup prepare : %s", err)
	}
	// Get the backup binlong info.
	gtid, err := GetXtrabackupGTIDPurged(utils.DataVolumeMountPath)
	if err == nil {
		cfg.GtidPurged = gtid
	}
	log.Info("get master gtid purged :", "gtid purged", cfg.GtidPurged)
	// Do not need Xtrabackup copy-back.

	// Remove Relaybin.
	// Because the relaybin is not used in the restore process,
	// we can remove it to prevent it to be used by salve in the future.
	cmd = exec.Command("rm", "-rf", utils.DataVolumeMountPath+"mysql-relay*")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove relay-bin : %s", err)
	}
	// Run chown -R mysql.mysql /var/lib/mysql
	log.Info("Run chown -R mysql.mysql /var/lib/mysql")
	cmd = exec.Command("chown", "-R", "mysql.mysql", utils.DataVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to chown -R mysql.mysql : %s", err)
	}
	log.Info("execute clone restore success")
	return nil
}

// Parse the xtrabackup_binlog_info, the format is filename \t position \t gitid1 \ngitid2 ...
// or filename \t position\n
// Get the gtid when it is existed, or return empty string.
// It used to purged the gtid when start the mysqld slave.
func GetXtrabackupGTIDPurged(backuppath string) (string, error) {
	byteStream, err := ioutil.ReadFile(fmt.Sprintf("%s/xtrabackup_binlog_info", backuppath))
	if err != nil {
		return "", err
	}
	line := strings.TrimSuffix(string(byteStream), "\n")
	ss := strings.Split(line, "\t")
	if len(ss) != 3 {
		return "", fmt.Errorf("info.file.content.invalid[%v]", string(byteStream))
	}
	// Replace multi gtidset \n
	return strings.Replace(ss[2], "\n", "", -1), nil
}

/*
`#!/bin/sh

		if [ ! -d  {{.DataDir}} ]; then
	        echo "is not exist the var lib mysql"
	        mkdir  {{.DataDir}}
	        chown -R mysql.mysql  {{.DataDir}}
	    fi
	    rm -rf  {{.DataDir}}/*
	    xtrabackup --defaults-file={{.MyCnfMountPath}} --use-memory=3072M --prepare --apply-log-only --target-dir=/backup/{{.XRestoreFrom}}
	    xtrabackup --defaults-file={{.MyCnfMountPath}} --use-memory=3072M --prepare --target-dir=/backup/{{.XRestoreFrom}}
	    chown -R mysql.mysql /backup/{{.XRestoreFromNFS}}
	    xtrabackup --defaults-file={{.MyCnfMountPath}} --datadir={{.DataDir}} --copy-back --target-dir=/backup/{{.XRestoreFrom}}
	    exit_code=$?
	    chown -R mysql.mysql {{.DataDir}}
	    exit $exit_code
*/
func (cfg *Config) ExecuteNFSRestore() error {
	if len(cfg.XRestoreFromNFS) == 0 {
		return fmt.Errorf("parameter XRestoreFromNFS empty, do next step")
	}
	if len(cfg.XRestoreFrom) == 0 {
		return fmt.Errorf("xrestore from is empty, do next step")
	}
	// Restore from NFS

	// Check /var/lib/mysql exists or not.
	if _, err := os.Stat(utils.DataVolumeMountPath); os.IsNotExist(err) {
		err = os.MkdirAll(utils.DataVolumeMountPath, 0755)
		if err != nil {
			return fmt.Errorf("create /var/lib/mysql fail : %s", err)
		}
		// Change the owner of /var/lib/mysql
		cmd := exec.Command("chown", "mysql.mysql", utils.DataVolumeMountPath)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to chown -R mysql.mysql : %s", err)
		}
	}
	// Remove the data directory
	cmd := exec.Command("rm", "-rf", utils.DataVolumeMountPath+"/*")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to rm -rf %s : %s", utils.DataVolumeMountPath, err)
	}
	// Prepare the append-only file
	cmd = exec.Command("xtrabackup", "--defaults-file="+utils.MysqlConfVolumeMountPath+"/my.cnf", "--use-memory=3072M", "--prepare", "--apply-log-only", "--target-dir=/backup/"+cfg.XRestoreFrom)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to xtrabackup prepare append-only: %s", err)
	}
	// Prepare the data directory
	cmd = exec.Command("xtrabackup", "--defaults-file="+utils.MysqlConfVolumeMountPath+"/my.cnf", "--use-memory=3072M", "--prepare", "--target-dir=/backup/"+cfg.XRestoreFrom)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to xtrabackup prepare: %s", err)
	}
	// Copy the data directory.
	cmd = exec.Command("xtrabackup", "--defaults-file="+utils.MysqlConfVolumeMountPath+"/my.cnf", "--datadir="+utils.DataVolumeMountPath, "--copy-back", "--target-dir=/backup/"+cfg.XRestoreFrom)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to xtrabackup copy-back: %s", err)
	}
	// Change owner of data directory
	log.Info(fmt.Sprintf("change owner of data directory %s", utils.DataVolumeMountPath))
	cmd = exec.Command("chown", "-R", "mysql.mysql", utils.DataVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to chown -R mysql.mysql : %s", err)
	}

	gtid, _ := GetXtrabackupGTIDPurged(path.Join("/backup/", cfg.XRestoreFrom))

	log.Info("get restore gtid:", "gtid", gtid)

	binDir := "/backup/" + cfg.XRestoreFrom + "bin"
	cmd = exec.Command("cp", "-rf", binDir, utils.InitFileVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy binlog to docker-entrypoint-initdb.d : %s", err)
	}
	cmd = exec.Command("chown", "-R", "mysql.mysql", utils.InitFileVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to chown -R mysql.mysql : %s", err)
	}
	restorePoint, err := time.Parse("2006-01-02 15:04:05", cfg.RestorePoint)
	if err != nil {
		return fmt.Errorf("restore point parse error : %s", err)
	}
	cfg.BuildScript(gtid, restorePoint)
	return nil
}

func (cfg *Config) ExecuteRemoteSource() error {
	// Check /var/lib/mysql exists or not.
	log.Info("now get data from remote source")
	if _, err := os.Stat(utils.DataVolumeMountPath); os.IsNotExist(err) {
		err = os.MkdirAll(utils.DataVolumeMountPath, 0755)
		if err != nil {
			return fmt.Errorf("create /var/lib/mysql fail : %s", err)
		}
		// Change the owner of /var/lib/mysql
		cmd := exec.Command("chown", "mysql.mysql", utils.DataVolumeMountPath)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to chown -R mysql.mysql : %s", err)
		}
	}
	// Remove the data directory
	cmd := exec.Command("rm", "-rf", utils.DataVolumeMountPath+"/*")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to rm -rf %s : %s", utils.DataVolumeMountPath, err)
	}
	//sshpass -p rootpass ssh root@172.16.0.29 xtrabackup --user=root --password='' --backup --stream=xbstream --target-dir=./  > backup.xbstream 2>backup.log
	CmdStr := "sshpass -p rootpass ssh  -o 'StrictHostKeyChecking no' root@`cat /etc/rsrc/host` xtrabackup --user=root --password=`cat /etc/rsrc/passwd` --backup --stream=xbstream --target-dir=/  > backup.xbstream 2>backup.log;"
	CmdStr += fmt.Sprintf("cd %s ; xbstream -x < /backup.xbstream", utils.DataVolumeMountPath)
	cmd = exec.Command("/bin/bash", "-c", "--", CmdStr)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to rm -rf %s : %s", utils.DataVolumeMountPath, err)
	}
	// Prepare the append-only file
	cmd = exec.Command("xtrabackup", "--defaults-file="+utils.MysqlConfVolumeMountPath+"/my.cnf", "--use-memory=3072M", "--prepare", "--apply-log-only", "--target-dir="+utils.DataVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to xtrabackup prepare append-only: %s", err)
	}
	// Prepare the data directory
	cmd = exec.Command("xtrabackup", "--defaults-file="+utils.MysqlConfVolumeMountPath+"/my.cnf", "--use-memory=3072M", "--prepare", "--target-dir="+utils.DataVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to xtrabackup prepare: %s", err)
	}
	cmd = exec.Command("chown", "-R", "mysql.mysql", utils.DataVolumeMountPath)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to chown -R mysql.mysql : %s", err)
	}
	return nil
}
