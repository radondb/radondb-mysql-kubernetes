English | [简体中文](../zh-cn/build_and_use_images.md)

## How to build and use images
   * [Operator](#operator)
   * [Sidecar](#sidecar)

### Operator

1. Build an operator image and push it to Docker Hub.

```
docker build -t {your repo}/mysql-operator:{your tag} . && docker push {your repo}/mysql-operator:{your tag}
```

2. Add the Helm repository of RadonDB MySQL.

```
helm repo add radondb https://radondb.github.io/radondb-mysql-kubernetes/
```

3. Install/Update the operator using the image.

```
helm upgrade demo radondb/mysql-operator --install --set manager.image={your repo}/mysql-operator --set manager.tag={your tag}
```

### Sidecar

1. Build a sidecar image and push it to Docker Hub.

```
docker build -f Dockerfile.sidecar -t {your repo}/mysql-sidecar:{your tag} . && docker push {your repo}/mysql-sidecar:{your tag}
```

2. Create a sample cluster.

```
kubectl apply -f https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqlcluster.yaml
```

3. Apply the image to the cluster.

```
kubectl patch mysql sample -p '{"spec": {"podPolicy": {"sidecarImage": "{your repo}/mysql-sidecar:{your tag}"}}}' --type=merge
```

> You can use your own cluster name to replace `sample`.