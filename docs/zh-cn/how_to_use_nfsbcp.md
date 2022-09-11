[English](../en-us/how_to_use_nfsbcp.md) | 简体中文
# NFS一键备份工具使用手册

## 目录
* [简介](#简介)
* [使用方法](#使用方法)
* [例子](#例子)
    * [验证备份](#验证备份)

## 简介
基于命令行工具`nfsbcp`，可实现对集群的`nfs`一键备份。
具体参数如下：

```shell
Create a nfs backup for RadonDB MySQL cluster

Usage:
  nfsbcp backup [flags]

Flags:
  -b, --backupImage string      The image of backup. (default "radondb/mysql57-sidecar:v2.2.0")
  -s, --capacity string         The capacity of nfs server. (default "30")
  -c, --cluster string          The cluster name to backup.
  -h, --help                    help for backup
  -n, --hostname string         The host for which to take backup.
  -p, --hostpath string         The local storage path of nfs.
  -i, --nfsServerImage string   The image of backup. (default "k8s.gcr.io/volume-nfs:0.8")
```
## 使用方法
```shell
cd bin
./nfsbcp backup -s capacity -c cluster -n hostname -p hostpath -b backupImage -i nfsServerImage
```
> 其中`cluster`、`hostpath`为必填项

`capacity`默认为`30`
`hostname`默认为`cluster`中`leader`节点
`backupImage`默认为`radondb/mysql57-sidecar:v2.2.0`
`nfsServerImage`默认为`radondb/mysql57-sidecar:v2.2.0`
## 例子
为`sample`集群的`sample-mysql-0`节点创建`nfs`备份，`nfs server`挂载地址为`/mnt/radondb-nfs-backup`，容量为`40Gi`，`backup`镜像为`radondb/mysql57-sidecar:vx.x.x`，`nfs server`镜像为`k8s.gcr.io/volume-nfs:x.x`
```shell
./nfsbcp backup -s 40 -c sample -n sample-mysql-0 -p /mnt/radondb-nfs-backup -b radondb/mysql57-sidecar:vx.x.x -i k8s.gcr.io/volume-nfs:x.x
```
### 验证备份
可以发现形如 `<cluster name>_<timestamp>` 的备份文件夹。如下命令可以查看备份文件夹：

```
kubectl exec -it <pod name of nfs server> -- ls /exports
index.html  initbackup  sample_2022419101946
```