# Quickstart for NFS backup
## Install NFS server

### 1. Prepare Storage

Create the NFS PV and SC.

```
kubectl apply -f config/samples/nfs_pv.yaml 
```

> You can specify the PVC you created by modifying the `persistentVolumeClaim.claimName` in the `config/samples/nfs_server.yaml`.

### 2. Create nfs server

Create ReplicationController and Service.

```
kubectl apply -f config/samples/nfs_server.yaml 
```

Get `NFSServerAddress`.

```
# example
kubectl get svc nfs-server --template={{.spec.clusterIP}}
10.96.253.82
```

## Create a NFS backup

### 1. Configure address of NFS Server.

```yaml
# config/samples/mysql_v1alpha1_backup.yaml
NFSServerAddress: "10.96.253.82"
```

### 2. Create a backup

```shell
kubectl apply -f config/samples/mysql_v1alpha1_backup.yaml
```

>  Notice: backup cr and mysqlcluster cr must be in the same namespace.
### 3. Verify your backup

You can find the backup folder called `<cluster name>_<timestamp>`.

```
kubectl exec -it <pod name of nfs server> -- ls /exports
index.html  initbackup  sample_2022419101946
```

 ## Restore cluster from exist NFS backup

Configure the `mysql_v1alpha1_cluster.yaml`, uncomment the `nfsServerAddress` field and fill in your own configuration.

 ```yaml
 ...
 restoreFrom: "sample_2022419101946"
 nfsServerAddress: 10.96.253.82
 ```
 
 > Notice: restoreFrom is the folder name of a backup. You can find it on the path mounted in NFS Server.

 Create cluster from NFS server backup copy:

 ```
kubectl apply -f config/samples/mysql_v1alpha1_cluster.yaml
 ```

 ## Build your own image

 ```
 docker build -f Dockerfile.sidecar -t  acekingke/sidecar:0.1 . && docker push acekingke/sidecar:0.1
 docker build -t acekingke/controller:0.1 . && docker push acekingke/controller:0.1
 ```
> You can replace acekingke/sidecar:0.1 with your own tag

 ## Deploy your own manager
```shell
make manifests
make install 
make deploy  IMG=acekingke/controller:0.1 KUSTOMIZE=~/radondb-mysql-kubernetes/bin/kustomize 
```
