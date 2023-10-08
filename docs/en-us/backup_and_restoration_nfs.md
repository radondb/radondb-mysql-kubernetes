English | [简体中文](../zh-cn/backup_and_restoration_nfs.md)

# Quickstart for NFS backups

## Contents

* [Install NFS server and resources](#install-nfs-server-and-resources)
    * [1. Install by Helm](#1-install-by-helm)
    * [2. Install by kubectl](#2-install-by-kubectl)
* [Obtain `nfsServerAddress`](#obtain-nfsserveraddress)
* [Create an NFS backup](#create-an-nfs-backup)
    * [1. Configure the NFS server address](#1-configure-the-nfs-server-address)
    * [2. Create a backup](#2-create-a-backup)
    * [3. Verify the backup](#3-verify-the-backup)
* [Restore the cluster from the NFS backup](#restore-the-cluster-from-the-nfs-backup)

## Install NFS server and resources

### 1. Install by Helm
```shell
helm install demo charts/mysql-operator  --set nfsBackup.installServer=true  --set nfsBackup.volume.createLocalPV=true
```
Or manually create the PVC and run:
```shell
helm install demo charts/mysql-operator  --set nfsBackup.installServer=true  --set nfsBackup.volume.specifiedPVC=XXXX
```
> `XXX` stands for the PVC name.

In this way, you can install the Pod and Service of the NFS server in the cluster while installing the operator.

### 2. Install by kubectl
```shell
kubectl apply -f config/samples/nfs_pv.yaml 
kubectl apply -f config/samples/nfs_server.yaml 
```

## Obtain `nfsServerAddress`
For example:
```shell

kubectl get svc nfs-server --template={{.spec.clusterIP}}
10.96.253.82
```
You can use `ClusterIp` to perform NFS backup. The cluster IP address in the example is `10.96.253.82`.

## Create an NFS backup
### 1. Configure the NFS server address

```yaml
# config/samples/mysql_v1beta1_backup.yaml
nfsServerAddress: "10.96.253.82"
```

### 2. Create a backup
```shell
kubectl apply -f config/samples/mysql_v1beta1_backup.yaml
```
> Note: The backup CRD and MySQL cluster CRD must be in the same namespace.

### 3. Verify the backup
View the backup directory `<cluster name>_<timestamp>` as follows.

```
kubectl exec -it <pod name of nfs server> -- ls /exports
index.html  initbackup  sample_2022419101946
```

 ## Restore the cluster from the NFS backup

Configure the `nfsServerAddress` attribute to the NFS server address in the `mysql_v1alpha1_cluster.yaml` file or `mysql_v1beta1_cluster.yaml` file.

 ```yaml
 ...
 restoreFrom: "sample_2022419101946"
 nfsServerAddress: 10.96.253.82
 ```
 
 > Notice: `restoreFrom` stands for the path name of the backup. You can get it by checking the path loaded by the NFS server.

Restore cluster from NFS backup as follows.

 ```
kubectl apply -f config/samples/mysql_v1alpha1_cluster.yaml
 ```