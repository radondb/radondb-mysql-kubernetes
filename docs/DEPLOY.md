# mysql-operator

## Quickstart

Install the operator named `test`:

```shell
helm install test https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql-operator.tgz
```

Then install the cluster named `sample`:
```shell
kubectl create -f config/samples/backup_secret.yaml
```

```shell
kubectl apply -f config/samples/mysql_v1alpha1_cluster.yaml     
```
## backup
After run cluster success
```shell
kubectl apply -f config/samples/samples/mysql_v1alpha1_backup.yaml
```

## Uninstall

Uninstall the cluster named `sample`:

```shell
kubectl delete clusters.mysql.radondb.com sample
```

Uninstall the operator name `test`:

```shell
helm uninstall test
```

Uninstall the crd:

```shell
kubectl delete customresourcedefinitions.apiextensions.k8s.io clusters.mysql.radondb.com
```

## configure backup

add the secret file
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: sample-backup-secret
  namespace: default
data:
  s3-endpoint: aHR0cDovL3MzLnNoMWEucWluZ3N0b3IuY29t
  s3-access-key: SEdKWldXVllLSENISllFRERKSUc=
  s3-secret-key: TU44TkNUdDJLdHlZREROTTc5cTNwdkxtNTlteE01blRaZlRQMWxoag==
  s3-bucket: bGFsYS1teXNxbA==
type: Opaque

```
s3-xxxx value is encode by base64, you can get like that
```shell
echo -n "hello"|base64
```
then, create the secret in k8s.
```
kubectl create -f config/samples/backup_secret.yaml
```
Please add the backupSecretName in mysql_v1_cluster.yaml, name as secret file:
```yaml
spec:
  replicas: 3
  mysqlVersion: "5.7"
  backupSecretName: sample-backup-secret
  ...
```
now create backup yaml file mysql_v1_backup.yaml like this:

```yaml
apiVersion: mysql.radondb.io/v1
kind: Backup
metadata:
  name: backup-sample1
spec:
  # Add fields here
  hostname: sample-mysql-0
  clustname: sample

```
| name | function  | 
|------|--------|
|hostname|pod name in cluser|
|clustname|cluster name|

## start backup
After cluster has started, if you want backup:
```shell
 kubectl apply -f config/samples/mysql_v1_backup.yaml
 ```

 ## build your own image
 such as :
 ```
 docker build -f Dockerfile.sidecar -t  acekingke/sidecar:0.1 . && docker push acekingke/sidecar:0.1
 docker build -t acekingke/controller:0.1 . && docker push acekingke/controller:0.1
 ```
 you can replace acekingke/sidecar:0.1 with your own tag

 ## deploy your own manager
```shell
make manifests
make install 
make deploy  IMG=acekingke/controller:0.1 KUSTOMIZE=~/radondb-mysql-kubernetes/bin/kustomize 
```
