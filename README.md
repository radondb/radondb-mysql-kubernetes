
# RadonDB MySQL
 ![](docs/images/logo_radondb-mysql.png) <br>

English | [中文](README_zh.md) 

## What is RadonDB MySQL

[RadonDB MySQL](https://github.com/radondb/radondb-mysql-kubernetes) is a open-source, cloud-native, highly available cluster solutions that is based on [MySQL](https://MySQL.org) database. With the Raft protocol，RadonDB MySQL provide the faster failover performance without losing any transactions.

At present, RadonDB MySQL has supported the deployment of MySQL high availability clusters on kubernetes and kubesphere platforms.

## Architecture

- Achieving decentralized selection through the Raft protocol.
- Synchronize data based on GTID mode through Semi-Sync.

![](docs/images/radondb-mysql_Architecture_1.png)

## Features

- High availability MySQL database
    - Non-centralized automatic leader selection
    - Second level switch leader to follower 
    - Strongly consistent data for cluster switching
- Cluster management
- Monitoring and alerting
- Logs
- Account management

## Installation

- [Deploy RadonDB MySQL on Kubernetes](docs/Kubernetes/deploy_radondb-mysql_on_kubernetes.md)
- [Deploy RadonDB MySQL on the appstore of KubeSphere](docs/KubeSphere/deploy_radondb-mysql_on_kubesphere.md)

## Release

| Release | Features  | Mode |
|------|--------|--------|
| 1.0 | High availability <br>  Non-centralized automatic leader election <br>  Second level switch <br>  Strongly consistent data <br> Cluster management <br> Monitoring and alerting <br> Logs <br> Account management | Helm |
| 2.0 | Node management <br> Automatic expansion and shrinkage capacity <br> Upgrade <br> Backups and Restorations <br> Automatic failover <br> Automatic rebuild node <br> Automatic restart service（all or signal node）<br> Account management（API）<br> Migrating Data online | Operator |
| 3.0 | Automatic O&M <br> Multiple node roles <br> Disaster Recovery <br> SSL transmission encryption  | Operator |

## Who are using RadonDB MySQL

![](docs/images/users.png)

## License

RadonDB MySQL is released under the Apache 2.0, see [LICENSE](./LICENSE).
## Discussion and Community

- Forum

The RadonDB MySQL topic in [Kubesphere Community](https://kubesphere.com.cn/forum/t/radondb).

- WeChat group

To enter the wechat group of radondb community, please add administrator wechat.

![](docs/images/wechat_admin.jpg)

Please submit any RadonDB MySQL bugs, issues, and feature requests to RadonDB MySQL GitHub Issue.
