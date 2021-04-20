# **在 Kubernetes 上部署 Xenondb 集群**

## **简介**

XenonDB 是基于 MySQL 的开源、高可用、云原生集群解决方案。通过使用 Raft 协议，XenonDB 可以快速进行故障转移，且不会丢失任何事务。

本教程演示如何使用命令行在 Kubernetes 上部署 XenonDB。

## **部署步骤**

### **步骤 1：拉取 XenonDB Chart**

使用 git 拉取 XenonDB chart。

```bash
git clone https://github.com/andyli029/xenondb-helm.git
```

> Chart 代表 [Helm](https://helm.sh/zh/docs/intro/using_helm/) 包，包含在 Kubernetes 集群内部运行应用程序、工具或服务所需的所有资源定义。

### **步骤 2：部署**

- **默认部署**

  需指定 release 名称，以下为指定 release 名为 `my-release` 的默认部署指令。
  
  > 说明：release 是运行在 Kubernetes 集群中的 Chart 的实例。

  ```bash
  ## For Helm v2
    cd charts
    helm install . my-release

  ## For Helm v3
    cd charts
    helm install my-release .
  ```

- **指定参数部署**

  在 `helm install` 时使用 `--set key=value[,key=value]` 指令。以下指令创建了一个用户名为 `my-user` ，密码为 `my-password` 的标准数据库用户，可访问名为 `my-database` 的数据库。

  ```bash
  cd charts
  helm install my-release \
  --set mysql.mysqlUser=my-user,mysql.mysqlPassword=my-password,mysql.database=my-database .
  ```

  通过 value.yaml 文件，在安装时配置指定参数，如下指令。更多安装过程中可配置的参数，请参考 [配置](#配置) 。

  ```bash
  cd charts
  helm install my-release -f values.yaml .
  ```

### **步骤 3：部署校验**

部署指令执行完成后，查看 XenonDB 有状态副本集，pod 状态及服务。可查看到相关信息，则 XenonDB 部署成功。

```bash
kubectl get statefulset,pod,svc
```

## **连接 XenonDB**

您需要准备一个用于连接 XenonDB 的客户端。

### **客户端和 XenonDB 在同一项目中**

当客户端和 Xenondb 集群在同一个项目中时，主节点 Host 可以使用 `<release名称>-xenondb-master` 代替，从节点 Host 可以使用 `<release名称>-xenondb-master` 代替。

连接主节点。

```bash
mysql -h <release名称>-xenondb-master -u <用户名> -p
```

连接从节点。

```bash
mysql -h <release名称>-xenondb-slave -u <用户名> -p
```

### **客户端和 XenonDB 不在同一项目中**

查看 pod，服务。

```bash
kubectl get pod,svc
```

查看节点地址。

```bash
kubectl describe <pod名称>
```

查看节点端口。

```bash
kubectl describe <服务名称>
```

连接节点。

```bash
mysql -p <节点地址> -u <用户名> -P <节点端口> -p
```

> 说明：使用外网主机连接可能会出现 `SSL connection error`，需要在加上 `--ssl-mode=DISABLE` 参数，关闭 SSL。

## **配置**

下表列出了 Xenondb Chart 的配置参数及对应的默认值。

| 参数                                   | 描述                                                                                             |  默认值                                          |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------ | ----------------------------------------------- |
| `imagePullPolicy`                            | 镜像拉取策略                                                                                       | `IfNotPresent`                                  |
| `fullnameOverride`                           | 自定义全名覆盖                                                                                      |                                  -               |
| `nameOverride`                               | 自定义名称覆盖                                                                                      |                          -                       |
| `replicaCount`                               | Pod 数目                                                                                          | `3`                                             |
| `busybox.image`                              | `busybox` 镜像库地址                                                                               | `busybox`                                       |
| `busybox.tag`                                | `busybox` 镜像标签                                                                                 | `1.32`                                          |
| `mysql.image`                                | `mysql` 镜像库地址                                                                                 | `zhyass/percona57`                              |
| `mysql.tag`                                  | `mysql` 镜像标签                                                                                   | `beta0.1.0`                                     |
| `mysql.mysqlReplicationPassword`             | `qc_repl` 用户密码                                                                                 | `Repl_123`, 如果没有设置则随机12字符                |
| `mysql.mysqlUser`                            | 新建用户的用户名                                                                                    | `qingcloud`                                      |
| `mysql.mysqlPassword`                        | 新建用户的密码                                                                                      | `Qing@123`, 如果没有设置则随机12字符                |
| `mysql.mysqlDatabase`                        | 将要创建的数据库名                                                                                   | `qingcloud`                                     |
| `mysql.initTokudb`                           | 安装 tokudb 引擎                                                                                   | `false`                                         |
| `mysql.uuid`                                 | mysql 的 Server_uuid                                                                              | 由 `uuidv4` 函数生成                             |
| `mysql.args`                                 | 要传递到 mysql 容器的其他参数                                                                         | `[]`                                            |
| `mysql.livenessProbe.initialDelaySeconds`    | Pod 启动后首次进行存活检查的等待时间                                                                    | 30                                              |
| `mysql.livenessProbe.periodSeconds`          | 存活检查的间隔时间                                                                                    | 10                                              |
| `mysql.livenessProbe.timeoutSeconds`         | 存活探针执行检测请求后，等待响应的超时时间                                                                | 5                                               |
| `mysql.livenessProbe.successThreshold`       | 存活探针检测失败后认为成功的最小连接成功次数                                                              | 1                                               |
| `mysql.livenessProbe.failureThreshold`       | 存活探测失败的重试次数，重试一定次数后将认为容器不健康                                                      | 3                                               |
| `mysql.readinessProbe.initialDelaySeconds`   | Pod 启动后首次进行就绪检查的等待时间                                                                    | 10                                              |
| `mysql.readinessProbe.periodSeconds`         | 就绪检查的间隔时间                                                                                    | 10                                              |
| `mysql.readinessProbe.timeoutSeconds`        | 就绪探针执行检测请求后，等待响应的超时时间                                                                | 1                                               |
| `mysql.readinessProbe.successThreshold`      | 就绪探针检测失败后认为成功的最小连接成功次数                                                               | 1                                               |
| `mysql.readinessProbe.failureThreshold`      | 就绪探测失败的重试次数，重试一定次数后将认为容器未就绪                                                       | 3                                               |
| `mysql.extraEnvVars`                         | 其他作为字符串传递给 `tpl` 函数的环境变量                                                                |         -                                        |
| `mysql.resources`                            | `MySQL` 的资源请求/限制                                                                               | 内存: `256Mi`, CPU: `100m`                    |
| `xenondb.image`                              | `xenondb` 镜像库地址                                                                                 | `zhyass/xenondb`                                |
| `xenondb.tag`                                | `xenondb` 镜像标签                                                                                   | `beta0.1.0`                                     |
| `xenondb.args`                               | 要传递到 xenondb 容器的其他参数                                                                        | `[]`                                            |
| `xenondb.extraEnvVars`                       | 其他作为字符串传递给 `tpl` 函数的环境变量                                                                 |                        -                         |
| `xenondb.livenessProbe.initialDelaySeconds`  | Pod 启动后首次进行存活检查的等待时间                                                                      | 30                                              |
| `xenondb.livenessProbe.periodSeconds`        | 存活检查的间隔时间                                                                                     | 10                                              |
| `xenondb.livenessProbe.timeoutSeconds`       | 存活探针执行检测请求后，等待响应的超时时间                                                                 | 5                                               |
| `xenondb.livenessProbe.successThreshold`     | 存活探针检测失败后认为成功的最小连接成功次数                                                               | 1                                               |
| `xenondb.livenessProbe.failureThreshold`     | 存活探测失败的重试次数，重试一定次数后将认为容器不健康                                                       | 3                                               |
| `xenondb.readinessProbe.initialDelaySeconds` | Pod 启动后首次进行就绪检查的等待时间                                                                     | 10                                              |
| `xenondb.readinessProbe.periodSeconds`       | 就绪检查的间隔时间                                                                                     | 10                                              |
| `xenondb.readinessProbe.timeoutSeconds`      | 就绪探针执行检测请求后，等待响应的超时时间                                                                 | 1                                               |
| `xenondb.readinessProbe.successThreshold`    | 就绪探针检测失败后认为成功的最小连接成功次数                                                                | 1                                               |
| `xenondb.readinessProbe.failureThreshold`    | 就绪探测失败的重试次数，重试一定次数后将认为容器未就绪                                                       | 3                                               |
| `xenondb.resources`                          | `xenondb` 的资源请求/限制                                                                             | 内存: `128Mi`, CPU: `50m`                     |
| `metrics.enabled`                            | 以 side-car 模式开启 Prometheus Exporter                                                              | `true`                                          |
| `metrics.image`                              | Exporter 镜像地址                                                                                     | `prom/mysqld-exporter`                          |
| `metrics.tag`                                | Exporter 标签                                                                                        | `v0.12.1`                                       |
| `metrics.annotations`                        | Exporter 注释                                                                                        | `{}`                                            |
| `metrics.livenessProbe.initialDelaySeconds`  | Pod 启动后首次进行存活检查的等待时间                                                                     | 15                                              |
| `metrics.livenessProbe.timeoutSeconds`       | 存活探针执行检测请求后，等待响应的超时时间                                                                 | 5                                               |
| `metrics.readinessProbe.initialDelaySeconds` | Pod 启动后首次进行就绪检查的等待时间                                                                     | 5                                               |
| `metrics.readinessProbe.timeoutSeconds`      | 就绪探针执行检测请求后，等待响应的超时时间                                                                 | 1                                               |
| `metrics.resources`                          | Exporter 资源 请求/限制                                                                               | 内存: `32Mi`, CPU: `10m`                      |
| `service.annotations`                        | Kubernetes 服务注释                                                                                  | {}                                              |
| `service.type`                               | Kubernetes 服务类型                                                                                  | NodePort                                        |
| `service.loadBalancerIP`                     | 服务负载均衡器 IP                                                                                     | `""`                                            |
| `service.nodePort`                           | 服务节点端口                                                                                          | `""`                                            |
| `service.clusterIP`                          | 服务集群 IP                                                                                          | `""`                                            |
| `service.port`                               | 服务端口                                                                                             | `3306`                                          |
| `schedulerName`                              | Kubernetes scheduler 名称(不包括默认)                                                                 | `nil`                                           |
| `priorityClassName`                          | 设置 Pod 的 priorityClassName                                                                       | `{}`                                            |
| `statefulsetAnnotations`                     | StatefulSet 注释                                                                                    | `{}`                                            |
| `podAnnotations`                             | Pod 注释 map                                                                                        | `{}`                                            |
| `podLabels`                                  | Pod 标签 map                                                                                        | `{}`                                            |
| `persistence.enabled`                        | 创建一个卷存储数据                                                                                    | true                                            |
| `persistence.size`                           | PVC 容量                                                                                           | 10Gi                                            |
| `persistence.storageClass`                   | PVC 类型                                                                                           | nil                                             |
| `persistence.accessMode`                     | 访问模式                                                                                            | ReadWriteOnce                                   |
| `persistence.annotations`                    | PV 注解                                                                                            | {}                                              |

## 持久化  

[MySQL](https://hub.docker.com/repository/docker/zhyass/percona57) 镜像在容器路径 `/var/lib/mysql` 中存储 MySQL 数据和配置。

默认情况下，PersistentVolumeClaim 不可用，需通过修改 values.yaml 文件来启用持久化。开启后， PersistentVolumeClaim 会被自动创建并挂载到目录中。

> 将Pod分配给节点时，需创建一个emptyDir卷，只要Pod在该节点上运行，该卷便存在。若将Pod从节点中删除，emptyDir中的数据将被永久删除。

当使用 PersistentVolumeClaim 启用持久化存储时，需要同时修改 **livenessProbe.initialDelaySeconds** 的参数值。

>说明：PersistentVolumeClaim 中可以使用不同特性的 PersistentVolume，其 IO 性能会影响数据库的初始化性能。数据库初始化的默认限制是60秒，即livenessProbe.initialDelaySeconds + livenessProbe.periodSeconds * livenessProbe.failureThreshold的值。若初始化时间超过限制，kubelet将重启数据库容器，数据库初始化被中断，会导致持久数据不可用。

## 自定义 MySQL 配置

执行如下命令，在 `mysql.configFiles` 中添加/更改 MySQL 配置。

```yaml
  configFiles:
    node.cnf: |
      [mysqld]
      default_storage_engine=InnoDB
      max_connections=65535

      # custom mysql configuration.
      expire_logs_days=7
```
