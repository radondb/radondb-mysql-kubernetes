[English](../en-us/juicefs_backup_and_restore.md) | 简体中文

目录
=============
 * [juiceopt备份](#开启juiceopt-备份)
    * [准备工作](#准备工作)
    * [填写配置](#填写backup-crd-的-yaml-配置信息)
    * [运行备份](#运行备份)
 
* [恢复](#恢复)
  * [准备工作](#恢复集群的准备工作)
  * [添加configmap](#添加config-map-配置)
  * [配置mysql cluster](#配置mysql-cluster-的yaml)
  * [应用新集群yaml](#使用kubectl-apply-应用yaml)

# 开启juiceopt 备份
## 准备工作
 1. 准备S3存储 （其他类型存储参见 juicefs 文档），获取 access-key 和 secret-key 本例中采用minio 安装,实例为 minio-1668754867, bucket 名为 test 的url 连接为 http://test.minio-1668754867.minio:9000/ 可以依据具体情况修改，参见 juicefs 官方文档
    
 2. 安装 redis , 虽然 juicefs 也支持使用 sqlite 作为元数据存储, 但是需要每次从 S3 存储的元数据备份中下载元数据文件, 因此不推荐用 sqlite 作为元数据看存储数据库. redis 的连接组成方式如下：
 ```
 redis://<redis-服务名/IP>:<端口>/<数据库编号>
 ```
 在本文挡中, redis-服务名为  redis-leader, 数据库号为1, 所以redis 连接串为 `redis://redis-leader:6379/1`

 3. 验证可用性: 假设备份的文件夹为 juicefs , 可以直接登录集群 Pod 的 backup 容器, 执行如下命令：
  
  ```
  juicefs format --storage s3 \
    --bucket http://test.minio-1668754867.minio:9000/  \
    --access-key <your access key> \
    --secret-key <your secrete key> > \
    redis://redis-leader:6379/1 \
    juicefs
  ```
 再执行：
 `juicefs mount -d redis://redis-leader:6379/1 /juicefs` 

查看当前目录是否存在juicefs 文件夹, 写入文件, 检查 S3 存储是否有变化

## 填写backup crd 的 yaml 配置信息
在 backup crd 的yaml 文件, 如 samples/mysql_v1alpha_backup.yaml 中, spec 字段下添加如下信息:

```
  juiceOpt:
    juiceMeta: <填写你的 redis url 信息>
    backupSecretName: <S3 的密钥secret名, 如sample-backup-secret>
    juiceName: <juicefs 备份文件夹>
```
示例:
```
  juiceOpt:
    juiceMeta: "redis://redis-leader:6379/1"
    backupSecretName: sample-backup-secret
    juiceName: juicefs
```
其他信息参见[备份与恢复配置](./backup_and_restoration_s3.md)

## 运行备份

使用 kubectl apply -f <你的备份yaml> , 如：

```
 kubectl apply -f config/samples/mysql_v1alpha1_backup.yaml 
``` 

# 恢复
## 恢复集群的准备工作
  假设要恢复的集群名称为sample2
###  添加config map 配置
   1. 首先给config map一个名字, 名字为 <恢复集群名称>-restore, 本文中假设要恢复的集群为sample2, 所以config map的名称为
      sample2-restore
   2. 创建config map
    * juiceopt 参数准备: 
        准备一个yaml文件, 名为juiceopt.yaml, 填写内容如下
        ```
        juiceMeta: <redis连接信息>
        backupSecretName: <S3的secret名称>
        juiceName: <S3 bucket下备份文件夹名称>
        ```
        例如: 本文示例中, juiceopt.yaml内容如下:
        ```
        juiceMeta: "redis://redis-leader:6379/1"
        backupSecretName: sample-backup-secret
        juiceName: juicefs
        ```
    * 使用`kubectl create configmap` 创建configmap
        config map 中必须的两个key 为 `from` 和 `juice.opt`  分别表示已经备份集群名称和juice备份的参数
        而 `date` key是可选的, 表示需要恢复的时间点(格式为:"2006-01-02 09:07:41"), 如果不选,则以当前时间点为准恢复, 使用如下命令执行:
        `kubectl create configmap sample2-restore --from-literal=from=sample --from-file="juice.opt"=./juiceopt.yaml `


###  配置mysql cluster 的yaml
     本例中, 需要恢复的集群为sample2, 配置方法参见[radondb cluster配置方法](./deploy_radondb-mysql_operator_on_k8s.md)
###  使用kubectl apply 应用yaml
    使用 `kubectl apply ` 应用yaml 文件, 本例中,使用:

    `kubectl apply -f config/samples/mysql_v1alpha1_mysqlcluster.yaml ` 
    