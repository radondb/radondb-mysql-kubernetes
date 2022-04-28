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
<a href="https://img.shields.io/github/license/radondb/radondb-mysql-kubernetes"><img src="https://img.shields.io/github/license/radondb/radondb-mysql-kubernetes"></a>
</p>

----

# 什么是 RadonDB MySQL

**RadonDB MySQL** 是基于 MySQL 的开源、高可用、云原生集群解决方案。支持一主多从高可用架构，并具备安全、自动备份、监控告警、自动扩容等全套管理功能。

**RadonDB MySQL Kubernetes**支持在 [Kubernetes](https://kubernetes.io/) 、[KubeSphere](https://kubesphere.com.cn/) 和 [Rancher](https://rancher.com) 上安装部署和管理，自动执行与运行 RadonDB MySQL 集群有关的任务。

# 最新版本

RadonDB MySQL Kubernetes 2.1.4 全新发布！多项功能优化，带来更好的用户体验，详见 [v2.1.4 发行记录](https://radondb.com/docs/mysql/v2.1.3/release/2.1.3)。

## 核心特性

🧠 **数据强一致性**：采用一主多备高可用架构，自动脑裂保护处理。

✏️ **高可用**：支持一主多备架构，灵活满足各类可用性需求。

💻 **自动运维**：可设置自动备份策略、监控告警策略、自动扩容策略。

🎈 **弹性扩缩容**：根据业务需要实时扩展数据库的 CPU、内存、存储容量。

## 架构图

1⃣️ 通过 Raft 协议实现无中心化领导者自动选举
2⃣️ 通过 Semi-Sync基于GTID 模式同步数据
3⃣️ 通过 [Xenon](https://github.com/radondb/xenon.git) 提供高可用能力

<p align="center">
<a href="https://github.com/radondb/"><img src="https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/images/radondb-mysql_Architecture.png?raw=true" alt="banner" width="800px"></a>
</p>

## 功能规划

```mermaid
graph LR
     A[3.0-Operator<br><br>自动化运维<br>多节点角色<br>灾备集群<br>SSL 传输加密]--- B[2.0-Operator<br><br>增删节点<br>自动扩缩容<br>升级集群<br>备份与恢复<br>故障自动转移<br>自动重建节点<br>自动重启服务<br>账户管理-提供 API 接口<br>在线迁移]---C[1.0-Helm<br><br>集群管理<br>监控告警<br>集群日志管理<br>账户管理]
```

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

**部署版本：RadonDB MySQL Kubernetes 2.1.3**

在 Kubernetes 上部署 RadonDB MySQL Operator 和 MySQL 集群，快速查看 👉  [Demo 视频](https://radondb.com/docs/mysql/v2.1.3/vadio/install/#content)

📖 了解更多，请查看文档：

* [在 Kubernetes 上部署 RadonDB MySQL 集群](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/zh-cn/deploy_radondb-mysql_operator_on_k8s.md)
* [在 KubeSphere 上部署 RadonDB MySQL 集群](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/zh-cn/deploy_radondb-mysql_operator_on_kubesphere.md)
* [在 Rancher 上部署 RadonDB MySQL 集群](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/zh-cn/deploy_radondb-mysql_operator_on_rancher.md)

## 协议

RadonDB MySQL 基于 Apache 2.0 协议，详见 [License](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/LICENSE)。

## 欢迎加入社区话题互动 ❤️

😊 社区官网：[https://radondb.com](https://radondb.com)

😁 社区论坛：请加入 KubeSphere 开发者论坛 [RadonDB](https://kubesphere.com.cn/forum/t/RadonDB) 板块。

😆 社区公众号：RadonDB 开源社区

🦉 社区微信群：请添加群助手 radondb 邀请进群

如有任何关于 RadonDB MySQL 的 Bug、问题或建议，请在 GitHub 提交 [issue](https://github.com/radondb/radondb-mysql-kubernetes/issues) 或[论坛](https://kubesphere.com.cn/forum/t/RadonDB)反馈。