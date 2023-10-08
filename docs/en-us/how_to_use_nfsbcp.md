English | [简体中文](../en-cn/how_to_use_nfsbcp.md)
# Quickstart for nfsbcp （Deprecated）

## Contents
* [Introduction](#Introduction)
* [How to use](#how-to-use)
* [Example](#Example)
     * [Verify the backup](#verify-the-backup)

## Introduction
Create backup based on the command line tool `nfsbcp` to simplify cluster nfs backup steps.
Parameters as follows:

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
## How to use
```shell
cd bin
./nfsbcp backup -s capacity -c cluster -n hostname -p hostpath -b backupImage -i nfsServerImage
```
> `cluster` and `hostpath` are required

The default value of `capacity` is `30`.
The default value of `hostname` is the `leader` node in `cluster`.
The default value of `backupImage` is `radondb/mysql57-sidecar:v2.2.0`.
The default value of `nfsServerImage` is `k8s.gcr.io/volume-nfs:0.8`.
## Example
Create `nfs` backup for the `sample-mysql-0` node of the `sample` cluster. 
The mount address of the `nfs server` is `/mnt/radondb-nfs-backup`, and the capacity of the `nfs server` is `40Gi`. The image of `backup` is `radondb/mysql57-sidecar:vx.x.x`, and the image of `nfs server` is `k8s.gcr.io/volume-nfs:x.x`.
```shell
./nfsbcp backup -s 40 -c sample -n sample-mysql-0 -p /mnt/radondb-nfs-backup -b radondb/mysql57-sidecar:vx.x.x -i k8s.gcr.io/volume-nfs:x.x
```
### Verify the backup
View the backup directory `<cluster name>_<timestamp>` as follows.
```
kubectl exec -it <pod name of nfs server> -- ls /exports
index.html  initbackup  sample_2022419101946
```