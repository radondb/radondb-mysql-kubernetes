
### Custom Resources


### Sub Resources

* [BackupOpts](#backupopts)
* [ClusterCondition](#clustercondition)
* [DataSource](#datasource)
* [ExporterSpec](#exporterspec)
* [LogOpts](#logopts)
* [MonitoringSpec](#monitoringspec)
* [MySQLConfigs](#mysqlconfigs)
* [MySQLStandbySpec](#mysqlstandbyspec)
* [MysqlCluster](#mysqlcluster)
* [MysqlClusterList](#mysqlclusterlist)
* [MysqlClusterSpec](#mysqlclusterspec)
* [MysqlClusterStatus](#mysqlclusterstatus)
* [NFSBackupDataSource](#nfsbackupdatasource)
* [NodeCondition](#nodecondition)
* [NodeStatus](#nodestatus)
* [RaftStatus](#raftstatus)
* [ReadOnlyType](#readonlytype)
* [RemoteDataSource](#remotedatasource)
* [RoStatus](#rostatus)
* [S3BackupDataSource](#s3backupdatasource)
* [ServiceSpec](#servicespec)
* [XenonOpts](#xenonopts)

#### BackupOpts



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| image | Image is the image of backup container. | string | false |
| resources | Changing this value causes MySQL More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers | corev1.ResourceRequirements | false |

[Back to Custom Resources](#custom-resources)

#### ClusterCondition

ClusterCondition defines type for cluster conditions.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type of cluster condition, values in (\"Initializing\", \"Ready\", \"Error\"). | ClusterConditionType | true |
| status | Status of the condition, one of (\"True\", \"False\", \"Unknown\"). | [corev1.ConditionStatus](https://pkg.go.dev/k8s.io/api/core/v1#ConditionStatus) | true |
| lastTransitionTime | The last time this Condition type changed. | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Time) | true |
| reason | One word, camel-case reason for current status of the condition. | string | false |
| message | Full text reason for current status of the condition. | string | false |

[Back to Custom Resources](#custom-resources)

#### DataSource



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| remote | Bootstraping from remote data source | [RemoteDataSource](#remotedatasource) | false |
| S3backup | Bootstraping from backup | [S3BackupDataSource](#s3backupdatasource) | false |
| Nfsbackup | restore from nfs | *[NFSBackupDataSource](#nfsbackupdatasource) | false |
| restorePoint | RestorePoint is the target date and time to restore data. The format is \"2006-01-02 15:04:05\" | string | true |

[Back to Custom Resources](#custom-resources)

#### ExporterSpec



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| customTLSSecret | Projected secret containing custom TLS certificates to encrypt output from the exporter web server | *corev1.SecretProjection | false |
| image | To specify the image that will be used for metrics container. | string | false |
| resources | Changing this value causes MySQL and the exporter to restart. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers | corev1.ResourceRequirements | false |
| enabled | enabled is used to enable/disable the exporter. | bool | false |

[Back to Custom Resources](#custom-resources)

#### LogOpts



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| image | To specify the image that will be used for log container. The busybox image. | string | false |
| slowLogTail | SlowLogTail represents if tail the mysql slow log. | bool | false |
| auditLogTail | AuditLogTail represents if tail the mysql audit log. | bool | false |
| errorLogTail | ErrorLogTail represents if tail the mysql error log. | bool | false |
| resources | Log container resources of a MySQL container. | corev1.ResourceRequirements | false |

[Back to Custom Resources](#custom-resources)

#### MonitoringSpec



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| exporter |  | [ExporterSpec](#exporterspec) | false |

[Back to Custom Resources](#custom-resources)

#### MySQLConfigs



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| configMapName | Name of the `ConfigMap` containing MySQL config. | string | false |
| myCnf | A map[string]string that will be passed to my.cnf file. The key/value pairs is persisted in the configmap. | map[string]string | false |
| pluginCnf |  | map[string]string | false |

[Back to Custom Resources](#custom-resources)

#### MySQLStandbySpec



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| enabled | Whether or not the MySQL cluster should be read-only. When this is true, the cluster will be read-only. When this is false, the cluster will run as writable. | bool | true |
| clusterName | The name of the MySQL cluster to follow for binlog. | string | false |
| host | Network address of the MySQL server to follow via via binlog replication. | string | false |
| port | Network port of the MySQL server to follow via binlog replication. | *int32 | false |

[Back to Custom Resources](#custom-resources)

#### MysqlCluster

MysqlCluster is the Schema for the mysqlclusters API

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ObjectMeta) | false |
| spec |  | [MysqlClusterSpec](#mysqlclusterspec) | false |
| status |  | [MysqlClusterStatus](#mysqlclusterstatus) | false |

[Back to Custom Resources](#custom-resources)

#### MysqlClusterList

MysqlClusterList contains a list of MysqlCluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ListMeta) | false |
| items |  | [][MysqlCluster](#mysqlcluster) | true |

[Back to Custom Resources](#custom-resources)

#### MysqlClusterSpec

MysqlClusterSpec defines the desired state of MysqlCluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| replicas | Replicas is the number of pods. | *int32 | false |
| readonlys | Readonlys Info. | *[ReadOnlyType](#readonlytype) | false |
| lag | Lagged | *int32 | false |
| user | Username of new user to create. Only be a combination of letters, numbers or underlines. The length can not exceed 26 characters. | string | false |
| mysqlConfig | MySQLConfig `ConfigMap` name of MySQL config. | [MySQLConfigs](#mysqlconfigs) | false |
| resources | Compute resources of a MySQL container. | corev1.ResourceRequirements | false |
| customTLSSecret | Containing CA (ca.crt) and server cert (tls.crt), server private key (tls.key) for SSL | *corev1.SecretProjection | false |
| storage | Defines a PersistentVolumeClaim for MySQL data. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes | [corev1.PersistentVolumeClaimSpec](https://pkg.go.dev/k8s.io/api/core/v1#PersistentVolumeClaimSpec) | true |
| mysqlVersion | Represents the MySQL version that will be run. The available version can be found here: This field should be set even if the Image is set to let the operator know which mysql version is running. Based on this version the operator can take decisions which features can be used. | string | false |
| xenonOpts | XenonOpts is the options of xenon container. | [XenonOpts](#xenonopts) | false |
| backupOpts | Backup is the options of backup container. | [BackupOpts](#backupopts) | false |
| monitoringSpec | Monitoring is the options of metrics container. | [MonitoringSpec](#monitoringspec) | false |
| image | Specifies mysql image to use. | string | false |
| maxLagTime | MaxLagSeconds configures the readiness probe of mysqld container if the replication lag is greater than MaxLagSeconds, the mysqld container will not be not healthy. | int | false |
| imagePullPolicy | ImagePullPolicy is used to determine when Kubernetes will attempt to pull (download) container images. More info: https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy | corev1.PullPolicy | false |
| tolerations | Tolerations of a MySQL pod. Changing this value causes MySQL to restart. More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration | []corev1.Toleration | false |
| affinity | Scheduling constraints of MySQL pod. Changing this value causes MySQL to restart. More info: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node | *corev1.Affinity | false |
| priorityClassName | Priority class name for the MySQL pods. Changing this value causes MySQL to restart. More info: https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/ | string | false |
| minAvailable | The number of pods from that set that must still be available after the eviction, even in the absence of the evicted pod | string | false |
| dataSource | Specifies a data source for bootstrapping the MySQL cluster. | [DataSource](#datasource) | false |
| standby | Run this cluster as a read-only copy of an existing cluster or archive. | *[MySQLStandbySpec](#mysqlstandbyspec) | false |
| enableAutoRebuild | If true, when the data is inconsistent, Xenon will automatically rebuild the invalid node. | bool | false |
| logOpts | LogOpts is the options of log settings. | [LogOpts](#logopts) | false |
| service | Specification of the service that exposes the MySQL leader instance. | *[ServiceSpec](#servicespec) | false |

[Back to Custom Resources](#custom-resources)

#### MysqlClusterStatus

MysqlClusterStatus defines the observed state of MysqlCluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| readyNodes | ReadyNodes represents number of the nodes that are in ready state. | int | false |
| state | State | ClusterState | false |
| lastbackup | LastBackup | string | false |
| lastbackupGtid |  | string | false |
| lastBackupTime | LastBackup Create time, just for filter | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Time) | false |
| conditions | Conditions contains the list of the cluster conditions fulfilled. | [][ClusterCondition](#clustercondition) | false |
| nodes | Nodes contains the list of the node status fulfilled. | [][NodeStatus](#nodestatus) | false |

[Back to Custom Resources](#custom-resources)

#### NFSBackupDataSource



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Backup name | string | true |
| volume | Secret name | corev1.NFSVolumeSource | false |

[Back to Custom Resources](#custom-resources)

#### NodeCondition

NodeCondition defines type for representing node conditions.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type of the node condition. | NodeConditionType | true |
| status | Status of the node, one of (\"True\", \"False\", \"Unknown\"). | [corev1.ConditionStatus](https://pkg.go.dev/k8s.io/api/core/v1#ConditionStatus) | true |
| lastTransitionTime | The last time this Condition type changed. | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Time) | true |

[Back to Custom Resources](#custom-resources)

#### NodeStatus

NodeStatus defines type for status of a node into cluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of the node. | string | true |
| message | Full text reason for current status of the node. | string | false |
| raftStatus | RaftStatus is the raft status of the node. | [RaftStatus](#raftstatus) | false |
| roStatus | (RO) ReadOnly Status | *[RoStatus](#rostatus) | false |
| conditions | Conditions contains the list of the node conditions fulfilled. | [][NodeCondition](#nodecondition) | false |

[Back to Custom Resources](#custom-resources)

#### RaftStatus



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| role | Role is one of (LEADER/CANDIDATE/FOLLOWER/IDLE/INVALID) | string | false |
| leader | Leader is the name of the Leader of the current node. | string | false |
| nodes | Nodes is a list of nodes that can be identified by the current node. | []string | false |

[Back to Custom Resources](#custom-resources)

#### ReadOnlyType

ReadOnly define the ReadOnly pods

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| num | ReadOnlys is the number of readonly pods. | int32 | true |
| hostname | When the host name is empty, use the leader to change master | string | true |
| resources | The compute resource requirements. | *corev1.ResourceRequirements | false |
| affinity |  | *corev1.Affinity | false |
| tolerations |  | []corev1.Toleration | false |

[Back to Custom Resources](#custom-resources)

#### RemoteDataSource



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| sourceConfig |  | *corev1.SecretProjection | false |

[Back to Custom Resources](#custom-resources)

#### RoStatus

(RO) node status

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| readOnlyReady |  | bool | false |
| Replication |  | bool | false |
| master |  | string | false |

[Back to Custom Resources](#custom-resources)

#### S3BackupDataSource



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Backup name | string | true |
| secretName | Secret name | string | true |

[Back to Custom Resources](#custom-resources)

#### ServiceSpec



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| nodePort | The port on which this service is exposed when type is NodePort or LoadBalancer. Value must be in-range and not in use or the operation will fail. If unspecified, a port will be allocated if this Service requires one. - https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport | *int32 | false |
| type | More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types | string | true |

[Back to Custom Resources](#custom-resources)

#### XenonOpts



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| image | To specify the image that will be used for xenon container. | string | false |
| admitDefeatHearbeatCount | High available component admit defeat heartbeat count. | *int32 | false |
| electionTimeout | High available component election timeout. The unit is millisecond. | *int32 | false |
| enableAutoRebuild | If true, when the data is inconsistent, Xenon will automatically rebuild the invalid node. | bool | false |
| resources | The compute resource requirements. | corev1.ResourceRequirements | false |

[Back to Custom Resources](#custom-resources)
