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
kubectl apply -f config/samples/samples/mysql_v1alpha1_mysqlbackup.yaml
```

## Uninstall

Uninstall the cluster named `sample`:

```shell
kubectl delete clusters.mysql.radondb.io sample
```

Uninstall the operator name `test`:

```shell
helm uninstall test
```

Uninstall the crd:

```shell
kubectl delete customresourcedefinitions.apiextensions.k8s.io clusters.mysql.radondb.io
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

now create mysqlbackup yaml file mysql_v1_mysqlbackup.yaml like this:

```yaml
apiVersion: mysql.radondb.io/v1
kind: MysqlBackup
metadata:
  name: mysqlbackup-sample1
spec:
  # Add fields here
  hostname: sample-mysql-0
  clustname: sample



```

## start backup
After cluster has started, if you want backup:
```shell
 kubectl apply -f config/samples/mysql_v1_mysqlbackup.yaml
 ```

 ## build your own image
 such as :
 ```
 docker build -f Dockerfile.sidecar -t  acekingke/sidecar:v01 . && docker push acekingke/sidecar:v01
 docker build -t acekingke/controller:v01 . && docker push acekingke/controller:v01
 ```
 you can replace acekingke/sidecar:v01 with your own tag

 ## deploy your own manager
```shell
make manifests
make install 
make deploy  IMG=acekingke/controller:v01 KUSTOMIZE=~/radondb-mysql-kubernetes/bin/kustomize 
```
