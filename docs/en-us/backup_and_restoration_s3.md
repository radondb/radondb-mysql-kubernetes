English | [简体中文](../zh-cn/backup_and_restoration_s3.md)

# Quickstart for S3 backups

## Contents

  - [Prerequisites](#Prerequisites)
  - [Overview](#Overview)
  - [Configure the backup](#Configure-the-backup)
    - [Step 1: Create the Secret file](#Step-1-Create-the-Secret-file)
    - [Step 2: Configure the Secret for the Operator cluster](#Step-2-Configure-the-Secret-for-the-Operator-cluster)
  - [Start the backup](#Start-the-backup)
  - [Restore the cluster from the backup](#Restore-the-cluster-from-the-backup)

## Prerequisites
* You need to deploy the [RadonDB MySQL cluster](./deploy_radondb-mysql_operator_on_k8s.md).

## Overview
This tutorial displays how to back up and restore the deployed RadonDB MySQL Operator cluster.

## Configure the backup

### Step 1: Create the Secret file
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
The value `s3-xxxx` is base64-encoded. You can encode the value as follows, and do not encode line breaks.

```shell
echo -n "your value"|base64
```
Then, create the backup Secret.
```
kubectl create -f config/samples/backup_secret.yaml
```

### Step 2: Configure the backup Secret for the Operator cluster
Configure the `backupSecretName` property in `mysql_v1alpha1_mysqlcluster.yaml`, for example, `sample-backup-secre`.

```yaml
spec:
  replicas: 3
  mysqlVersion: "5.7"
  backupSecretName: sample-backup-secret
  ...
```
Create the backup YAML file `mysql_v1alpha1_backup.yaml` as follows.

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
| Parameter   | Description  |
| ----------- | ------------ |
| hostName    | Pod name     |
| clusterName | Cluster name |

## Start the backup

Before starting the backup, you need to start the cluster.

```shell
kubectl apply -f config/samples/mysql_v1alpha1_backup.yaml
```

After starting the backup successfully, view the backup status as follows.

```
kubectl get backups.mysql.radondb.com 
NAME            BACKUPNAME             BACKUPDATE            TYPE
backup-sample   sample_2022526155115   2022-05-26T15:51:15   S3
```

## Restore the cluster from the backup
Check the S3 bucket and set the `RestoreFrom` property in the `mysql_v1alpha1_mysqlcluster.yaml` file to the backup directory, for example, `sample_2022526155115`.

```yaml
...
spec:
  replicas: 3
  mysqlVersion: "5.7"
  backupSecretName: sample-backup-secret
  restoreFrom: "sample_2022526155115"
...
```
Then, start the cluster and the database will be restored from the backup directory.

```shell
kubectl apply -f config/samples/mysql_v1alpha1_mysqlcluster.yaml     
```
