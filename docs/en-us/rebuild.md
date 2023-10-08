English | [简体中文](../zh-cn/rebuild.md) (Deprecated)

# Reason for rebuilding Pod
RadonDB cluster implements MySQL semisynchronous replication. Semisynchronous replication may cause the replica nodes to update more data than the source node. So, the Pod needs to be rebuilt when it is detected as invalid by Xenon. 

# How to rebuild Pod
Before rebuilding, please manually ensure that the data in the cluster is consistent, and the rebuilding is safe.

```shell
./hack/rebuild.sh PODNAME
```
**Example**
```shell
./hack/rebuild.sh sample-mysql-2
```

# Automatic rebuilding
Automatically rebuild the Pod, taking `sample-mysql-0` as an example:

```shell
kubectl label pods sample-mysql-0 rebuild=true 
```