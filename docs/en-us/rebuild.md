# Why need rebuild ?
RadonDB cluster is semisynchronous replication mysql cluster. Because MySQL Semisynchronous Replication, It has a chance that the slave has more data than master, So when the xenon check it, It will lable the pod INVALID. When it happend, You
need rebuild the INVALID pod.

# How to use ?
Before you want to rebuild the pod, you need to manually check the security and consistency of the cluster.

```shell
./hack/rebuild.sh PODNAME
```
**for example**
```shell
./hack/rebuild.sh sample-mysql-2
```