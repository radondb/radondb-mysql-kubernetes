Contents
=================

   * [在 Kubernetes 上部署 RadonDB MySQL 集群](#在-kubernetes-上部署-radondb-mysql-集群)
      * [简介](#简介)
      * [部署准备](#部署准备)
      * [部署步骤](#部署步骤)
         * [通过 Git 部署](#通过-git-部署)
            * [步骤 1：克隆 RadonDB MySQL Chart](#步骤-1克隆-radondb-mysql-chart)
            * [步骤 2：部署](#步骤-2部署)
         * [通过 repo 部署](#通过-repo-部署)
            * [步骤 1 : 添加仓库](#步骤-1--添加仓库)
            * [步骤 2 : 部署](#步骤-2--部署)
         * [部署校验](#部署校验)
      * [连接 RadonDB MySQL](#连接-radondb-mysql)
         * [客户端与 RadonDB MySQL 在同一 NameSpace 中](#客户端与-radondb-mysql-在同一-namespace-中)
         * [客户端与 RadonDB MySQL 不在同一 NameSpace 中](#客户端与-radondb-mysql-不在同一-namespace-中)
      * [配置](#配置)
      * [持久化](#持久化)
      * [自定义 MYSQL 配置](#自定义-mysql-配置)

# 在 Kubernetes 上部署 RadonDB MySQL 集群

## 简介

RadonDB MySQL 是基于 MySQL 的开源、高可用、云原生集群解决方案。通过使用 Raft 协议，RadonDB MySQL 可以快速进行故障转移，且不会丢失任何事务。

本教程演示如何使用命令行在 Kubernetes 上部署 RadonDB MySQL。

## 部署准备

- 已成功部署 Kubernetes 集群。

## 部署步骤

### 通过 Git 部署

#### 步骤 1：克隆 RadonDB MySQL Chart

执行如下命令，将 RadonDB MySQL Chart 克隆到 Kubernetes 中。

```bash
git clone https://github.com/radondb/radondb-mysql-kubernetes.git
```

> Chart 代表 [Helm](https://helm.sh/zh/docs/intro/using_helm/) 包，包含在 Kubernetes 集群内部运行应用程序、工具或服务所需的所有资源定义。

#### 步骤 2：部署

在 radondb-mysql-kubernetes 目录路径下，选择如下方式，部署 release 实例。

> release 是运行在 Kubernetes 集群中的 Chart 的实例。通过命令方式部署，需指定 release 名称。

以下命令指定 release 名为 `demo`，将创建一个名为 `demo-radondb-mysql` 的有状态副本集。

* **默认部署方式**

   ```bash
   <For Helm v2>
    cd charts/helm
    helm install . demo

   <For Helm v3>
    cd charts/helm
    helm install demo .
  ```

* **指定参数部署方式**

  在 `helm install` 时使用 `--set key=value[,key=value]` ，可指定参数部署。
  
  以下示例以创建一个用户名为 `my-user` ，密码为 `my-password` 的标准数据库用户，可访问名为 `my-database` 的数据库。

  ```bash
  cd charts/helm
  helm install demo \
  --set mysql.mysqlUser=my-user,mysql.mysqlPassword=my-password,mysql.database=my-database .
  ```

* **配置 yaml 参数方式**

  执行如下命令，可通过 value.yaml 配置文件，在安装时配置指定参数。更多安装过程中可配置的参数，请参考 [配置](#配置) 。

  ```bash
  cd charts/helm
  helm install demo -f values.yaml .
  ```

### 通过 repo 部署

#### 步骤 1 : 添加仓库

添加并更新 helm 仓库。

```bash
$ helm repo add test https://charts.kubesphere.io/test
$ helm repo update
```

#### 步骤 2 : 部署

以下命令指定 release 名为 `demo`，将创建一个名为 `demo-radondb-mysql` 的有状态副本集。

```bash
$ helm install demo test/radondb-mysql
NAME: demo
LAST DEPLOYED: Wed Apr 28 08:08:15 2021
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
The cluster is comprised of 3 pods: 1 leader and 2 followers. Each instance is accessible within the cluster through:

    <pod-name>.demo-radondb-mysql

To connect to your database:

1. Get mysql user `qingcloud`s password:

    kubectl get secret -n default demo-radondb-mysql -o jsonpath="{.data.mysql-password}" | base64 --decode; echo

2. Run an Ubuntu pod that you can use as a client:

    kubectl run ubuntu -n default --image=ubuntu:focal -it --rm --restart='Never' -- bash -il

3. Install the mysql client:

    apt-get update && apt-get install mysql-client -y

4. To connect to leader service in the Ubuntu pod:

    mysql -h demo-radondb-mysql-leader -u qingcloud -p

5. To connect to follower service (read-only) in the Ubuntu pod:

    mysql -h demo-radondb-mysql-follower -u qingcloud -p
```

分别执行如下指令，查看到 `release` 名为 `demo` 的有状态副本集 `demo-radondb-mysql`，则 RadonDB MySQL 部署成功。

```bash
$ helm list
NAME        NAMESPACE REVISION UPDATED                                 STATUS   CHART               APP VERSION
demo        default   1        2021-04-28 08:08:15.828384203 +0000 UTC deployed radondb-mysql-1.0.0 5.7.34   

$ kubectl get statefulset
NAME                 READY   AGE
demo-radondb-mysql   3/3     25h
```

### 部署校验

部署指令执行完成后，查看 RadonDB MySQL 有状态副本集，pod 状态及服务。可查看到相关信息，则 RadonDB MySQL 部署成功。

```bash
kubectl get statefulset,pod,svc
```

## 连接 RadonDB MySQL

您需要准备一个用于连接 RadonDB MySQL 的客户端。

### 客户端与 RadonDB MySQL 在同一 NameSpace 中

当客户端与 RadonDB MySQL 集群在同一个 NameSpace 中时，可使用 leader/follower service 名称代替具体的 ip 和端口。

- 连接主节点(读写节点)。
   ```bash
   mysql -h <leader service 名称> -u <用户名> -p
   ```

- 连接从节点(只读节点)。
  ```bash
  mysql -h <follower service 名称> -u <用户名> -p
  ```

### 客户端与 RadonDB MySQL 不在同一 NameSpace 中

当客户端与 RadonDB MySQL 集群不在同一个 NameSpace 中时，需先分别获取连接所需的节点地址、节点端口、服务名称。

1. 查询 pod 列表和服务列表，分别获取 pod 名称和服务名称。

   ```bash
   kubectl get pod,svc
   ```

2. 开启服务网络访问。

   执行如下命令，打开服务配置文件，将 spec 下 type 参数设置为 `NodePort`。

      ```bash
      kubectl edit svc <服务名称>
      ```

3. 分别获取 pod 所在的节点地址和节点端口。

   ```bash
   kubectl describe pod <pod名称>
   ```

   ```bash
   kubectl describe svc <服务名称>
   ```

4. 连接节点。

   ```bash
   mysql -p <节点地址> -u <用户名> -P <节点端口> -p
   ```

> 说明：使用外网主机连接可能会出现 `SSL connection error`，需要加上 `--ssl-mode=DISABLE` 参数，关闭 SSL。

## 配置

下表列出了 RadonDB MySQL Chart 的配置参数及对应的默认值。

| 参数                                          | 描述                                                     |  默认值                                 |
| -------------------------------------------- | -------------------------------------------------------- | -------------------------------------- |
| `imagePullPolicy`                            | 镜像拉取策略                                               | `IfNotPresent`                         |
| `fullnameOverride`                           | 自定义全名覆盖                                             |                                         |
| `nameOverride`                               | 自定义名称覆盖                                             |                                         |
| `replicaCount`                               | Pod 数目                                                 | `3`                                     |
| `busybox.image`                              | `busybox` 镜像库地址                                       | `busybox`                               |
| `busybox.tag`                                | `busybox` 镜像标签                                        | `1.32`                                   |
| `mysql.image`                                | `mysql` 镜像库地址                                         | `radondb/percona`                     |
| `mysql.tag`                                  | `mysql` 镜像标签                                          | `5.7.34`                               |
| `mysql.allowEmptyRootPassword`               | 如果为 `true`，允许 root 账号密码为空                       | `true`                                  |
| `mysql.mysqlRootPassword`                    | `root` 用户密码                                          |                                          |
| `mysql.mysqlReplicationPassword`             | `qc_repl` 用户密码                                         | `Repl_123`, 如果没有设置则随机12个字符      |
| `mysql.mysqlUser`                            | 新建用户的用户名                                           | `qingcloud`                              |
| `mysql.mysqlPassword`                        | 新建用户的密码                                             | `Qing@123`, 如果没有设置则随机12个字符      |
| `mysql.mysqlDatabase`                        | 将要创建的数据库名                                          | `qingcloud`                             |
| `mysql.initTokudb`                           | 安装 tokudb 引擎                                          | `false`                                 |
| `mysql.args`                                 | 要传递到 mysql 容器的其他参数                                | `[]`                                    |
| `mysql.configFiles.node.cnf`                 | Mysql 配置文件                                            | 详见 `values.yaml`                      |
| `mysql.livenessProbe.initialDelaySeconds`    | Pod 启动后首次进行存活检查的等待时间                          | 30                                      |
| `mysql.livenessProbe.periodSeconds`          | 存活检查的间隔时间                                           | 10                                      |
| `mysql.livenessProbe.timeoutSeconds`         | 存活探针执行检测请求后，等待响应的超时时间                       | 5                                       |
| `mysql.livenessProbe.successThreshold`       | 存活探针检测失败后认为成功的最小连接成功次数                     | 1                                       |
| `mysql.livenessProbe.failureThreshold`       | 存活探测失败的重试次数，重试一定次数后将认为容器不健康             | 3                                       |
| `mysql.readinessProbe.initialDelaySeconds`   | Pod 启动后首次进行就绪检查的等待时间                           | 10                                      |
| `mysql.readinessProbe.periodSeconds`         | 就绪检查的间隔时间                                           | 10                                      |
| `mysql.readinessProbe.timeoutSeconds`        | 就绪探针执行检测请求后，等待响应的超时时间                       | 1                                       |
| `mysql.readinessProbe.successThreshold`      | 就绪探针检测失败后认为成功的最小连接成功次数                      | 1                                      |
| `mysql.readinessProbe.failureThreshold`      | 就绪探测失败的重试次数，重试一定次数后将认为容器未就绪              | 3                                      |
| `mysql.extraEnvVars`                         | 其他作为字符串传递给 `tpl` 函数的环境变量                       |                                         |
| `mysql.resources`                            | `MySQL` 的资源请求/限制                                      | 内存: `256Mi`, CPU: `100m`              |
| `xenon.image`                                | `xenon` 镜像库地址                                          | `radondb/xenon`                       |
| `xenon.tag`                                  | `xenon` 镜像标签                                            | `1.1.5-alpha`                          |
| `xenon.args`                                 | 要传递到 xenon 容器的其他参数                                 | `[]`                                   |
| `xenon.extraEnvVars`                         | 其他作为字符串传递给 `tpl` 函数的环境变量                        |                                        |
| `xenon.livenessProbe.initialDelaySeconds`    | Pod 启动后首次进行存活检查的等待时间                             | 30                                     |
| `xenon.livenessProbe.periodSeconds`          | 存活检查的间隔时间                                            | 10                                     |
| `xenon.livenessProbe.timeoutSeconds`         | 存活探针执行检测请求后，等待响应的超时时间                        | 5                                      |
| `xenon.livenessProbe.successThreshold`       | 存活探针检测失败后认为成功的最小连接成功次数                      | 1                                      |
| `xenon.livenessProbe.failureThreshold`       | 存活探测失败的重试次数，重试一定次数后将认为容器不健康              | 3                                      |
| `xenon.readinessProbe.initialDelaySeconds`   | Pod 启动后首次进行就绪检查的等待时间                            | 10                                     |
| `xenon.readinessProbe.periodSeconds`         | 就绪检查的间隔时间                                            | 10                                     |
| `xenon.readinessProbe.timeoutSeconds`        | 就绪探针执行检测请求后，等待响应的超时时间                        | 1                                      |
| `xenon.readinessProbe.successThreshold`      | 就绪探针检测失败后认为成功的最小连接成功次数                       | 1                                      |
| `xenon.readinessProbe.failureThreshold`      | 就绪探测失败的重试次数，重试一定次数后将认为容器未就绪              | 3                                      |
| `xenon.resources`                            | `xenon` 的资源请求/限制                                      | 内存: `128Mi`, CPU: `50m`               |
| `metrics.enabled`                            | 以 side-car 模式开启 Prometheus Exporter                     | `true`                                 |
| `metrics.image`                              | Exporter 镜像地址                                            | `prom/mysqld-exporter`                 |
| `metrics.tag`                                | Exporter 标签                                               | `v0.12.1`                              |
| `metrics.annotations`                        | Exporter 注释                                               | `{}`                                   |
| `metrics.livenessProbe.initialDelaySeconds`  | Pod 启动后首次进行存活检查的等待时间                            | 15                                     |
| `metrics.livenessProbe.timeoutSeconds`       | 存活探针执行检测请求后，等待响应的超时时间                        | 5                                      |
| `metrics.readinessProbe.initialDelaySeconds` | Pod 启动后首次进行就绪检查的等待时间                            | 5                                      |
| `metrics.readinessProbe.timeoutSeconds`      | 就绪探针执行检测请求后，等待响应的超时时间                        | 1                                      |
| `metrics.serviceMonitor.enabled`             | 若设置为 `true`, 将为 Prometheus operator 创建 ServiceMonitor | `true`                                 |
| `metrics.serviceMonitor.namespace`           | 创建 ServiceMonitor 时，可指定命名空间                         | `nil`                                   |
| `metrics.serviceMonitor.interval`            | 数据采集间隔，若未指定，将使用 Prometheus 默认设置                | 10s                                     |
| `metrics.serviceMonitor.scrapeTimeout`       | 数据采集超时时间，若未指定，将使用 Prometheus 默认设置             | `nil`                                   |
| `metrics.serviceMonitor.selector`            | 默认为 kube-prometheus                                       | `{ prometheus: kube-prometheus }`       |
| `slowLogTail`                                | 若设置为 `true`，将启动一个容器用来查看 mysql-slow.log           | `true`                                 |
| `resources`                                  | 资源 请求/限制                                               | 内存: `32Mi`, CPU: `10m`                |
| `service.annotations`                        | Kubernetes 服务注释                                         | {}                                     |
| `service.type`                               | Kubernetes 服务类型                                         | NodePort                                |
| `service.loadBalancerIP`                     | 服务负载均衡器 IP                                            | `""`                                   |
| `service.nodePort`                           | 服务节点端口                                                 | `""`                                   |
| `service.clusterIP`                          | 服务集群 IP                                                 | `""`                                   |
| `service.port`                               | 服务端口                                                    | `3306`                                 |
| `rbac.create`                                | 若为 true,将创建和使用 RBAC 资源                              | `true`                                  |
| `serviceAccount.create`                      | 指定是否创建 ServiceAccount                                  | `true`                                  |
| `serviceAccount.name`                        | ServiceAccount 的名称                                       |                                         |
| `persistence.enabled`                        | 创建一个卷存储数据                                           | true                                   |
| `persistence.size`                           | PVC 容量                                                  | 10Gi                                   |
| `persistence.storageClass`                   | PVC 类型                                                  | nil                                    |
| `persistence.accessMode`                     | 访问模式                                                   | ReadWriteOnce                          |
| `persistence.annotations`                    | PV 注解                                                   | {}                                     |
| `priorityClassName`                          | 设置 Pod 的 priorityClassName                              | `{}`                                   |
| `schedulerName`                              | Kubernetes scheduler 名称(不包括默认)                        | `nil`                                  |
| `statefulsetAnnotations`                     | StatefulSet 注释                                           | `{}`                                   |
| `podAnnotations`                             | Pod 注释 map                                               | `{}`                                   |
| `podLabels`                                  | Pod 标签 map                                               | `{}`                                   |

## 持久化  

[MySQL](https://hub.docker.com/repository/docker/radondb/percona) 镜像在容器路径 `/var/lib/mysql` 中存储 MYSQL 数据和配置。

默认情况下，会创建一个 PersistentVolumeClaim 并将其挂载到指定目录中。 若想禁用此功能，您可以更改 `values.yaml` 禁用持久化，改用 emptyDir。 

> *"当 Pod 分配给节点时，将首先创建一个 emptyDir 卷，只要该 Pod 在该节点上运行，该卷便存在。 当 Pod 从节点中删除时，emptyDir 中的数据将被永久删除."*

**注意**：PersistentVolumeClaim 中可以使用不同特性的 PersistentVolume，其 IO 性能会影响数据库的初始化性能。所以当使用 PersistentVolumeClaim 启用持久化存储时，可能需要调整 livenessProbe.initialDelaySeconds 的值。数据库初始化的默认限制是60秒 (livenessProbe.initialDelaySeconds + livenessProbe.periodSeconds * livenessProbe.failureThreshold)。如果初始化时间超过限制，kubelet将重启数据库容器，数据库初始化被中断，会导致持久数据不可用。

## 自定义 MYSQL 配置

在 `mysql.configFiles` 中添加/更改 MySQL 配置。

```yaml
  configFiles:
    node.cnf: |
      [mysqld]
      default_storage_engine=InnoDB
      max_connections=65535

      # custom mysql configuration.
      expire_logs_days=7
```
