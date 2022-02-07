Contents
=============

   * [配置参数](#配置参数)
      * [容器配置](#容器配置)
      * [节点配置](#节点配置)
      * [持久化配置](#持久化配置)

# 配置参数

## 容器配置

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

## 节点配置

| 参数                        | 描述                                             | 默认值                    |
| :-------------------------- | :----------------------------------------------- | :------------------------ |
| Replicas                    | 集群节点数，只允许为0、2、3、5                   | 3                         |
| PodPolicy.ImagePullPolicy   | 镜像拉取策略, 只允许为 Always/IfNotPresent/Never | IfNotPresent              |
| PodPolicy.Labels            | 节点 pod [标签](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/labels/))                         | -                         |
| PodPolicy.Annotations       | 节点 pod [注解](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/annotations/)                         | -                         |
| PodPolicy.Affinity          | 节点 pod [亲和性](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/assign-pod-node/#%E4%BA%B2%E5%92%8C%E6%80%A7%E4%B8%8E%E5%8F%8D%E4%BA%B2%E5%92%8C%E6%80%A7)                     | -                         |
| PodPolicy.PriorityClassName | 节点 pod [优先级](https://kubernetes.io/zh/docs/concepts/configuration/pod-priority-preemption/)对象名称             | -                         |
| PodPolicy.Tolerations       | 节点 pod [污点容忍度](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/taint-and-toleration/)列表               | -                         |
| PodPolicy.SchedulerName     | 节点 pod [调度器](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/kube-scheduler/)名称                 | -                         |
| PodPolicy.ExtraResources    | 节点容器配额（除 MySQL 和 Xenon 之外的容器）     | 预留: cpu 10m, 内存 32Mi  |
| PodPolicy.SidecarImage      | Sidecar 镜像                                     | radondb/mysql-sidecar:latest |
| PodPolicy.BusyboxImage      | Busybox 镜像                                     | busybox:1.32              |
| PodPolicy.SlowLogTail       | 是否开启慢日志跟踪                               | false                     |
| PodPolicy.AuditLogTail      | 是否开启审计日志跟踪                             | false                     |

## 持久化配置

| 参数                     | 描述           | 默认值        |
| :----------------------- | :------------- | :------------ |
| Persistence.Enabled      | 是否启用持久化 | true          |
| Persistence.AccessModes  | 存储卷访问模式 | ReadWriteOnce |
| Persistence.StorageClass | 存储卷类型     | -             |
| Persistence.Size         | 存储卷容量     | 10Gi          |
