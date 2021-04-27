# RadonDB-MySQL
RadonDB-MySQL is a open-source, cloud-native, High-Availability cluster solutions that is based on MySQL.

# Github

https://github.com/radondb/radondb-mysql-kubernetes
# Features
- High availability MySQL database
    - Non-centralized automatic leader election
    - Second level switch leader to follower 
    - Strongly consistent data for cluster switching
- Cluster management
- Monitoring and alerting
- Logs
- Account management

# Introduction

This chart bootstraps a single leader and multiple followers [MySQL](https://MySQL.org) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

# Prerequisites

- Kubernetes 1.10+ with Beta APIs enabled
- PV provisioner support in the underlying infrastructure
- Helm 2.11+ or Helm 3.0-beta3+

# Installing the Chart

To install the chart with the release name `my-release`:

```bash
## For Helm v2
$ helm install . --name my-release

## For Helm v3
$ helm install --name my-release .
```

The command deploys MySQL cluster on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

# Uninstall

To uninstall/delete the `my-release` deployment:

```bash
$ helm delete my-release
```

To delete the pvc:

```
kubectl delete pvc data-my-release-mysql-0
kubectl delete pvc data-my-release-mysql-1
kubectl delete pvc data-my-release-mysql-2
```

The commands remove all the Kubernetes components associated with the chart and deletes the release completely.

# Configuration

The following table lists the configurable parameters of the xenondb chart and their default values.


| Parameter                                    | Description                                                                                       | Default                                     |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| `imagePullPolicy`                            | Image pull policy                                                                                 | `IfNotPresent`                              |
| `fullnameOverride`                           | Custom fullname override for the chart                                                            |                                             |
| `nameOverride`                               | Custom name override for the chart                                                                |                                             |
| `replicaCount`                               | The number of pods                                                                                | `3`                                         |
| `busybox.image`                              | `busybox` image repository.                                                                       | `busybox`                                   |
| `busybox.tag`                                | `busybox` image tag.                                                                              | `1.32`                                      |
| `mysql.image`                                | `mysql` image repository.                                                                         | `xenondb/percona`                           |
| `mysql.tag`                                  | `mysql` image tag.                                                                                | `5.7.33`                                    |
| `mysql.allowEmptyRootPassword`               | If set true, allow a empty root password.                                                         | `true`                                      |
| `mysql.mysqlRootPassword`                    | Password for the `root` user.                                                                     |                                             |
| `mysql.mysqlReplicationPassword`             | Password for the `qc_repl` user.                                                                  | `Repl_123`, random 12 characters if not set |
| `mysql.mysqlUser`                            | Username of new user to create.                                                                   | `qingcloud`                                 |
| `mysql.mysqlPassword`                        | Password for the new user.                                                                        | `Qing@123`, random 12 characters if not set |
| `mysql.mysqlDatabase`                        | Name for new database to create.                                                                  | `qingcloud`                                 |
| `mysql.initTokudb`                           | Install tokudb engine.                                                                            | `false`                                     |
| `mysql.args`                                 | Additional arguments to pass to the MySQL container.                                              | `[]`                                        |
| `mysqlconfigFiles.node.cnf`                  | Mysql configuration file                                                                          | See `values.yaml`                           |
| `mysql.livenessProbe.initialDelaySeconds`    | Delay before mysql liveness probe is initiated                                                    | 30                                          |
| `mysql.livenessProbe.periodSeconds`          | How often to perform the mysql probe                                                              | 10                                          |
| `mysql.livenessProbe.timeoutSeconds`         | When the mysql probe times out                                                                    | 5                                           |
| `mysql.livenessProbe.successThreshold`       | Minimum consecutive successes for the mysql probe to be considered successful after having failed.| 1                                           |
| `mysql.livenessProbe.failureThreshold`       | Minimum consecutive failures for the mysql probe to be considered failed after having succeeded.  | 3                                           |
| `mysql.readinessProbe.initialDelaySeconds`   | Delay before mysql readiness probe is initiated                                                   | 10                                          |
| `mysql.readinessProbe.periodSeconds`         | How often to perform the mysql probe                                                              | 10                                          |
| `mysql.readinessProbe.timeoutSeconds`        | When the mysql probe times out                                                                    | 1                                           |
| `mysql.readinessProbe.successThreshold`      | Minimum consecutive successes for the mysql probe to be considered successful after having failed.| 1                                           |
| `mysql.readinessProbe.failureThreshold`      | Minimum consecutive failures for the mysql probe to be considered failed after having succeeded.  | 3                                           |
| `mysql.extraEnvVars`                         | Additional environment variables as a string to be passed to the `tpl` function                   |                                             |
| `mysql.resources`                            | CPU/Memory resource requests/limits for mysql.                                                    | Memory: `256Mi`, CPU: `100m`                |
| `xenon.image`                                | `xenon` image repository.                                                                         | `xenondb/xenon`                             |
| `xenon.tag`                                  | `xenon` image tag.                                                                                | `1.1.5-alpha`                               |
| `xenon.args`                                 | Additional arguments to pass to the xenon container.                                              | `[]`                                        |
| `xenon.extraEnvVars`                         | Additional environment variables as a string to be passed to the `tpl` function                   |                                             |
| `xenon.livenessProbe.initialDelaySeconds`    | Delay before xenon liveness probe is initiated                                                    | 30                                          |
| `xenon.livenessProbe.periodSeconds`          | How often to perform the xenon probe                                                              | 10                                          |
| `xenon.livenessProbe.timeoutSeconds`         | When the xenon probe times out                                                                    | 5                                           |
| `xenon.livenessProbe.successThreshold`       | Minimum consecutive successes for xenon probe to be considered successful after having failed.    | 1                                           |
| `xenon.livenessProbe.failureThreshold`       | Minimum consecutive failures for the xenon probe to be considered failed after having succeeded.  | 3                                           |
| `xenon.readinessProbe.initialDelaySeconds`   | Delay before xenon readiness probe is initiated                                                   | 10                                          |
| `xenon.readinessProbe.periodSeconds`         | How often to perform the xenon probe                                                              | 10                                          |
| `xenon.readinessProbe.timeoutSeconds`        | When the xenon probe times out                                                                    | 1                                           |
| `xenon.readinessProbe.successThreshold`      | Minimum consecutive successes for xenon probe to be considered successful after having failed.    | 1                                           |
| `xenon.readinessProbe.failureThreshold`      | Minimum consecutive failures for the xenon probe to be considered failed after having succeeded.  | 3                                           |
| `xenon.resources`                            | CPU/Memory resource requests/limits for xenon.                                                    | Memory: `128Mi`, CPU: `50m`                 |
| `metrics.enabled`                            | Start a side-car prometheus exporter                                                              | `true`                                      |
| `metrics.image`                              | Exporter image                                                                                    | `prom/mysqld-exporter`                      |
| `metrics.tag`                                | Exporter image                                                                                    | `v0.12.1`                                   |
| `metrics.annotations`                        | Exporter annotations                                                                              | `{}`                                        |
| `metrics.livenessProbe.initialDelaySeconds`  | Delay before metrics liveness probe is initiated                                                  | 15                                          |
| `metrics.livenessProbe.timeoutSeconds`       | When the probe times out                                                                          | 5                                           |
| `metrics.readinessProbe.initialDelaySeconds` | Delay before metrics readiness probe is initiated                                                 | 5                                           |
| `metrics.readinessProbe.timeoutSeconds`      | When the probe times out                                                                          | 1                                           |
| `metrics.serviceMonitor.enabled`             | Set this to `true` to create ServiceMonitor for Prometheus operator                               | `true`                                      |
| `metrics.serviceMonitor.namespace`           | Optional namespace in which to create ServiceMonitor                                              | `nil`                                       |
| `metrics.serviceMonitor.interval`            | Scrape interval. If not set, the Prometheus default scrape interval is used                       | 10s                                         |
| `metrics.serviceMonitor.scrapeTimeout`       | Scrape timeout. If not set, the Prometheus default scrape timeout is used                         | `nil`                                       |
| `metrics.serviceMonitor.selector`            | Default to kube-prometheus install, but should be set according to Prometheus install             | `{ prometheus: kube-prometheus }`           |
| `slowLogTail`                                | If set to `true` runs a container to tail mysql-slow.log in the pod                               | `true`                                      |
| `resources`                                  | Resource requests/limit                                                                           | Memory: `32Mi`, CPU: `10m`                  |
| `service.annotations`                        | Kubernetes annotations for service                                                                | {}                                          |
| `service.type`                               | Kubernetes service type                                                                           | NodePort                                    |
| `service.loadBalancerIP`                     | The service loadBalancer IP                                                                       | `""`                                        |
| `service.nodePort`                           | The service nodePort                                                                              | `""`                                        |
| `service.clusterIP`                          | The service clusterIP                                                                             | `""`                                        |
| `service.port`                               | The service port                                                                                  | `3306`                                      |
| `rbac.create`                                | If true, create & use RBAC resources                                                              | `true`                                      |
| `serviceAccount.create`                      | Specifies whether a ServiceAccount should be created                                              | `true`                                      |
| `serviceAccount.name`                        | The name of the ServiceAccount to use                                                             |                                             |
| `persistence.enabled`                        | Create a volume to store data                                                                     | true                                        |
| `persistence.storageClass`                   | Type of persistent volume claim                                                                   | nil                                         |
| `persistence.accessMode`                     | Access mode                                                                                       | ReadWriteOnce                               |
| `persistence.size`                           | Size of persistent volume claim                                                                   | 10Gi                                        |
| `persistence.annotations`                    | Persistent Volume annotations                                                                     | {}                                          |
| `priorityClassName`                          | Set pod priorityClassName                                                                         | `{}`                                        |
| `schedulerName`                              | Name of the k8s scheduler (other than default)                                                    | `nil`                                       |
| `statefulsetAnnotations`		                 | Map of annotations for statefulset							                                                   | `{}`			                			             |
| `podAnnotations`                             | Map of annotations to add to the pods                                                             | `{}`                                        |
| `podLabels`                                  | Map of labels to add to the pods                                                                  | `{}`                                        |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
$ cd charts
$ helm install my-release \
  --set mysql.mysqlUser=my-user,mysql.mysqlPassword=my-password,mysql.database=my-database .
```

The above command creates a standard database user named `my-user`, with the password `my-password`, who has access to a database named `my-database`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
$ cd charts
$ helm install my-release -f values.yaml .
```

# Persistence

The [MySQL](https://hub.docker.com/repository/docker/zhyass/percona57) image stores the MySQL data and configurations at the `/var/lib/mysql` path of the container.

By default a PersistentVolumeClaim is created and mounted into the directory. In order to disable this functionality, you can change the `values.yaml` to disable persistence and use an emptyDir instead.

> *"An emptyDir volume is first created when a Pod is assigned to a Node, and exists as long as that Pod is running on that node. When a Pod is removed from a node for any reason, the data in the emptyDir is deleted forever."*

**Notice**: You may need to increase the value of `livenessProbe.initialDelaySeconds` when enabling persistence by using PersistentVolumeClaim from PersistentVolume with varying properties. Since its IO performance has impact on the database initialization performance. The default limit for database initialization is `60` seconds (`livenessProbe.initialDelaySeconds` + `livenessProbe.periodSeconds` * `livenessProbe.failureThreshold`). Once such initialization process takes more time than this limit, kubelet will restart the database container, which will interrupt database initialization then causing persisent data in an unusable state.

# Custom MySQL configuration

You can add or modify the mysql configuration on the `mysql.configFiles`.

```yaml
  configFiles:
    node.cnf: |
      [mysqld]
      default_storage_engine=InnoDB
      max_connections=65535

      # custom mysql configuration.
      expire_logs_days=7
```
