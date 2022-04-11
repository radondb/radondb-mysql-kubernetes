# Deploy
use sample `config/samples/mysql_v1alpha1_mysqlcluster_mysql8.yaml` , modify the `spec.podPolicy.sidecarImage` to MySQL8 version sidecar image.
install step 
[deploy reference](deploy_radondb-mysql_operator_on_kubesphere.md) 
# How to build sidecar images?
After you had installed docker , use command:
```
make mysql8-sidecar
```
it will build mysql8 sidecar image.Then `docker push xxx` to push to docker hub.