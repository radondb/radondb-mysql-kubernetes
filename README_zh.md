![](docs/images/logo_radondb-mysql.png)  <br>

[English](README.md) | 中文 

# 什么是RadonDB MySQL

[RadonDB MySQL](https://github.com/radondb/radondb-mysql-kubernetes) 是基于 MySQL 的开源、高可用、云原生集群解决方案。支持一主多从高可用架构，并具备安全、自动备份、监控告警、自动扩容等全套管理功能。

## RadonDB MySQL Kubernetes

RadonDB MySQL Kubernetes支持在[Kubernetes](https://kubernetes.io)和[KubeSphere](https://kubesphere.com.cn)上安装部署和管理，自动执行与运行RadonDB MySQL集群有关的任务。

## 架构图

- 通过 Raft 协议实现无中心化领导者自动选举
- 通过 Semi-Sync基于GTID 模式同步数据
- 通过 [Xenon](https://github.com/radondb/xenon.git) 提供高可用能力

![](docs/images/radondb-mysql_Architecture.png)

## 核心功能

- MySQL 高可用
    - 无中心化自动选主
    - 主从秒级切换
    - 集群切换的数据强一致性
- 集群管理
- 监控告警
- 集群日志管理
- 账户管理

## 快速开始

- [ 在 Kubernetes 上部署 RadonDB MySQL 集群](docs/Kubernetes/deploy_radondb-mysql_on_kubernetes.md)
- [在 KubeSphere 上通过 Git 部署 RadonDB MySQL 集群](docs/KubeSphere/deploy_radondb-mysql_on_kubesphere.md)
- [在 KubeSphere 上通过 Helm Repo 部署 RadonDB MySQL 集群](docs/KubeSphere/deploy_radondb-mysql_on_kubesphere_repo.md)
- [在 KubeSphere 上通过应用商店部署 RadonDB MySQL 集群](docs/KubeSphere/deploy_radondb-mysql_on_kubesphere_appstore.md)

## 路线图

| 版本 | 功能  | 实现方式 |
|------|--------|------| 
| 1.0 | MySQL 高可用 <br> 无中心化领导者自动选举<br> 主从秒级切换<br> 数据强一致性 <br> 集群管理 <br> 监控告警 <br> 集群日志管理 <br> 账户管理 | Helm |
| 2.0  | 增删节点 <br> 自动扩缩容 <br> 升级集群 <br> 备份与恢复 <br> 故障自动转移 <br> 自动重建节点 <br> 自动重启服务 <br> 账户管理（提供 API 接口）<br> 在线迁移   |  Operator |
| 3.0  | 自动化运维 <br> 多节点角色 <br> 灾备集群 <br> SSL 传输加密 | Operator |

## 用户案例

![](docs/images/users.png)

## 协议

RadonDB MySQL 基于 Apache 2.0 协议，详见 [LICENSE](./LICENSE)。

## 欢迎加入社区话题互动

- 论坛

    请加入[Kubesphere 开发者社区](https://kubesphere.com.cn/forum/t/radondb) RadonDB MySQL 话题专区。
    
- 微信群 
    
    进入 RadonDB 社区微信群，请添加管理员微信。

    ![](docs/images/wechat_admin.jpg)

有任何 RadonDB MySQL 错误、问题和功能，请您在 RadonDB MySQL GitHub 上提交Issue。
