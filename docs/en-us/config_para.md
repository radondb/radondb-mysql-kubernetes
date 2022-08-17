English | [简体中文](../zh-cn/config_para.md)

Contents
=============

   * [Parameter Configuration](#parameter-configuration)
      * [Container](#Container)
      * [Pod](#pod)
      * [Persistence](#Persistence)

# Parameter Configuration

## Container

| Parameter                               | Description                        | Default                                                      |
| :--------------------------------- | :-------------------------- | :---------------------------------------------------------- |
| MysqlVersion                       | MySQL version                | 5.7                                                         |
| MysqlOpts.RootPassword             | MySQL root user password         | ""                                                          |
| MysqlOpts.User                     | Default MySQL username   | radondb_usr                                                 |
| MysqlOpts.Password                 | Default MySQL user password   | RadonDB@123                                                 |
| MysqlOpts.Database                 | Default database name | radondb                                                     |
| MysqlOpts.InitTokuDB               | TokuDB enabled              | true                                                        |
| MysqlOpts.MysqlConf                | MySQL configuration                  | -                                                           |
| MysqlOpts.Resources                | MySQL container resources              | Reserve: CPU 100M, memory 256Mi; </br> limit: CPU 500M, memory 1Gi  |
| XenonOpts.Image                    | Xenon (HA MySQL) image       | radondb/xenon:v2.2.1                                   |
| XenonOpts.AdmitDefeatHearbeatCount | Maximum heartbeat failures allowed  | 5                                                           |
| XenonOpts.ElectionTimeout          | Election timeout period (milliseconds)    | 10000 ms                                                     |
| XenonOpts.Resources                | Xenon container resources              | Reserve: CPU 50M, memory 128Mi; </br> limit: CPU 100M, memory 256Mi |
| MetricsOpts.Enabled                | Metrics (monitor) container enabled  | false                                                       |
| MetricsOpts.Image                  | Metrics container image        | prom/mysqld-exporter:v0.12.1                                |
| MetricsOpts.Resources              | Metrics container resources            | Reserve: CPU 10M, memory 32Mi; </br>limit: CPU 100M, memory 128Mi |

## Pod

| Parameter                        | Description                                             | Default                    |
| :-------------------------- | :----------------------------------------------- | :------------------------ |
| Replicas                    | The number of cluster nodes. The value 0, 2, 3 and 5 are allowed.                   | 3                         |
| PodPolicy.ImagePullPolicy   | The image pull policy is only allowed to be Always/IfnNotPresent/Never. | IfNotPresent              |
| PodPolicy.Labels            | Pod [labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)                         | -                         |
| PodPolicy.Annotations       | Pod [annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/)                         | -                         |
| PodPolicy.Affinity          | Pod [affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/)                     | -                         |
| PodPolicy.PriorityClassName | Pod [priority](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/) class name             | -                         |
| PodPolicy.Tolerations       | Pod [toleration](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) list               | -                         |
| PodPolicy.SchedulerName     | Pod [scheduler](https://kubernetes.io/docs/concepts/scheduling-eviction/kube-scheduler/) name                 | -                         |
| PodPolicy.ExtraResources    | Node resources (containers except MySQL and Xenon)     | Reserve: CPU 10M, memory 32Mi  |
| PodPolicy.SidecarImage      | Sidecar image                                     | radondb/mysql-sidecar:latest |
| PodPolicy.BusyboxImage      | Busybox image                                     | busybox:1.32              |
| PodPolicy.SlowLogTail       | SlowLogTail enabled                               | false                     |
| PodPolicy.AuditLogTail      | AuditLogTail enabled                             | false                     |

## Persistence

| Parameter                     | Description           | Default        |
| :----------------------- | :------------- | :------------ |
| Persistence.Enabled      | Persistence enabled | true          |
| Persistence.AccessModes  | Access mode | ReadWriteOnce |
| Persistence.StorageClass | Storage class     | -             |
| Persistence.Size         | Size     | 10Gi          |
