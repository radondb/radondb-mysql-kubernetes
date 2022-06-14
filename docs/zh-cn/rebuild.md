# 为什么需要 rebuild ?

RadonDB 集群是一种半同步的MySQL复制集群, 由于半同步复制存在机会使得slave的数据更新量比master多, 因此当xenon检查到INVALID pod时, 需要重新构建INVALID pod.

# 如何使用 ?
在执行 rebuild 之前, 请手动检查集群数据是否一致, 并且确认 rebuild 动作是安全的

```shell
./hack/rebuild.sh PODNAME
```
**例子**
```shell
./hack/rebuild.sh sample-mysql-2
```