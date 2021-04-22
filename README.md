
# XenonDB

English | 中文 

## 什么是 XenonDB？

XenonDB 是基于 MySQL 的开源、高可用、云原生集群解决方案。支持一主多从高可用架构，并具备安全、自动备份、监控告警、自动扩容等全套管理功能。

目前已支持 Kubernetes 和 KubeSphere 平台的部署。

如果您想在以上的两个平台部署 MySQL 高可用集群，XenonDB 可以成为您的选择。 
## 架构图

![](docs/images/XenonDB_Architecture_1.png)

- 通过 Raft 协议实现无中心化选主
- 通过 Semi-Sync 基于 GTID 模式同步数据
## 核心功能
- MySQL高可用
    - 无中心化自动选主
    - 主从秒级别切换
    - 集群切换的数据强一致性
- 集群部署/销毁
- 集群配置变更
- 监控告警
- 查看集群日志
- 账户管理
## 快速部署
Kubernetes 平台部署（制作中）

KubeSphere 应用商店部署【链接补充】
## 文档
待补充
## 产品路线

| 版本 | 1.0  | 2.0  | 3.0   |
|------|--------|--------|---------|
|  实<br>现<br>方<br>式  | Helm  | Operator | Operator |
| 实<br>现<br>功<br>能  | MySQL 高可用<br>     无中心化自动选主<br>     主从秒级别切换<br>     集群切换的数据强一致性<br>集群部署 / 销毁<br>集群配置变更<br>监控告警<br>查看集群日志<br>账户管理 |  增删节点<br>集群扩缩容（手动 / 自动）<br>升级集群 / Operator<br>集群备份<br>备份恢复<br>集群故障自动转移、重建节点、重启服务（所有节点 / 单节点）<br>账户管理（提供接口）<br>在线迁移<br>配置变更细化  |自动化运维<br>多节点角色支持（ Proxy 实例 / 主实例 / 只读实例）<br>灾备集群<br>SSL 传输加密 |

## 客户

![](docs/images/users.png)
## 协议

XenonDB 基于 Apache 2.0 协议。具体详见 [LICENSE](./LICENSE) 文件。

## 欢迎加入社区话题互动
XenonDB 

- 论坛：【KubeSphere 开发者社区 | XenonDB 专题 】
- 微信群

![](docs/images/wechat_group.png)
