> [English](README.md) | 简体中文

<p align="center">
<a href="https://radondb.com/"><img src="https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/images/logo_radondb-mysql.png?raw=true" alt="banner" width="200px"></a>
</p>
<p align="center">
<b><i>面向云原生、容器化的数据库开源社区</i></b>
</p>

<p align=center>
<a href="https://goreportcard.com/report/github.com/radondb/radondb-mysql-kubernetes"><img src="https://goreportcard.com/badge/github.com/radondb/radondb-mysql-kubernetes" alt="A+"></a>
<a href="https://img.shields.io/github/stars/radondb/radondb-mysql-kubernetes"><img src="https://img.shields.io/github/stars/radondb/radondb-mysql-kubernetes"></a>
<a href="https://img.shields.io/github/issues/radondb/radondb-mysql-kubernetes"><img src="https://img.shields.io/github/issues/radondb/radondb-mysql-kubernetes"></a>
<a href="https://img.shields.io/github/forks/radondb/radondb-mysql-kubernetes"><img src="https://img.shields.io/github/forks/radondb/radondb-mysql-kubernetes"></a>
<a href="https://img.shields.io/github/v/release/radondb/radondb-mysql-kubernetes?include_prereleases"><img src="https://img.shields.io/github/v/release/radondb/radondb-mysql-kubernetes?include_prereleases"></a>
<a href="https://img.shields.io/github/license/radondb/radondb-mysql-kubernetes"><img src="https://img.shields.io/github/license/radondb/radondb-mysql-kubernetes"></a>
</p>

----

# 什么是 RadonDB MySQL

**RadonDB MySQL** 是基于 MySQL 的开源、高可用、云原生集群解决方案。支持一主多从高可用架构，并具备安全、自动备份、监控告警、自动扩容等全套管理功能。

**RadonDB MySQL Kubernetes**支持在 [Kubernetes](https://kubernetes.io/) 、[KubeSphere](https://kubesphere.com.cn/) 和 [Rancher](https://rancher.com) 上安装部署和管理，自动执行与运行 RadonDB MySQL 集群有关的任务。

## 核心功能
🧠 **MySQL 高可用**：无中心化自动选主、主从秒级切换、集群切换的数据强一致性
 

✏️ **集群管理**

💻 [**监控告警**](docs/zh-cn/deploy_monitoring.md)

✍️ [**S3 备份**](docs/zh-cn/backup_and_restoration_s3.md)和 [**NFS 备份**](docs/zh-cn/backup_and_restoration_nfs.md)

🎈 **集群日志管理**

👨 [**账户管理**](docs/zh-cn/mgt_mysqluser.md)


## 架构图

1、 通过 Raft 协议实现无中心化领导者自动选举

2、 通过 Semi-Sync基于GTID 模式同步数据

3、 通过 [Xenon](https://github.com/radondb/xenon.git) 提供高可用能力

<p align="center">
<a href="https://github.com/radondb/"><img src="https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/images/radondb-mysql_Architecture.png?raw=true" alt="banner" width="800px"></a>
</p>

## 功能规划

| 版本 | 功能  | 实现方式 |
|------|--------|------| 
| 3.0  | 自动化运维 <br> 多节点角色 <br> 灾备集群 <br> SSL 传输加密 | Operator |
| 2.0  | 增删节点 <br> 升级集群 <br> 备份与恢复 <br> 故障自动转移 <br> 重建节点 <br> 账户管理   |  Operator |
| 1.0 |  集群管理 <br> 监控告警 <br> 集群日志管理 <br> 账户管理 | Helm |

# 快速开始

👀 本教程主要演示如何在 Kubernetes 上部署 RadonDB MySQL 集群(Operator)。

## 部署准备

📦 已准备可用 Kubernetes 集群。
## 部署步骤

### 步骤 1: 添加 Helm 仓库

```plain
helm repo add radondb https://radondb.github.io/radondb-mysql-kubernetes/
```
### 步骤 2: 部署 Operator

以下指定 release 名为 `demo` , 创建一个名为 `demo-mysql-operator` 的 [Deployment](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/)。

```plain
helm install demo radondb/mysql-operator
```
>**说明**
>在这一步骤中默认将同时创建集群所需的 [CRD](https://kubernetes.io/zh/docs/concepts/extend-kubernetes/api-extension/custom-resources/)。 

### 步骤 3: 部署 RadonDB MySQL 集群

执行以下指令，以默认参数为 CRD `mysqlclusters.mysql.radondb.com` 创建一个实例，即创建 RadonDB MySQL 集群。您可参见[配置参数](https://./config_para.md)说明，自定义集群部署参数。

```plain
kubectl apply -f https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqlcluster.yaml
```

## 操作视频

在 Kubernetes 上部署 RadonDB MySQL Operator 和 MySQL 集群，快速查看 👉  [Demo 视频](https://radondb.com/docs/mysql/v2.1.3/vadio/install/#content)

📖 了解更多，请查看文档：

* [在 Kubernetes 上部署 RadonDB MySQL 集群](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/zh-cn/deploy_radondb-mysql_operator_on_k8s.md)
* [在 KubeSphere 上部署 RadonDB MySQL 集群](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/zh-cn/deploy_radondb-mysql_operator_on_kubesphere.md)
* [在 Rancher 上部署 RadonDB MySQL 集群](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/zh-cn/deploy_radondb-mysql_operator_on_rancher.md)

# 用户案例

![](docs/images/%E5%AE%A2%E6%88%B7%E6%A1%88%E4%BE%8B.png)

## 协议

RadonDB MySQL 基于 Apache 2.0 协议，详见 [License](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/LICENSE)。

## 欢迎加入社区话题互动 ❤️

😊 社区官网：[https://radondb.com](https://radondb.com)

😁 社区论坛：请加入 KubeSphere 开发者论坛 [RadonDB](https://kubesphere.com.cn/forum/t/RadonDB) 板块。

😆 社区公众号：RadonDB 开源社区

🦉 社区微信群：请添加群助手 radondb 邀请进群

如有任何关于 RadonDB MySQL 的 Bug、问题或建议，请在 GitHub 提交 [issue](https://github.com/radondb/radondb-mysql-kubernetes/issues) 或[论坛](https://kubesphere.com.cn/forum/t/RadonDB)反馈。

