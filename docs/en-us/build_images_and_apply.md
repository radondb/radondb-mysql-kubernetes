## How to build images and apply

### Operator

1. Build operator image and push to your dockerhub.

```
docker build -t {your repo}/mysql-operator:{your tag} . && docker push {your repo}/mysql-operator:{your tag}
```

2. Add radondb mysql helm repo.

```
helm repo add radondb https://radondb.github.io/radondb-mysql-kubernetes/
```

3. Install/Update operator using your image.

```
helm upgrade demo radondb/mysql-operator --install --set manager.image={your repo}/mysql-operator --set manager.tag={your tag}
```

### Sidecar

1. Build sidecar image and push to your dockerhub.

```
docker build -f Dockerfile.sidecar -t {your repo}/mysql-sidecar:{your tag} . && docker push {your repo}/mysql-sidecar:{your tag}
```

2. Create a sample cluster if not exist.

```
kubectl apply -f https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqlcluster.yaml
```

3. Apply image to existing cluster.

```
kubectl patch mysql sample -p '{"spec": {"podPolicy": {"sidecarImage": "{your repo}/mysql-sidecar:{your tag}"}}}' --type=merge
```

> This example is for cluster named `sample`, you can modify as your cluster.

