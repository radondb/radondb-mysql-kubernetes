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

var (
	// MySQLDefaultVersion is the version for mysql that should be used
	MySQLDefaultVersion = "5.7.33"

	// MySQLTagsToSemVer maps simple version to semver versions
	MySQLTagsToSemVer = map[string]string{
		"5.7": "5.7.33",
	}

	// MysqlImageVersions is a map of supported mysql version and their image
	MysqlImageVersions = map[string]string{
		"5.7.33": "percona/percona-server:5.7.33",
	}
)

const (
	// init containers
	ContainerInitSidecarName = "init-sidecar"
	ContainerInitMysqlName   = "init-mysql"

	// containers
	ContainerMysqlName    = "mysql"
	ContainerXenonName    = "xenon"
	ContainerMetricsName  = "metrics"
	ContainerSlowLogName  = "slowlog"
	ContainerAuditLogName = "auditlog"
	ContainerBackupName   = "backup"

	// BackupPort
	XBackupPortName = "xtrabackup"
	XBackupPort     = 8082

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
	ReplicationUser = "qc_repl"
	// The name of the MySQL metrics user.
	MetricsUser = "qc_metrics"
	// The MySQL user used for operator to connect to the mysql node for configuration.
	OperatorUser = "qc_operator"

	// volumes names.
	ConfVolumeName     = "conf"
	ConfMapVolumeName  = "config-map"
	LogsVolumeName     = "logs"
	DataVolumeName     = "data"
	SysVolumeName      = "host-sys"
	ScriptsVolumeName  = "scripts"
	XenonVolumeName    = "xenon"
	InitFileVolumeName = "init-mysql"

	// volumes mount path.
	MyCnfMountPath      = "/etc/mysql/my.cnf"
	ConfVolumeMountPath = "/etc/mysql/conf.d"
	XtrabackupPV        = "backup"

	// volumes mount path.

	ConfMapVolumeMountPath  = "/mnt/config-map"
	LogsVolumeMountPath     = "/var/log/mysql"
	DataVolumeMountPath     = "/var/lib/mysql"
	SysVolumeMountPath      = "/host-sys"
	ScriptsVolumeMountPath  = "/scripts"
	XenonVolumeMountPath    = "/etc/xenon"
	InitFileVolumeMountPath = "/docker-entrypoint-initdb.d"
	SideCarImage            = "acekingke/sidecar:v01"

	XtrabckupLocal = "/backup"
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
	// Secret is the name of the secret that contains operator related credentials.
	Secret ResourceName = "secret"

	// Role is the alias of the role resource.
	Role ResourceName = "role"
	// RoleBinding is the alias of the rolebinding resource.
	RoleBinding ResourceName = "rolebinding"
	// ServiceAccount is the alias of the serviceaccount resource.
	ServiceAccount ResourceName = "service-account"
)
