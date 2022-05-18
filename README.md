> English | [简体中文](README_zh.md)


<p align="center">
<a href="https://radondb.com/"><img src="https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/images/logo_radondb-mysql.png?raw=true" alt="banner" width="200px"></a>
</p>
<p align="center">
<b><i>Cloud-Native and Databases on Kubernetes</i></b>
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

# What is RadonDB MySQL

**RadonDB MySQL** is an open-source, cloud-native, highly availability cluster solutions based on MySQL. With the Raft protocol，RadonDB MySQL provides faster failover performance without losing any transactions.

**RadonDB MySQL Kubernetes** supports deployment and management of RaodnDB MySQL clusters on [Kubernetes](https://kubernetes.io/), [KubeSphere](https://kubesphere.com.cn/) and [Rancher](https://rancher.com) automates tasks related to operating a RadonDB MySQL cluster.

## Features
🧠 **High Availability MySQL**: Non-centralized automatic leader selection, Leader-follower switching in second-level, Strongly consistent data for cluster switching
 
✏️ **Cluster Management**

💻 **Monitoring and Alerting**

✍️ [**Backup for S3**](docs/en-us/deploy_backup_restore_s3.md)

🎈 **Log Management**

👨 **Account Management**

🎨 [**Others**](docs/en-us/)


## Architecture

1. Decentralized leader automatic election through Raft protocol.

2. Synchronizing data based on GTID mode through Semi-Sync.

3. Supporting high-availability through [Xenon](https://github.com/radondb/xenon.git).

<p align="center">
<a href="https://github.com/radondb/"><img src="https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/images/radondb-mysql_Architecture.png?raw=true" alt="banner" width="800px"></a>
</p>

## Roadmap

| Release | Features  | Mode |
|------|--------|------| 
| 3.0  | Automatic O&M <br> Multiple Node Roles <br> Disaster Recovery <br> SSL Transmission Encryption | Operator |
| 2.0  | Node Management <br> Upgrade Cluster <br> Backups and Restorations <br> Automatic Failover <br> Automatic Rebuild Node <br> Account management（API）   |  Operator |
| 1.0 |  Cluster Management <br> Monitoring and Alerting <br> Log Management <br> Account management | Helm |

# Quick Start

👀 Demonstrate how to deploy RadonDB MySQL Cluster (operator) on Kubernetes.

## Prepare

📦 You need to prepare a Kubernetes cluster.

## Deployment steps

### Step 1: Add a Helm Repository

```plain
helm repo add radondb https://radondb.github.io/radondb-mysql-kubernetes/
```
### Step 2: Install Operator

The following sets the release name to `demo` and creates a [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) named `demo-mysql-operator`.

```plain
helm install demo radondb/mysql-operator
```
> **Notic**
> 
> This step also creates the [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) required by the cluster.

### Step 3: Install a RadonDB MySQL Cluster

Run the following command to create an instance of the `mysqlclusters.mysql.radondb.com` CRD and thereby create a RadonDB MySQL cluster by using the default parameters. You can customize the cluster parameters by referring to [Configuration Parameters](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/zh-cn/config_para.md).

```plain
kubectl apply -f https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqlcluster.yaml
```


📖 For more information, see the documentation:

* [Deploy RadonDB MySQL on Kubernetes](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/en-us/deploy_radondb-mysql_operator_on_k8s.md)
* [Deploy RadonDB MySQL on Kubernetes](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/en-us/deploy_radondb-mysql_operator_on_kubesphere.md)
* [Deploy RadonDB MySQL on Rancher](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/en-us/deploy_radondb-mysql_operator_on_rancher.md)
* [All Documents](https://radondb.com/en/docs/mysql/)


# Who are using RadonDB MySQL

![](docs/images/%E5%AE%A2%E6%88%B7%E6%A1%88%E4%BE%8B.png)

## License

RadonDB MySQL is based on Apache 2.0 protocol. See [License](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/LICENSE)。

## Welcome to join us ❤️

😊 Website: [https://radondb.com/](https://radondb.com/en/)

😁 Forum：Please join the [RadonDB section](https://kubesphere.com.cn/forum/t/RadonDB) of kubesphere Developer Forum.

🦉 Community wechat group：Please add a group assistant **radondb** to invite you into the group
