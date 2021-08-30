# CRD 管理用户使用说明

此文档供测试期间使用，非最终版本。

目前仅支持使用用户名，密码，hosts创建或删除用户。

## CRD 介绍

```
  ## 要创建的用户名
  user: sample_user
  ## 允许访问的主机，可以填多个
  hosts: 
    - "%"
  ## 绑定一个 mysql 集群，需要指定集群名字和 ns。
  clusterBinder:
    clusterName: sample-mysql
    nameSpace: default
  ## 绑定一个 secret，secret里保存密码。
  secretBinder:
    secretName: sample-user-password
    secretKey: password
  ## 是否开启 ssl
  enableSsl: true
  ## 授权信息列表，一个 permission 对象只能指定一个数据库。
  permissions:
    - database: "*"
      tables: 
        - "*"
      privileges: 
        - SELECT
```

## 步骤

### 1.安装 CRD 

使用以下指令安装 operator 时将自动安装 User CRD。

```
helm install demo charts/mysql-operator/
```

### 2.安装 RadonDB MySQL 集群

Cluster CRD 中 `metadata.name` 的值对应 User CRD 中的 `clusterBinder.clusterName`。

示例：现有 RadonDB MySQL 集群 `metadata.name` 值为 `sample`， 需要将用户 CRD 中 `clusterBinder.clusterName` 的值设置为 `sample`，确保 User CR 能绑定到期望的数据库。

```
kubectl apply -f config/samples/mysql_v1alpha1_cluster.yaml 
```

### 3.创建 Secret

```
kubectl apply -f config/samples/mysql_v1alpha1_user_secret.yaml 
```

### 4.创建 User

```
kubectl apply -f config/samples/mysql_v1alpha1_user.yaml 
```

### 5.删除 User

```
kubectl delete -f config/samples/mysql_v1alpha1_user.yaml 
```
