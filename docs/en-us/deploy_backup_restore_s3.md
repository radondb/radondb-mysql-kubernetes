# mysql-operator

## Quickstart for backup S3

Install the operator named `test`:

```shell
helm install test charts/mysql-operator
```

### configure backup for S3

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
Please add the backupSecretName in mysql_v1alpha1_mysqlcluster.yaml, name as secret file:
```yaml
spec:
  replicas: 3
  mysqlVersion: "5.7"
  backupSecretName: sample-backup-secret
  ...
```
now create backup yaml file mysql_v1alpha1_backup.yaml like this:

```yaml
apiVersion: mysql.radondb.com/v1alpha1
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

### start cluster

```shell
kubectl apply -f config/samples/mysql_v1alpha1_mysqlcluster.yaml     
```
### start backup
After run cluster success
```shell
kubectl apply -f config/samples/mysql_v1alpha1_backup.yaml
```

## Uninstall

Uninstall the cluster named `sample`:

```shell
kubectl delete mysqlclusters.mysql.radondb.com sample
```

Uninstall the operator name `test`:

```shell
helm uninstall test
kubectl delete -f config/samples/mysql_v1alpha1_backup.yaml
```

Uninstall the crd:

```shell
kubectl delete customresourcedefinitions.apiextensions.k8s.io mysqlclusters.mysql.radondb.com
```


## restore cluster from backup copy
check your s3 bucket, get the directory where your backup to， such as `backup_2021720827`.
add  it to RestoreFrom in yaml file
```yaml
...
spec:
  replicas: 3
  mysqlVersion: "5.7"
  backupSecretName: sample-backup-secret
  restoreFrom: "backup_2021720827"
...
```
Then you use:
```shell
kubectl apply -f config/samples/mysql_v1alpha1_mysqlcluster.yaml     
```
could restore a cluster from the `backup_2021720827 ` copy in the S3 bucket. 

if you want backup to NFS server or restore from NFS server, do it as follow:

## Quickstart for  NFS backup

### Create NFS server
first create your PVC, such as "neosan1", fill it to `config/samples/nfs_rc.yaml ` 
```yaml
...
    volumes:
        - name: nfs-export-fast
          persistentVolumeClaim:
            claimName: neosan1
```
 ```shell
# create the nfs pod
kubectl apply -f config/samples/nfs_rc.yaml 
# create the nfs service
kubectl apply -f config/samples/nfs_server.yaml 
 ```
if create the nfs server successful, get the then:

## config `mysql_v1alpha1_backup.yaml ` and backup
Add field in `mysql_v1alpha1_backup.yaml ` as follow:
```yaml
BackupToNFS: "IP of NFS server"
```
use commadn as follow to backup
```shell
kubectl apply -f config/samples/mysql_v1alpha1_backup.yaml
```
 
 ## Restore cluster from exist NFS backup
 first, configure the `mysql_v1alpha1_cluster.yaml`, uncomment the `restoreFromNFS` field:
 ```yaml
 ....
 restoreFrom: "backup_202196152239"
 restoreFromNFS : 10.96.253.82
 ```
 `backup_202196152239` is the nfs server backup path, change it to yours.
 the `10.96.253.82` is the NFS server ip, change it to yours.
 use command as follow to create cluster from NFS server backup copy:

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
