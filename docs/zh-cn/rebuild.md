[English](../en-us/rebuild.md) | 简体中文

# 为什么需要重新构建 Pod？

RadonDB 集群是一种半同步的MySQL复制集群。在某些情况下，半同步复制可能使从节点的数据更新量比主节点多，因此当 Xenon 检查到 INVALID Pod 时，需要重新构建该 Pod。

# 如何使用？
在执行 rebuild 之前，请手动检查集群数据是否一致，并且确认 rebuild 动作是安全的。

```shell
./hack/rebuild.sh PODNAME
```
**例子**
```shell
./hack/rebuild.sh sample-mysql-2
```