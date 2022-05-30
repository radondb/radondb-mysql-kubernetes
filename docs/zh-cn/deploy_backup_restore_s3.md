
Contents
=============

  - [前提条件](#前提条件)
  - [简介](#简介)
  - [配置backup](#配置backup)
    - [步骤 1: 添加 secret 文件](#1-添加-secret-文件)
    - [步骤 2: 将 secret 配置到 Operator 集群](#2-将-secret-配置到-Operator-集群)
  - [启动备份](#启动备份)
  - [从备份副本恢复到新集群](#从备份副本恢复到新集群)

## 前提条件

* 已部署 [RadonDB MySQL 集群](kubernetes/deploy_radondb-mysql_operator_on_k8s.md)。

## 简介

本文档介绍如何对部署的 RadonDB MySQL Operator 集群进行备份和恢复。

## 配置backup

### 1. 添加 secret 文件
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: sample-backup-secret
  namespace: default
data:
  s3-endpoint: aHR0cDovL3MzLnNoMWEucWluZ3N0b3IuY29t
  s3-access-key: SEdKWldXVllLSENISllFRERKSUc=
  s3-secret-key: TU44TkNUdDJLdHlZREROTTc5cTNwdkxtNTlteE01blRaZlRQMWxoag==
  s3-bucket: bGFsYS1teXNxbA==
type: Opaque

```
s3-xxxx 值用 base64 编码,注意不要把换行符编码进去, 可以用如下命令获取base64编码:
```
echo -n "替换为您的S3-XXX值"|base64
```
随后, 用如下命令创建备份secret:

```
kubectl create -f config/samples/backup_secret.yaml
```
### 2. 将 secret 配置到 Operator 集群
将备份secret名称添加到mysql_v1alpha1_mysqlcluster.yaml中, 本例中名称为sample-backup-secret:

```yaml
spec:
  replicas: 3
  mysqlVersion: "5.7"
  backupSecretName: sample-backup-secret
  ...
```
now create backup yaml file mysql_v1alpha1_backup.yaml like this:

```yaml
apiVersion: mysql.radondb.com/v1alpha1
kind: Backup
metadata:
  name: backup-sample1
spec:
  # Add fields here
  hostName: sample-mysql-0
  clusterName: sample

```
| name | function  | 
|------|--------|
|hostName| 集群中Pod的名称 |
|clusterName| 数据库集群名称 |


##  启动备份
启动集群后,才可以进行备份操作
```shell
kubectl apply -f config/samples/mysql_v1alpha1_backup.yaml
```
执行成功后, 可以通过如下命令查看备份状况
```
kubectl get backups.mysql.radondb.com 
NAME            BACKUPNAME             BACKUPDATE            TYPE
backup-sample   sample_2022526155115   2022-05-26T15:51:15   S3
```

## 从备份副本恢复到新集群
检查您的 s3 bucket, 得到您需要的备份文件夹如 `sample_2022526155115`.
添加 RestoreFrom 字段到 mysql_v1alpha1_backup.yaml 中, 如下:

```yaml
...
spec:
  replicas: 3
  mysqlVersion: "5.7"
  backupSecretName: sample-backup-secret
  restoreFrom: "sample_2022526155115"
...
```
随后, 启动集群, 将会从备份文件夹中恢复数据库.:
```shell
kubectl apply -f config/samples/mysql_v1alpha1_mysqlcluster.yaml     
```
