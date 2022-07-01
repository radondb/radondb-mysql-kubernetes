[English](../en-us/build_and_use_images.md) | 简体中文

## 如何创建和应用自己的镜像
   * [Operator](#operator)
   * [Sidecar](#sidecar)

### Operator

1. 创建自己的镜像并上传到 Docker Hub 上。

```
docker build -t {your repo}/mysql-operator:{your tag} . && docker push {your repo}/mysql-operator:{your tag}
```

2. 添加 RadonDB MySQL 的 Helm 库。

```
helm repo add radondb https://radondb.github.io/radondb-mysql-kubernetes/
```

3. 使用自己的镜像来安装或更新 Operator。

```
helm upgrade demo radondb/mysql-operator --install --set manager.image={your repo}/mysql-operator --set manager.tag={your tag}
```

### Sidecar

1. 创建自己的 sidecar 镜像并推送到 Docker Hub 中。

```
docker build -f Dockerfile.sidecar -t {your repo}/mysql-sidecar:{your tag} . && docker push {your repo}/mysql-sidecar:{your tag}
```

2. 创建一个集群。

```
kubectl apply -f https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqlcluster.yaml
```

3. 在已有的集群中应用自己的镜像。

```
kubectl patch mysql sample -p '{"spec": {"podPolicy": {"sidecarImage": "{your repo}/mysql-sidecar:{your tag}"}}}' --type=merge
```

> 本例中集群名称为 `sample`，您可以修改为您自己的集群名称。