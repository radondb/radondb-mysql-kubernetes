[English](../en-us/manage_mysql_user.md) | 简体中文

目录
==========
   * [1. 前提条件](#1.-前提条件)
   * [2. 创建用户帐号](#2.-创建用户帐号)
      * [2.1 校验 CRD](#2.1-校验-CRD)
      * [2.2 创建用户](#2.2-创建用户)
      * [2.3 查看用户](#2.3-查看用户)
   * [3. 登录用户](#3.-登录用户)
   * [4. 删除用户](#4.-删除用户)
   * [5. 参数说明](#5.-参数说明)

# 使用 MysqlUser CRD 管理 MySQL 用户

##  1. 前提条件

* 已部署 [RadonDB MySQL 集群](kubernetes/deploy_radondb-mysql_operator_on_k8s.md)。

##  2. 创建用户帐号

###  2.1 校验 CRD

使用如下指令，将查看到名称为 `mysqlusers.mysql.radondb.com` 的 CRD。

```plain
kubectl get crd | grep mysqluser
mysqlusers.mysql.radondb.com                          2021-09-21T09:15:08Z
```

###  2.2 创建用户

使用如下指令，将创建一个名为 `normal_user` 的普通用户和一个名为 `super_user` 的超级用户。用户密码保存在名为 `sample-user-password` 的 [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) 中。

```plain
kubectl apply -f https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqluser.yaml 
```

> 示例中创建的普通用户和超级用户的密码均为 RadonDB@123

###  2.3 查看用户

```
kubectl get mysqluser -o wide                                                                                      
NAME          USERNAME      SUPERUSER   HOSTS   TLSTYPE   CLUSTER   NAMESPACE   AVAILABLE   SECRETNAME             SECRETKEY
normal-user   normal_user   false       ["%"]   NONE      sample    default     True        sample-user-password   normalUser
super-user    super_user    true        ["%"]   NONE      sample    default     True        sample-user-password   superUser
```

## 3. 登录用户

使用如下指令，使用 `super_user` 用户连接到 MySQL 集群主节点。

```
kubectl exec -it svc/sample-leader -c mysql -- mysql -usuper_user -pRadonDB@123
```

##  4. 删除用户

使用如下指令将删除示例中创建的 MysqlUser CR 和对应的用户。

```plain
kubectl delete mysqluser normal-user super-user
```

##  5. 参数说明

| 参数                      | 描述                                       |
| ------------------------- | ------------------------------------------ |
| user                      | 用户名                                     |
| hosts                     | 允许访问的主机列表；%表示全部              |
| withGrantOption           | 用户是否能够给其他用户授权；默认为 false   |
| tlsOptions.type           | TLS类型；必须为 NONE/SSL/X509；默认为 NONE |
| permissions.database      | 授权的数据库；*表示全部                    |
| permissions.tables        | 授权的表；*表示全部                        |
| permissions.privileges    | 授权项                                     |
| userOwner.clusterName     | 用户所在集群的名称                         |
| userOwner.nameSpace       | 用户所在集群的命名空间                     |
| secretSelector.secretName | 保存用户密码的 secret 名称                 |
| secretSelector.secretKey  | 保存用户密码的 secret key                  |

> 详情请参考 https://dev.mysql.com/doc/refman/5.7/en/account-management-statements.html

> **注意：** 直接修改 `spec.user` （用户名）等同于以新用户名创建一个用户。如需创建多个用户，请确保 metadata.name （CR 实例名）与 spec.user（用户名）一一对应。