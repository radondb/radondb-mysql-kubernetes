Contents
=============

   * [在 Kubernetes 上部署 RadonDB MySQL 集群](#在-kubernetes-上部署-radondb-mysql-集群)
      * [简介](#简介)
      * [部署准备](#部署准备)
      * [部署步骤](#部署步骤)
         * [步骤 1：克隆代码](#步骤-1-克隆代码)
         * [步骤 2：部署 Operator](#步骤-2-部署-operator)
         * [步骤 3：部署 RadonDB MySQL 集群](#步骤-3-部署-radondb-mysql-集群)
      * [部署校验](#部署校验)
         * [校验 RadonDB MySQL Operator](#校验-radondb-mysql-operator)
         * [校验 RadonDB MySQL 集群](#校验-radondb-mysql-集群)
      * [连接 RadonDB MySQL](#连接-radondb-mysql)
         * [同 NameSpace 访问](#同-namespace-访问)
         * [跨 NameSpace 访问](#跨-namespace-访问)
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

### 步骤 1: 克隆代码

```
git clone https://github.com/radondb/radondb-mysql-kubernetes.git
```

### 步骤 2: 部署 Operator

使用 Helm(V3版本) 安装 chart 的指令如下。

```
helm install [NAME] [CHART] [flags]
```

以下指定 release 名为 `demo` , 创建一个名为 `demo-mysql-operator` 的 [Deployment](#7-deployments)。

```
helm install demo radondb-mysql-kubernetes/charts/mysql-operator
```

> 说明：在这一步骤中默认将同时创建一个名为 `clusters.mysql.radondb.com` 的 [CRD](#8-CRD)。

### 步骤 3: 部署 RadonDB MySQL 集群

执行以下指令，以默认参数为 CRD `clusters.mysql.radondb.com` 创建一个实例，即创建 RadonDB MySQL 集群。您可以参照[配置](#配置)自定义集群部署参数。

```kubectl
kubectl apply -f radondb-mysql-kubernetes/config/samples/mysql_v1alpha1_cluster.yaml
```

## 部署校验

### 校验 RadonDB MySQL Operator

查看 `demo` 的 Deployment 和对应监控服务，回显如下信息则部署成功。

```shell
kubectl get deployment,svc
NAME                  READY   UP-TO-DATE   AVAILABLE   AGE
demo-mysql-operator   1/1     1            1           7h50m


NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/mysql-operator-metrics   ClusterIP   10.96.142.22    <none>        8443/TCP   8h
```

### 校验 RadonDB MySQL 集群

执行如下命令，将查看到名为 `clusters.mysql.radondb.com` 的 CRD。

```shell
kubectl get crd
NAME                                                  CREATED AT
clusters.mysql.radondb.com                            2021-06-29T02:28:36Z
```

以默认部署为例，执行如下命令将查看到名为 `sample-mysql` 的三节点 RadonDB MySQL 集群及用于访问节点的服务。

```shell
kubectl get statefulset,svc
NAME           READY   AGE
sample-mysql   3/3     7h33m

NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/sample-follower          ClusterIP   10.96.131.84    <none>        3306/TCP   7h37m
service/sample-leader            ClusterIP   10.96.111.214   <none>        3306/TCP   7h37m
service/sample-mysql             ClusterIP   None            <none>        3306/TCP   7h37m
```

## 连接 RadonDB MySQL

您需要准备一个用于连接 MySQL 的客户端。

### 同 NameSpace 访问

当客户端与 RadonDB MySQL 集群在同一个 NameSpace 中时，可使用 leader/follower service 名称代替具体的 IP 和端口。

* 连接主节点(读写节点)。

    ```shell
    $ mysql -h <leader service 名称> -u <用户名> -p
    ```

   用户名为 `qc_usr`，release 名为 `sample` ，连接主节点示例如下：

    ```shell
    $ mysql -h sample-leader -u qc_usr -p
    ```

* 连接从节点(只读节点)。

    ```shell
    $ mysql -h <follower service 名称> -u <用户名> -p
    ```

   用户名为 `qc_usr`，release 名为 `sample` ，连接从节点示例如下：

    ```shell
    $ mysql -h sample-follower -u qc_usr -p  
    ```

### 跨 NameSpace 访问

当客户端与 RadonDB MySQL 集群不在同一个 NameSpace 中时，可以通过 podIP 或服务 ClusterIP 来连接对应节点。

1. 查询 pod 列表和服务列表，分别获取需要连接的节点所在的pod 名称或对应的服务名称。

    ```shell
    $ kubectl get pod,svc
    ```

2. 查看 pod/服务的详细信息，获取对应的IP。

    ```shell
    $ kubectl describe pod <pod 名称>
    $ kubectl describe svc <服务名称>
    ```

    > 注意：pod 重启后 pod IP 会更新，需重新获取 pod IP，建议使用服务的 ClusterIP 来连接节点。

3. 连接节点。

    ```shell
    $ mysql -h <pod IP/服务 ClusterIP> -u <用户名> -p
    ```

    用户名为 `qc_usr`，pod IP 为 `10.10.128.136` ，连接示例如下：

    ```shell
    $ mysql -h 10.10.128.136 -u qc_usr -p
    ```

## 卸载

### 卸载 Operator

卸载 release 名为 `demo` 的 RadonDB MySQL Operator。

```shell
helm delete demo-mysql-operator
```

### 卸载 RadonDB MySQL

卸载 release 名为 `sample` RadonDB MySQL 集群。

```shell
kubectl delete clusters.mysql.radondb.com sample
```

### 卸载自定义资源

```shell
kubectl delete customresourcedefinitions.apiextensions.k8s.io clusters.mysql.radondb.io
```

## 配置

### 容器配置

| 参数                               | 描述                        | 默认值                                                      |
|:---------------------------------- |:---------------------------|:----------------------------------------------------------- |
| MysqlVersion                       | MySQL 版本号                | 5.7                                                         |
| MysqlOpts.RootPassword             | MySQL Root 用户密码         | ""                                                          |
| MysqlOpts.User                     | 默认新建的 MySQL 用户名称     | qc_usr                                                      |
| MysqlOpts.Password                 | 默认新建的 MySQL 用户密码     | Qing@123                                                    |
| MysqlOpts.Database                 | 默认新建的 MySQL 数据库名称   | qingcloud                                                   |
| MysqlOpts.InitTokuDB               | 是否启用TokuDB              | true                                                        |
| MysqlOpts.MysqlConf                | MySQL 配置                  |  -                                                          |
| MysqlOpts.Resources                | MySQL 容器配额              | 预留: cpu 100m, 内存 256Mi; </br> 限制: cpu 500m, 内存 1Gi  |
| XenonOpts.Image                    | xenon(高可用组件)镜像       | radondb/xenon:1.1.5-alpha                                    |
| XenonOpts.AdmitDefeatHearbeatCount | 允许的最大心跳检测失败次数    | 5                                                           |
| XenonOpts.ElectionTimeout          | 选举超时时间(单位为毫秒)      | 10000ms                                                     |
| XenonOpts.Resources                | xenon 容器配额              | 预留: cpu 50m, 内存 128Mi; </br> 限制: cpu 100m, 内存 256Mi |
| MetricsOpts.Enabled                | 是否启用 Metrics(监控)容器      | false                                                       |
| MetricsOpts.Image                  | Metrics 容器镜像地址            | prom/mysqld-exporter:v0.12.1                                |
| MetricsOpts.Resources              | Metrics 容器配额            | 预留: cpu 10m, 内存 32Mi; </br> 限制: cpu 100m, 内存 128Mi  |

### 节点配置

| 参数                      | 描述                                                | 默认值                   |
|:------------------------- |:-------------------------------------------------- |:------------------------ |
| Replicas                  | 集群节点数，只允许为0、2、3、5                        | 3                        |
| PodSpec.ImagePullPolicy   | 镜像拉取策略, 只允许为 Always/IfNotPresent/Never     | IfNotPresent             |
| PodSpec.Labels            | 节点 pod [标签](#1-标签)                            |     -                     |
| PodSpec.Annotations       | 节点 pod [注解](#2-注解)                            |     -                    |
| PodSpec.Affinity          | 节点 pod [亲和性](#3-亲和性)                         |     -                  |
| PodSpec.PriorityClassName | 节点 pod [优先级](#4-优先级)对象名称                   |     -                  |
| PodSpec.Tolerations       | 节点 pod [污点容忍度](#5-容忍)列表                   |     -                |
| PodSpec.SchedulerName     | 节点 pod [调度器](#6-调度器)名称                     | -                         |
| PodSpec.Resources         | 节点 pod 配额                                        | 预留: cpu 10m, 内存 32Mi |
| PodSpec.SidecarImage      | Sidecar 镜像                                        | radondb/mysql-sidecar:0.1       |
| PodSpec.BusyboxImage      | Busybox 镜像                                         | busybox:1.32             |
| PodSpec.SlowLogTail       | 是否开启慢日志跟踪                                     | false                    |
| PodSpec.AuditLogTail      | 是否开启审计日志跟踪                                    | false                    |

### 持久化配置

| 参数                     | 描述           | 默认值        |
|:------------------------ |:-------------- |:------------- |
| Persistence.Enabled      | 是否启用持久化 | true          |
| Persistence.AccessModes  | 存储卷访问模式 | ReadWriteOnce |
| Persistence.StorageClass | 存储卷类型     |   -            |
| Persistence.Size         | 存储卷容量     | 10Gi         |

## 参考

#### 1. [标签](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/labels/)               
#### 2. [注解](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/annotations/)
#### 3. [亲和性](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/assign-pod-node/#%E4%BA%B2%E5%92%8C%E6%80%A7%E4%B8%8E%E5%8F%8D%E4%BA%B2%E5%92%8C%E6%80%A7) 
#### 4. [优先级](https://kubernetes.io/zh/docs/concepts/configuration/pod-priority-preemption/)     
#### 5. [污点容忍度](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/taint-and-toleration/)
#### 6. [调度器](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/kube-scheduler/)
#### 7. [Deployments](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/)
#### 8. [CRD](https://kubernetes.io/zh/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
