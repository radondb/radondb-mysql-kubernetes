# XenonDB

> [English](README.md)| 中文 

## 什么是 XenonDB？

XenonDB 是基于 MySQL 的开源、高可用、云原生集群解决方案。支持一主多从高可用架构，并具备安全、自动备份、监控告警、自动扩容等全套管理功能。


## 架构图

- 通过 Raft 协议实现无中心化选主
- 通过 Semi-Sync 基于 GTID 模式同步数据

![](docs/images/XenonDB_Architecture_1.png)

## 核心功能

- MySQL 高可用
    - 无中心化自动选主
    - 主从秒级别切换
    - 集群切换的数据强一致性
- 集群管理
- 监控告警
- 集群日志管理
- 账户管理

## 快速部署

目前已支持在 Kubernetes 和 KubeSphere 平台的部署。
- [在 Kubernetes 平台部署](docs/Kubernetes/deploy_xenondb_on_kubernetes.md)
- [在 KubeSphere 应用商店部署](docs/KubeSphere/deploy_xenondb_on_kubesphere.md)

## 版本路线

| 版本 | 功能  |
|------|--------|
| Helm  | MySQL 高可用 <br> 无中心化自动选主<br> 主从秒级别切换<br> 数据强一致性 <br> 集群管理 <br> 监控告警 <br> 集群日志管理 <br> 账户管理 |
| Operator 1.0  | 增删节点 <br> 自动扩缩容 <br> 升级集群 <br> 备份与恢复 <br> 故障自动转移 <br> 自动重建节点 <br> 自动重启服务 <br> 账户管理（提供 API 接口）<br> 在线迁移   |
| Operator 2.0  | 自动化运维 <br> 多节点角色 <br> 灾备集群 <br> SSL 传输加密 |


## 应用案例

![](docs/images/users.png)

## 协议

XenonDB 基于 Apache 2.0 协议，详见[LICENSE](./LICENSE)。

## 欢迎加入社区话题互动

- 论坛：
  
  欢迎请加入 [Kubesphere 开发者社区](https://github.com/kubesphere/community)XenonDB 技术专区。

- 微信群
  
   ![](docs/images/wechat_group.png)
 