## 如何创建自己的镜像
   * [Operator](#operator)
      * [Sidecar](#sidecar)

### Operator

1. 创建自己的镜像并上传到docker hub上.

```
docker build -t {your repo}/mysql-operator:{your tag} . && docker push {your repo}/mysql-operator:{your tag}
```

2. 添加 radondb mysql 的 helm 库.

```
helm repo add radondb https://radondb.github.io/radondb-mysql-kubernetes/
```

3.  使用自己的镜像来安装/更新 operator.

```
helm upgrade demo radondb/mysql-operator --install --set manager.image={your repo}/mysql-operator --set manager.tag={your tag}
```

### Sidecar

1. 创建自己的 sidecar 镜像并 push 到 docker hub 中.

```
docker build -f Dockerfile.sidecar -t {your repo}/mysql-sidecar:{your tag} . && docker push {your repo}/mysql-sidecar:{your tag}
```

2. 创建sample cluster.

```
kubectl apply -f https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqlcluster.yaml
```

3. 从已有的 cluster 中应用自己的镜像.

```
kubectl patch mysql sample -p '{"spec": {"podPolicy": {"sidecarImage": "{your repo}/mysql-sidecar:{your tag}"}}}' --type=merge
```

> 本例中集群名称为 `sample`, 您可以修改为您自己的集群名称.

