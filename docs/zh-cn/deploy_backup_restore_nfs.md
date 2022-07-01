[English](../en-us/deploy_backup_restore_nfs.md) | 简体中文

# NFS 备份快速手册


## 目录
* [安装 NFS server 与资源](#安装-nfs-server-与资源)
    * [方法一：使用 Helm 安装](#方法一使用-helm-安装)
    * [方法二：使用 kubectl 安装](#方法二使用-kubectl-安装)
* [获取 `nfsServerAddress`](#获取-nfsserveraddress)
* [创建 NFS 备份](#创建-nfs-备份)
    * [1. 配置 NFS server 的地址](#1-配置-nfs-server-的地址)
    * [2. 创建备份](#2-创建备份)
    * [3. 验证备份](#3-验证备份)
* [从已有的 NFS 备份恢复集群](#从已有的-nfs-备份恢复集群)

##  安装 NFS server 与资源

### 方法一：使用 Helm 安装

```shell
helm install demo charts/mysql-operator  --set nfsBackup.installServer=true  --set nfsBackup.volume.createLocalPV=true
```
或者手动创建 PVC，再使用
```shell
helm install demo charts/mysql-operator  --set nfsBackup.installServer=true  --set nfsBackup.volume.specifiedPVC=XXXX
```
> 其中 `XXX` 为 PVC 名称

用该方法，可以在安装 operator 时, 也将 NFS server 的 Pod 和 Service 安装到集群中。

### 方法二：使用 kubectl 安装

```shell
kubectl apply -f config/samples/nfs_pv.yaml 
kubectl apply -f config/samples/nfs_server.yaml 
```

## 获取 `nfsServerAddress`

例如：
```shell

kubectl get svc nfs-server --template={{.spec.clusterIP}}
10.96.253.82
```
获取到 `ClusterIp`，即可以使用该地址进行 NFS 备份。这里 IP 为 `10.96.253.82`。

## 创建 NFS 备份

### 1. 配置 NFS server 的地址

```yaml
# config/samples/mysql_v1alpha1_backup.yaml
nfsServerAddress: "10.96.253.82"
```

### 2. 创建备份
    

```shell
kubectl apply -f config/samples/mysql_v1alpha1_backup.yaml
```
> 注意：备份 CRD 与 MySQL 集群 CRD 必须在同一个命名空间中。

### 3. 验证备份

可以发现形如 `<cluster name>_<timestamp>` 的备份文件夹。如下命令可以查看备份文件夹：

```
kubectl exec -it <pod name of nfs server> -- ls /exports
index.html  initbackup  sample_2022419101946
```

## 从已有的 NFS 备份恢复集群

配置 `mysql_v1alpha1_cluster.yaml`，将 `nfsServerAddress` 设置为 NFS server 的地址。

 ```yaml
 ...
 restoreFrom: "sample_2022419101946"
 nfsServerAddress: 10.96.253.82
 ```
 
 > 注意：`restoreFrom` 是备份的路径名称。可以从 NFS server 加载的路径中看到。

 然后从 NFS 备份副本恢复集群，如下：

 ```
kubectl apply -f config/samples/mysql_v1alpha1_cluster.yaml
 ```