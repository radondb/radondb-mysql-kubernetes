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

package utils

import "net/http"

var (
	// MySQLDefaultVersion is the version for mysql that should be used
	MySQLDefaultVersion = "5.7.34"

	// InvalidMySQLVersion is used for set invalid version that we do not support.
	InvalidMySQLVersion = "0.0.0"

	// MySQLTagsToSemVer maps simple version to semver versions
	MySQLTagsToSemVer = map[string]string{
		"5.7": "5.7.34",
		"8.0": "8.0.25",
	}

	// MysqlImageVersions is a map of supported mysql version and their image
	MysqlImageVersions = map[string]string{
		"5.7.33": "percona/percona-server:5.7.33",
		"5.7.34": "percona/percona-server:5.7.34",
		"8.0.25": "percona/percona-server:8.0.25",
		"0.0.0":  "errimage",
	}

	// XenonHttpUrls saves the xenon http url and its corresponding request type.
	XenonHttpUrls = map[XenonHttpUrl]string{
		RaftStatus:      http.MethodGet,
		RaftTryToLeader: http.MethodPost,
		XenonPing:       http.MethodGet,
		ClusterAdd:      http.MethodPost,
		ClusterRemove:   http.MethodPost,
	}
)

const (
	// init containers
	ContainerInitSidecarName = "init-sidecar"
	ContainerInitMysqlName   = "init-mysql"

	// containers
	ContainerMysqlName     = "mysql"
	ContainerXenonName     = "xenon"
	ContainerMetricsName   = "metrics"
	ContainerSlowLogName   = "slowlog"
	ContainerAuditLogName  = "auditlog"
	ContainerBackupName    = "backup"
	ContainerBackupJobName = "backup-job"

	// xtrabackup
	XBackupPortName = "xtrabackup"
	XBackupPort     = 8082
	XtrabackupPV    = "backup"
	XtrabckupLocal  = "/backup"

	// MySQL port.
	MysqlPortName = "mysql"
	MysqlPort     = 3306

	// Metrics port.
	MetricsPortName = "metrics"
	MetricsPort     = 9104

	// Xenon port.
	XenonPortName = "xenon"
	XenonPort     = 8801
	XenonPeerPort = 6601

	// The name of the MySQL replication user.
	ReplicationUser = "radondb_repl"
	// The name of the MySQL metrics user.
	MetricsUser = "radondb_metrics"
	// The MySQL user used for operator to connect to the mysql node for configuration.
	OperatorUser = "radondb_operator"
	// The name of the MySQL root user.
	RootUser = "root"

	// xtrabackup http server user
	BackupUser = "sys_backup"

	// volumes names.
	MysqlConfVolumeName = "mysql-conf"
	MysqlCMVolumeName   = "mysql-cm"
	XenonMetaVolumeName = "xenon-meta"
	XenonCMVolumeName   = "xenon-cm"
	LogsVolumeName      = "logs"
	DataVolumeName      = "data"
	SysVolumeName       = "host-sys"
	ScriptsVolumeName   = "scripts"
	XenonConfVolumeName = "xenon-conf"
	InitFileVolumeName  = "init-mysql"

	// volumes mount path.
	MysqlConfVolumeMountPath = "/etc/mysql"
	MysqlCMVolumeMountPath   = "/mnt/mysql-cm"
	XenonMetaVolumeMountPath = "/var/lib/xenon"
	XenonCMVolumeMountPath   = "/mnt/xenon-cm"
	LogsVolumeMountPath      = "/var/log/mysql"
	DataVolumeMountPath      = "/var/lib/mysql"
	SysVolumeMountPath       = "/host-sys"
	ScriptsVolumeMountPath   = "/scripts"
	XenonConfVolumeMountPath = "/etc/xenon"
	InitFileVolumeMountPath  = "/docker-entrypoint-initdb.d"

	// Volume timezone name.
	SysLocalTimeZone = "localtime"

	// Volume host path for time zone.
	SysLocalTimeZoneHostPath = "/etc/localtime"

	// Volume mount path for time zone.
	SysLocalTimeZoneMountPath = "/etc/localtime"

	// The path to the client MySQL client configuration.
	// The file used to liveness and readiness check.
	ConfClientPath = "/etc/mysql/client.conf"

	// preUpdate file
	FileIndicateUpdate = "PreUpdating"

	// LeaderHost is the alias for leader`s host.
	LeaderHost = "leader-host"

	// PluginConfigs is the alias for mysql plugin config.
	PluginConfigs = "plugin.cnf"
)

// ResourceName is the type for aliasing resources that will be created.
type ResourceName string

const (
	// HeadlessSVC is the alias of the headless service resource.
	HeadlessSVC ResourceName = "headless"
	// StatefulSet is the alias of the statefulset resource.
	StatefulSet ResourceName = "mysql"
	// ConfigMap is the alias for mysql configs, the config map resource.
	ConfigMap ResourceName = "config-files"
	// LeaderService is the name of the service that points to leader node.
	LeaderService ResourceName = "leader-service"
	// FollowerService is the name of a service that points healthy followers (excludes leader).
	FollowerService ResourceName = "follower-service"
	// MetricsService is the name of the metrics service that points to all nodes.
	MetricsService ResourceName = "metrics-service"
	// Secret is the name of the secret that contains operator related credentials.
	Secret ResourceName = "secret"
	// Role is the alias of the role resource.
	Role ResourceName = "role"
	// RoleBinding is the alias of the rolebinding resource.
	RoleBinding ResourceName = "rolebinding"
	// ServiceAccount is the alias of the serviceaccount resource.
	ServiceAccount ResourceName = "service-account"
	// PodDisruptionBudget is the name of pod disruption budget for the statefulset.
	PodDisruptionBudget ResourceName = "pdb"
	// XenonMetaData is the name of the configmap that contains xenon metadata.
	XenonMetaData ResourceName = "xenon-metadata"
)

// JobType
const BackupJobTypeName = ContainerBackupName

// RaftRole is the role of the node in raft.
type RaftRole string

const (
	Leader    RaftRole = "LEADER"
	Follower  RaftRole = "FOLLOWER"
	Candidate RaftRole = "CANDIDATE"
	Unknown   RaftRole = "UNKNOWN"
)

const LableRebuild = "rebuild"

// XenonHttpUrl is a http url corresponding to the xenon instruction.
type XenonHttpUrl string

const (
	RaftStatus      XenonHttpUrl = "/v1/raft/status"
	XenonPing       XenonHttpUrl = "/v1/xenon/ping"
	ClusterAdd      XenonHttpUrl = "/v1/cluster/add"
	ClusterRemove   XenonHttpUrl = "/v1/cluster/remove"
	RaftTryToLeader XenonHttpUrl = "/v1/raft/trytoleader"
)
