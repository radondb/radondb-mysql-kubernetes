[English](../en-us/mgt_mysqlusser.md) | 简体中文

目录
==========
   * [1. 前提条件](#1.-前提条件)
   * [2. 创建用户帐号](#2.-创建用户帐号)
      * [2.1.1 校验 CRD](#2.1.1-校验-CRD)
      * [2.2.2 创建 Secret](#2.2.2-创建-Secret)
      * [2.3 创建用户](#2.3-创建用户)
   * [3. 修改用户帐号](#3.-修改用户帐号)
      * [3.1 授权 IP](#3.1-授权-IP)
      * [3.2 用户权限](#3.2-用户权限)
   * [4. 删除用户帐号](#4.-删除用户帐号)
   * [5. 示例配置](#5.-示例配置)
      * [5.1 Secret](#5.1-Secret)
      * [5.2 MysqlUser](#5.2-MysqlUser)

# 使用 MysqlUser CRD 管理 MySQL 用户

##  1. 前提条件

* 已部署 [RadonDB MySQL 集群](kubernetes/deploy_radondb-mysql_operator_on_k8s.md)。

##  2. 创建用户帐号

###  2.1.1 校验 CRD

使用如下指令，将查看到名称为 `mysqlusers.mysql.radondb.com` 的 CRD。

```plain
kubectl get crd | grep mysqluser
mysqlusers.mysql.radondb.com                          2021-09-21T09:15:08Z
```

###  2.2.2 创建 Secret

RadonDB MySQL 使用 K8S 中的 [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) 对象保存用户的密码。

使用如下指令，将使用[示例配置](#Secret)创建一个名为 `sample-user-password` 的 Secret。

```plain
kubectl apply -f https://raw.githubusercontent.com/radondb/radondb-mysql-kubernetes/main/config/samples/mysqluser_secret.yaml
```

###  2.3 创建用户

使用如下指令，将使用[示例配置](#MysqlUser)创建一个名为 `sample_user` 的用户。

```plain
kubectl apply -f https://raw.githubusercontent.com/radondb/radondb-mysql-kubernetes/main/config/samples/mysql_v1alpha1_mysqluser.yaml 
```

> **注意：** 直接修改 `spec.user` （用户名）等同于以新用户名创建一个用户。如需创建多个用户，请确保 metadata.name （CR 实例名）与 spec.user（用户名）一一对应。

##  3. 修改用户帐号

用户帐号信息由 spec 字段中参数定义，目前支持：

* 修改 `hosts` 参数。
* 新增 `permissions` 参数。

###  3.1 授权 IP

允许使用用户帐号的 IP，通过 `hosts` 字段参数定义。

* % 表示所有 IP 均可访问。
* 可修改一个或多个 IP。

```plain
  hosts: 
    - "%"
```

###  3.2 用户权限

用户帐号数据库访问权限，通过 MysqlUser 中 permissions 字段参数定义。可通过新增 permissions 字段参数值，实现用户帐号权限的新增。

```plain
permissions:
    - database: "*"
      tables:
        - "*"
      privileges:
        - SELECT
```

* `database`  参数表示该用户帐号允许访问的数据库。* 代表允许访问集群所有数据库。
* `tables`  参数表示该用户帐号允许访问的数据库表。 * 代表允许访问数据库中所有表。
* `privileges`  参数表示该用户帐号被授权的数据库权限。更多权限说明，请参见 [Privileges Supported by MySQL](https://dev.mysql.com/doc/refman/5.7/en/grant.html)。

##  4. 删除用户帐号

使用如下指令将删除使用[示例配置](#MysqlUser)创建的 MysqlUser CR。

```plain
kubectl delete mysqluser sample-user-cr
```

>说明：删除 MysqlUser CR 会自动删除 CR 对应的 MySQL 用户。

##  5. 示例配置

###  5.1 Secret

```plain
apiVersion: v1
kind: Secret
metadata:
  name: sample-user-password   # 密钥名称。应用于 MysqlUser 中的 secretSelector.secretName。  
data:
  pwdForSample: UmFkb25EQkAxMjMKIA==  #密钥键，应用于 MysqlUser 中的 secretSelector.secretKey。示例密码为 base64 加密的 RadonDB@123
  # pwdForSample2:
  # pwdForSample3:
```

###  5.2 MysqlUser

```plain
apiVersion: mysql.radondb.com/v1alpha1
kind: MysqlUser
metadata:
 
  name: sample-user-cr  # 用户 CR 名称，建议使用一个用户 CR 管理一个用户。
spec:
  user: sample_user  # 需要创建/更新的用户的名称。
  hosts:            # 支持访问的主机，可以填多个，% 代表所有主机。 
       - "%"
  permissions:
    - database: "*"  # 数据库名称，* 代表所有数据库。 
      tables:        # 表名称，* 代表所有表。
         - "*"
      privileges:     # 权限，参考 https://dev.mysql.com/doc/refman/5.7/en/grant.html。
         - SELECT
  
  userOwner:  # 指定被操作用户所在的集群。不支持修改。  
    clusterName: sample
    nameSpace: default # radondb mysql 集群所在的命名空间。
  
  secretSelector:  # 指定用户的密钥和保存当前用户密码的键。
    secretName: sample-user-password  # 密钥名称。   
    secretKey: pwdForSample  # 密钥键，一个密钥可以保存多个用户的密码，以键区分。
```