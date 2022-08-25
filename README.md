> English | [简体中文](README_zh.md)


<p align="center">
<a href="https://radondb.com/"><img src="https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/images/logo_radondb-mysql.png?raw=true" alt="banner" width="200px"></a>
</p>
<p align="center">
<b><i>Open-source cloud-native database on Kubernetes</i></b>
</p>

<p align=center>
<a href="https://goreportcard.com/report/github.com/radondb/radondb-mysql-kubernetes"><img src="https://goreportcard.com/badge/github.com/radondb/radondb-mysql-kubernetes" alt="A+"></a>
<a href="https://img.shields.io/github/stars/radondb/radondb-mysql-kubernetes"><img src="https://img.shields.io/github/stars/radondb/radondb-mysql-kubernetes"></a>
<a href="https://img.shields.io/github/issues/radondb/radondb-mysql-kubernetes"><img src="https://img.shields.io/github/issues/radondb/radondb-mysql-kubernetes"></a>
<a href="https://img.shields.io/github/forks/radondb/radondb-mysql-kubernetes"><img src="https://img.shields.io/github/forks/radondb/radondb-mysql-kubernetes"></a>
<a href="https://img.shields.io/github/v/release/radondb/radondb-mysql-kubernetes?include_prereleases"><img src="https://img.shields.io/github/v/release/radondb/radondb-mysql-kubernetes?include_prereleases"></a>
<a href="https://img.shields.io/github/license/radondb/radondb-mysql-kubernetes"><img src="https://img.shields.io/github/license/radondb/radondb-mysql-kubernetes"></a>
<a href="https://app.fossa.com/projects/git%2Bgithub.com%2Fradondb%2Fradondb-mysql-kubernetes?ref=badge_shield" alt="FOSSA Status"><img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2Fradondb%2Fradondb-mysql-kubernetes.svg?type=shield"/></a>
</p>

----

# What is RadonDB MySQL

**RadonDB MySQL** is an open-source, cloud-native, and high-availability cluster solution based on MySQL. It adopts the architecture of one leader node and multiple replicas, with management capabilities for security, automatic backups, monitoring and alerting, automatic scaling, and so on.

**RadonDB MySQL Kubernetes** supports installation, deployment and management of RadonDB MySQL clusters on [Kubernetes](https://kubernetes.io/), [KubeSphere](https://kubesphere.com.cn/) and [Rancher](https://rancher.com), and automates tasks involved in running RadonDB MySQL clusters.

## Features
🧠 **High-availability MySQL**: Automatic decentralized leader election, failover within seconds, and strong data consistency in cluster switching

✏️ **Cluster management**

💻 **Monitoring and alerting**

✍️ [**S3 backups**](docs/en-us/backup_and_restoration_s3.md) and [**NFS backups**](docs/en-us/backup_and_restoration_nfs.md)

🎈 **Log management**

👨 **Account management**

🎨 [**Others**](docs/en-us/)


## Architecture

1. Automatic decentralized leader election by the Raft protocol

2. Synchronizing data by Semi-Sync replication based on GTID mode

3. Supporting high-availability through [Xenon](https://github.com/radondb/xenon.git)

<p align="center">
<a href="https://github.com/radondb/"><img src="https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/images/radondb-mysql_Architecture.png?raw=true" alt="banner" width="800px"></a>
</p>

## Roadmap

| Version | Features  | Mode |
|------|--------|------| 
| 3.0  | Automatic O&M <br> Multiple node roles <br> Disaster recovery <br> SSL transmission encryption | Operator |
| 2.0  | Node management <br> Cluster upgrade <br> Backup and recovery <br> Automatic failover <br> Automatic node rebuilding <br> Account management (API)   |  Operator |
| 1.0 |  Cluster management <br> Monitoring and alerting <br> Log management <br> Account management | Helm |

# Quick start

👀 This tutorial demonstrates how to deploy a RadonDB MySQL cluster (Operator) on Kubernetes.

## Preparation

📦 Prepare a Kubernetes cluster.

## Steps

### Step 1: Add a Helm repository.

```plain
helm repo add radondb https://radondb.github.io/radondb-mysql-kubernetes/
```
### Step 2: Install Operator.

Set the release name to `demo` and create a [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) named `demo-mysql-operator`.

```plain
helm install demo radondb/mysql-operator
```
> **Notice**

> This step also creates the [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) required by the cluster.

### Step 3: Deploy a RadonDB MySQL Cluster.

Run the following command to create an instance of the `mysqlclusters.mysql.radondb.com` CRD and thereby create a RadonDB MySQL cluster by using the default parameters. To customize the cluster parameters, see [Configuration Parameters](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/zh-cn/config_para.md).

```plain
kubectl apply -f https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqlcluster.yaml
```

📖 For more information, see the documentation:

* [Deploy RadonDB MySQL on Kubernetes](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/en-us/deploy_radondb-mysql_operator_on_k8s.md)
* [Deploy RadonDB MySQL on KubeSphere](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/en-us/deploy_radondb-mysql_operator_on_kubesphere.md)
* [Deploy RadonDB MySQL on Rancher](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/docs/en-us/deploy_radondb-mysql_operator_on_rancher.md)
* [All documents](https://radondb.com/en/docs/mysql/)


# Who are using RadonDB MySQL

![](docs/images/%E5%AE%A2%E6%88%B7%E6%A1%88%E4%BE%8B.png)

## License

RadonDB MySQL is based on Apache 2.0 protocol. See [License](https://github.com/radondb/radondb-mysql-kubernetes/blob/main/LICENSE).


[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fradondb%2Fradondb-mysql-kubernetes.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fradondb%2Fradondb-mysql-kubernetes?ref=badge_large)

## Welcome to join us ❤️

😊 Website: [https://radondb.com/](https://radondb.com/en/)

😁 Forum: Please join the [RadonDB](https://kubesphere.com.cn/forum/t/RadonDB) section of kubesphere Developer Forum.

🦉 Community WeChat group: Please add the group assistant **radondb** to invite you into the group.

For any bugs, questions, or suggestions about RadonDB MySQL, please create an [issue](https://github.com/radondb/radondb-mysql-kubernetes/issues) on GitHub or feedback on the [forum](https://kubesphere.com.cn/forum/t/RadonDB).

![Alt](https://repobeats.axiom.co/api/embed/19bb69a6ba32252bdcbdbfb393cfbebd070b3b9f.svg "Repobeats analytics image")