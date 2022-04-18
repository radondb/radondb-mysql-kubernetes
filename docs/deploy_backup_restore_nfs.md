# mysql-operator

## Quickstart for  NFS backup
### Create NFS server
First create your PVC, such as "neosan1", 
Or use local storage pvc, see `config/sample/pv-claim.yaml`
and `pv-volume.yaml` for more details.
Then create NFS server, such as "nfs-server",
```yaml
fill it to `config/samples/nfs_rc.yaml ` 
```yaml
...
    volumes:
        - name: nfs-export-fast
          persistentVolumeClaim:
            claimName: neosan1 // or backup-pv-claim
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
use command as follow to backup
```shell
kubectl apply -f config/samples/mysql_v1alpha1_backup.yaml
```
>  Notice: backup cr and mysqlcluster cr must be in the same namespace.
 ## Restore cluster from exist NFS backup
 first, configure the `mysql_v1alpha1_cluster.yaml`, uncomment the `restoreFromNFS` field:
 ```yaml
 ....
 restoreFrom: "sample_202196152239"
 restoreFromNFS : 10.96.253.82
 ```
 `sample_202196152239` is the nfs server backup path, change it to yours.
 the `10.96.253.82` is the NFS server ip, change it to yours.
 
 > Notice: you can find the `sample_202196152239` in the nfs_server pod, at `/exports` path
 or  find it in node `/mnt/backup` path if you use the local pesistent volume with `sample/pv-volume.yaml`.

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
