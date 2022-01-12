Contents
=============

   * [在 Kubernetes 上部署 RadonDB MySQL 集群](#在-kubernetes-上部署-radondb-mysql-集群)
      * [简介](#简介)
      * [部署准备](#部署准备)
      * [部署步骤](#部署步骤)
         * [步骤 1：添加 Helm 仓库](#步骤-1-添加-helm-仓库)
         * [步骤 2：部署 Operator](#步骤-2-部署-operator)
         * [步骤 3：部署 RadonDB MySQL 集群](#步骤-3-部署-radondb-mysql-集群)
      * [部署校验](#部署校验)
         * [校验 RadonDB MySQL Operator](#校验-radondb-mysql-operator)
         * [校验 RadonDB MySQL 集群](#校验-radondb-mysql-集群)
      * [访问 RadonDB MySQL](#访问-radondb-mysql)
      * [卸载](#卸载)
         * [卸载 Operator](#卸载-Operator)
         * [卸载 RadonDB MySQL](#卸载-RadonDB-MySQL)
         * [卸载自定义资源](#卸载自定义资源)
      * [配置](#配置)
         * [容器配置](#容器配置)
         * [节点配置](#节点配置)
         * [持久化配置](#持久化配置)
      * [参考](#参考)

# 在 Kubernetes 上部署 RadonDB MySQL 集群(Operator)

## 简介

RadonDB MySQL 是一款基于 MySQL 的开源、高可用、云原生集群解决方案。支持一主多从高可用架构，并具备安全、自动备份、监控告警、自动扩容等全套管理功能。目前已经在生产环境中大规模的使用，包含**银行、保险、传统大企业**等。

RadonDB MySQL 支持在 Kubernetes 上安装部署和管理，自动执行与运行 RadonDB MySQL 集群有关的任务。

本教程主要演示如何在 Kubernetes 上部署 RadonDB MySQL 集群(Operator)。

## 部署准备

* 已准备可用 Kubernetes 集群。

## 部署步骤

### 步骤 1: 添加 Helm 仓库

```
$ helm repo add radondb https://radondb.github.io/radondb-mysql-kubernetes/
```

校验仓库，可查看到名为 `radondb/mysql-operator` 的 chart。
```
$ helm search repo
NAME                            CHART VERSION   APP VERSION                     DESCRIPTION                 
radondb/mysql-operator          0.1.0           v2.1.0                          Open Source，High Availability Cluster，based on MySQL                     
```

### 步骤 2: 部署 Operator


以下指定 release 名为 `demo` , 创建一个名为 `demo-mysql-operator` 的 [Deployment](#7-deployments)。

```
$ helm install demo radondb/mysql-operator
```

> 说明：在这一步骤中默认将同时创建集群所需的 [CRD](#8-CRD)。

### 步骤 3: 部署 RadonDB MySQL 集群

执行以下指令，以默认参数为 CRD `mysqlclusters.mysql.radondb.com` 创建一个实例，即创建 RadonDB MySQL 集群。您可以参照[配置](#配置)自定义集群部署参数。

```kubectl
$ wget https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqlcluster.yaml
$ kubectl apply -f mysql_v1alpha1_mysqlcluster.yaml
```

## 部署校验

### 校验 RadonDB MySQL Operator

查看 `demo` 的 Deployment 和对应监控服务，回显如下信息则部署成功。

```shell
$ kubectl get deployment,svc
NAME                  READY   UP-TO-DATE   AVAILABLE   AGE
demo-mysql-operator   1/1     1            1           7h50m


NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/mysql-operator-metrics   ClusterIP   10.96.142.22    <none>        8443/TCP   8h
```

### 校验 RadonDB MySQL 集群

执行如下命令，将查看到如下 CRD。

```shell
$ kubectl get crd | grep mysql.radondb.com
backups.mysql.radondb.com                             2021-11-02T07:00:01Z
mysqlclusters.mysql.radondb.com                       2021-11-02T07:00:01Z
mysqlusers.mysql.radondb.com                          2021-11-02T07:00:01Z
```

以默认部署为例，执行如下命令将查看到名为 `sample-mysql` 的三节点 RadonDB MySQL 集群及用于访问节点的服务。

```shell
$ kubectl get statefulset,svc
NAME           READY   AGE
sample-mysql   3/3     7h33m

NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/sample-follower          ClusterIP   10.96.131.84    <none>        3306/TCP   7h37m
service/sample-leader            ClusterIP   10.96.111.214   <none>        3306/TCP   7h37m
service/sample-mysql             ClusterIP   None            <none>        3306/TCP   7h37m
```

## 访问 RadonDB MySQL

> **注意**
> 
> 准备可用于连接 MySQL 的客户端。

- 当客户端的与数据库部署在不同 Kubernetes 集群，请参考 [Kubernetes 访问集群中的应用程序](https://kubernetes.io/zh/docs/tasks/access-application-cluster/)，配置端口转发、负载均衡等连接方式。

- 在 Kubernetes 集群内，支持使用 `service_name` 或者 `clusterIP` 方式，访问 RadonDB MySQL。
  
   > **说明**
   > 
   > RadonDB MySQL 提供 leader 服务和 follower 服务用于分别访问主从节点。leader 服务始终指向主节点（读写），follower 服务始终指向从节点（只读）。

以下为客户端与数据库在同一 Kubernetes 集群内，访问 RadonDB MySQL 的方式。

### `service_name` 方式

* 连接 leader 服务(RadonDB MySQL 主节点)

    ```shell
    $ mysql -h <leader_service_name>.<namespace> -u <user_name> -p
    ```

   用户名为 `radondb_usr`，release 名为 `sample`，RadonDB MySQL 命名空间为 `default` ，连接示例如下：

    ```shell
    $ mysql -h sample-leader.default -u radondb_usr -p
    ```

* 连接 follower 服务(RadonDB MySQL 从节点)

    ```shell
    $ mysql -h <follower_service_name>.<namespace> -u <user_name> -p
    ```

   用户名为 `radondb_usr`，release 名为 `sample`，RadonDB MySQL 命名空间为 `default` ，连接示例如下：

    ```shell
    $ mysql -h sample-follower.default -u radondb_usr -p  
    ```

### `clusterIP` 方式

RadonDB MySQL 的高可用读写 IP 指向 leader 服务的 `clusterIP`，高可用只读 IP 指向 follower 服务的 `clusterIP`。

```shell
$ mysql -h <clusterIP> -P <mysql_Port> -u <user_name> -p
```

以下示例用户名为 `radondb_usr`， leader 服务的 clusterIP 为 `10.10.128.136` ，连接示例如下：

```shell
$ mysql -h 10.10.128.136 -P 3306 -u radondb_usr -p
```

## 卸载

### 卸载 Operator

卸载当前命名空间下 release 名为 `demo` 的 RadonDB MySQL Operator。

```shell
$ helm delete demo
```

### 卸载 RadonDB MySQL

卸载 release 名为 `sample` RadonDB MySQL 集群。

```shell
$ kubectl delete mysqlclusters.mysql.radondb.com sample
```

### 卸载自定义资源

```shell
$ kubectl delete customresourcedefinitions.apiextensions.k8s.io mysqlclusters.mysql.radondb.com
$ kubectl delete customresourcedefinitions.apiextensions.k8s.io mysqlusers.mysql.radondb.com
$ kubectl delete customresourcedefinitions.apiextensions.k8s.io backups.mysql.radondb.com
```

## 配置

### 容器配置

| 参数                               | 描述                        | 默认值                                                      |
| :--------------------------------- | :-------------------------- | :---------------------------------------------------------- |
| MysqlVersion                       | MySQL 版本号                | 5.7                                                         |
| MysqlOpts.RootPassword             | MySQL Root 用户密码         | ""                                                          |
| MysqlOpts.User                     | 默认新建的 MySQL 用户名称   | radondb_usr                                                 |
| MysqlOpts.Password                 | 默认新建的 MySQL 用户密码   | RadonDB@123                                                 |
| MysqlOpts.Database                 | 默认新建的 MySQL 数据库名称 | radondb                                                     |
| MysqlOpts.InitTokuDB               | 是否启用TokuDB              | true                                                        |
| MysqlOpts.MysqlConf                | MySQL 配置                  | -                                                           |
| MysqlOpts.Resources                | MySQL 容器配额              | 预留: cpu 100m, 内存 256Mi; </br> 限制: cpu 500m, 内存 1Gi  |
| XenonOpts.Image                    | xenon(高可用组件)镜像       | radondb/xenon:1.1.5-alpha                                   |
| XenonOpts.AdmitDefeatHearbeatCount | 允许的最大心跳检测失败次数  | 5                                                           |
| XenonOpts.ElectionTimeout          | 选举超时时间(单位为毫秒)    | 10000ms                                                     |
| XenonOpts.Resources                | xenon 容器配额              | 预留: cpu 50m, 内存 128Mi; </br> 限制: cpu 100m, 内存 256Mi |
| MetricsOpts.Enabled                | 是否启用 Metrics(监控)容器  | false                                                       |
| MetricsOpts.Image                  | Metrics 容器镜像        | prom/mysqld-exporter:v0.12.1                                |
| MetricsOpts.Resources              | Metrics 容器配额            | 预留: cpu 10m, 内存 32Mi; </br> 限制: cpu 100m, 内存 128Mi  |

### 节点配置

| 参数                        | 描述                                             | 默认值                    |
| :-------------------------- | :----------------------------------------------- | :------------------------ |
| Replicas                    | 集群节点数，只允许为0、2、3、5                   | 3                         |
| PodPolicy.ImagePullPolicy   | 镜像拉取策略, 只允许为 Always/IfNotPresent/Never | IfNotPresent              |
| PodPolicy.Labels            | 节点 pod [标签](#1-标签)                         | -                         |
| PodPolicy.Annotations       | 节点 pod [注解](#2-注解)                         | -                         |
| PodPolicy.Affinity          | 节点 pod [亲和性](#3-亲和性)                     | -                         |
| PodPolicy.PriorityClassName | 节点 pod [优先级](#4-优先级)对象名称             | -                         |
| PodPolicy.Tolerations       | 节点 pod [污点容忍度](#5-容忍)列表               | -                         |
| PodPolicy.SchedulerName     | 节点 pod [调度器](#6-调度器)名称                 | -                         |
| PodPolicy.ExtraResources    | 节点容器配额（除 MySQL 和 Xenon 之外的容器）     | 预留: cpu 10m, 内存 32Mi  |
| PodPolicy.SidecarImage      | Sidecar 镜像                                     | radondb/mysql-sidecar:latest |
| PodPolicy.BusyboxImage      | Busybox 镜像                                     | busybox:1.32              |
| PodPolicy.SlowLogTail       | 是否开启慢日志跟踪                               | false                     |
| PodPolicy.AuditLogTail      | 是否开启审计日志跟踪                             | false                     |

### 持久化配置

| 参数                     | 描述           | 默认值        |
| :----------------------- | :------------- | :------------ |
| Persistence.Enabled      | 是否启用持久化 | true          |
| Persistence.AccessModes  | 存储卷访问模式 | ReadWriteOnce |
| Persistence.StorageClass | 存储卷类型     | -             |
| Persistence.Size         | 存储卷容量     | 10Gi          |

## 参考

#### 1. [标签](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/labels/)               
#### 2. [注解](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/annotations/)
#### 3. [亲和性](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/assign-pod-node/#%E4%BA%B2%E5%92%8C%E6%80%A7%E4%B8%8E%E5%8F%8D%E4%BA%B2%E5%92%8C%E6%80%A7) 
#### 4. [优先级](https://kubernetes.io/zh/docs/concepts/configuration/pod-priority-preemption/)     
#### 5. [污点容忍度](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/taint-and-toleration/)
#### 6. [调度器](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/kube-scheduler/)
#### 7. [Deployments](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/)
#### 8. [CRD](https://kubernetes.io/zh/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
