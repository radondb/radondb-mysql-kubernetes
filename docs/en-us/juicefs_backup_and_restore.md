English | [简体中文](../zh-cn/juicefs_backup_and_restore.md) | 

目录
=============
 * [juiceopt backup]()
    * [prerequest](#startup-juiceopt-backup)
    * [configuration](#fill-backup-crds-yaml-configs)
    * [running backup](#running-backup)
 
* [restore](#restore)
  * [prepare job](#prerequest-for-restore)
  * [add configmap](#and-config-map)
  * [config mysql cluster](#config-mysql-clusters-yaml)
  * [app the cluster restore from backup](#use-kubectl-apply-the-yaml)

# startup juiceopt backup
## prerequire
 1. prepare S3 strorage （if you want other types of storage,  reference the juicfs document），Obtain the access-key and secret-key . In the example of this article, it use minio ,and the instance is minio-1668754867, and the bucket named `test` so the url is http://test.minio-1668754867.minio:9000/ , you can modify it according your situations, and how to do you can refer to juicefs documents.
    
 2. Install the redis , although juicefs also support sqlite as meta data storage, buf if you do so, you should sync the meta file from s3 at first,and I do not recommend it. redis url is the form as follow：
 ```
 redis://<redis-server-name/IP>:<port>/<NO of database>
 ```
in the example of this article, redis-server-name is  redis-leader, the number of database is 1, So the redis url is  `redis://redis-leader:6379/1`

 3. Verfiy whether it works: suppose the backup directory is juicefs , you can login in  Pod's backup container , execute commanas as follow：
  
  ```
  juicefs format --storage s3 \
    --bucket http://test.minio-1668754867.minio:9000/  \
    --access-key <your access key> \
    --secret-key <your secrete key> > \
    redis://redis-leader:6379/1 \
    juicefs
  ```
 then execute ：
 `juicefs mount -d redis://redis-leader:6379/1 /juicefs` 

check whether juicefs is exist, write files, and check S3 storage whether has changed.

## fill backup crd's yaml configs
In backup crd's yaml file, such as in samples/mysql_v1alpha_backup.yaml, add fields information under spec:

```
  juiceOpt:
    juiceMeta: <fill your redis url>
    backupSecretName: <S3's secret name, for example:sample-backup-secret>
    juiceName: <juicefs backup directory>
```
for example:
```
  juiceOpt:
    juiceMeta: "redis://redis-leader:6379/1"
    backupSecretName: sample-backup-secret
    juiceName: juicefs
```
Others refer to [backup and restore config](./backup_and_restoration_s3.md)

## Running backup.

use command `kubectl apply -f <your backup crd's yaml>` , for examples：

```
 kubectl apply -f config/samples/mysql_v1alpha1_backup.yaml 
``` 

# Restore
## prerequest for restore
  I suppose that the cluster you want restore is `sample2`
###  and `config map`
   1. At first give the `config map` a name,name's form is <name of restore cluster>-restore, this article suppose that cluster name is sample2, so `config map`'s name is  `sample2-restore`
   2. Create config map
    * prepare for juiceopt parameters: 
        build a yaml file, named `juiceopt.yaml`, fill it with:
        ```
        juiceMeta: <redis url>
        backupSecretName: <S3's secret name>
        juiceName: <backup directory under S3 bucket>
        ```
        for example,  in the example of this article, juiceopt.yaml is:
        ```
        juiceMeta: "redis://redis-leader:6379/1"
        backupSecretName: sample-backup-secret
        juiceName: juicefs
        ```
    * use `kubectl create configmap` create a configmap
        configmap has two keys , `from` and `juice.opt`  that respectively indicate the cluster has been backuped which we should restore from, and the juice parameter.
        but `date` key is optional, it indicates the time where restore to (format is:"2006-01-02 09:07:41"), if it does not have got this key , it will restore to now, use the commands as follows:
        `kubectl create configmap sample2-restore --from-literal=from=sample --from-file="juice.opt"=./juiceopt.yaml `


###  config mysql cluster's yaml
     in the example of this article, we suppose the cluster need to restore is sample2, the config method can refer to [radondb cluster configuration](./deploy_radondb-mysql_operator_on_k8s.md)
###  use kubectl apply the yaml
    use `kubectl apply ` apply the yaml file, for the example, use the commands as follow:

    `kubectl apply -f config/samples/mysql_v1alpha1_mysqlcluster.yaml ` 
    