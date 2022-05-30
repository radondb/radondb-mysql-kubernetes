# NFS 备份快速手册

## 安装 NFS server 与资源

### 方法一： 使用 helm 安装

```shell
helm install demo charts/mysql-operator  --set nfsBackup.installServer=true  --set nfsBackup.volume.createLocalPV=true
```
或者手动创建PVC, 再使用
```shell
helm install demo charts/mysql-operator  --set nfsBackup.installServer=true  --set nfsBackup.volume.specifiedPVC=XXXX
```
> 其中 XXX 为 pvc 名称

用该方法, 可以在安装operator的时候, 也将 NFS server 的 Pod 和 Service 安装到集群中.

### 方法二： 使用 kubectl 安装

```shell
kubectl apply -f config/samples/nfs_pv.yaml 
 kubectl apply -f config/samples/nfs_server.yaml 
```

## 获取 `nfsServerAddress`.

例如
```shell

kubectl get svc nfs-server --template={{.spec.clusterIP}}
10.96.253.82
```
获取到ClusterIp, 即可以使用该地址进行 NFS backup. 这里 IP 为 `10.96.253.82`

## 创建 NFS backup

### 1. 配置 NFS server 的地址

```yaml
# config/samples/mysql_v1alpha1_backup.yaml
nfsServerAddress: "10.96.253.82"
```

### 2. 创建一个backup
    

```shell
kubectl apply -f config/samples/mysql_v1alpha1_backup.yaml
```
>  注意: backup cr 与 mysqlcluster cr 必须在同一个 namespace 中.

### 3. 验证备份

可以发现备份形如 `<cluster name>_<timestamp>` 的备份文件夹. 如下命令可以查看备份文件夹:

```
kubectl exec -it <pod name of nfs server> -- ls /exports
index.html  initbackup  sample_2022419101946
```

 ## 从已有的 NFS backup 恢复集群

配置 `mysql_v1alpha1_cluster.yaml`, 将 `nfsServerAddress` 设置为 NFS server 的地址.

 ```yaml
 ...
 restoreFrom: "sample_2022419101946"
 nfsServerAddress: 10.96.253.82
 ```
 
 > 注意: restoreFrom 是备份的路径名称. 可以从 nfs server 加载的路径中看到.

 然后从 NFS 备份副本恢复集群, 如下:

 ```
kubectl apply -f config/samples/mysql_v1alpha1_cluster.yaml
 ```