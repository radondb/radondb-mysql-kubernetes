English | [简体中文](../zh-cn/deploy_radondb-mysql_operator_on_k8s_offline.md)

Contents
=============

- [Offline Deployment of RadonDB MySQL Cluster on Kubernetes](#Offline-Deployment-of-RadonDB-MySQL-Cluster-on-Kubernetes)
  - [Overview](#Overview)
  - [Prerequisite](#Prerequisite)
  - [Precedure](#Precedure)
    - [Step 1: Prepare resources](#Step-1-Prepare-resources)
    - [Step 2: Deploy Operator](#Step-2-Deploy-Operator)
    - [Step 3: Deploy RadonDB MySQL cluster](#Step-3-Deploy-RadonDB-MySQL-cluster)
  - [Verification](#Verification)
    - [Verify RadonDB MySQL Operator](#Verify-radondb-mysql-operator)
    - [Verify RadonDB MySQL 集群](#Verify-radondb-mysql-cluster)
  - [Access RadonDB MySQL](#Access-radondb-mysql)
    - [Use `service_name`](#Use-service_name)
    - [Use `clusterIP`](#Use-clusterip)
  - [Uninstallation](#Uninstallation)
    - [Uninstall Operator](#Uninstall-operator)
    - [Uninstall RadonDB MySQL](#Uninstall-radondb-mysql)
    - [Uninstall custom resources](#Uninstall-custom-resources)

# Offline Deployment of RadonDB MySQL Cluster on Kubernetes

## Overview

RadonDB MySQL is an open-source, high-availability, and cloud-native database cluster solution based on MySQL. It supports the high-availability architecture of one leader node and multiple replicas, with a set of management functions for security, automatic backup, monitoring and alarming, automatic scaling, and so on. RadonDB MySQL has been widely used in production by **banks, insurance enterprises, and other traditional large enterprises**.

RadonDB MySQL supports installation, deployment and management on Kubernetes, automating tasks involved in running RadonDB MySQL clusters.

This tutorial demonstrates how to deploy the RadonDB MySQL Operator and cluster offline on Kubernetes.

## Prerequisite

- You need to prepare a Kubernetes cluster.

## Precedure

### Step 1: Prepare resources

#### Download offline resources
Download the `radondb/mysql-operator, radondb/mysql57-sidecar, radondb/mysql80-sidercar,percona/percona-server:5.7.34, percona/percona-server:8.0.25` image from Docker Hub and load them to available worker nodes.


#### Import image (perform it on each worker node running the database)

```
docker load -i XXXX
```
Replace `XXXX` with the name of the downloaded image file.

### Step 2: Deploy Operator

The following sets the release name to `demo`, and creates a [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) named `demo-mysql-operator`.

```
helm install demo radondb-mysql-resources/operator-chart .
```

> **Note**
> 
> By default, this step will also create the CRD required by the cluster. You can find the corresponding [release](https://github.com/radondb/radondb-mysql-kubernetes/releases).

### Step 3: Deploy RadonDB MySQL cluster

Create an instance for the `mysqlclusters.mysql.radondb.com` CRD, and thereby create a RadonDB MySQL cluster, with the default parameters as follows. To customize cluster parameters, see [Configure parameters](./config_para.md).

```kubectl
kubectl apply -f radondb-mysql-resources/cluster-sample/mysql_v1alpha1_mysqlcluster.yaml
```

## Verification

### Verify RadonDB MySQL Operator

Check the `demo` Deployment and its monitoring service. The deployment is successful if the following information is displayed.

```kubectl
$ kubectl get deployment,svc
NAME                  READY   UP-TO-DATE   AVAILABLE   AGE
demo-mysql-operator   1/1     1            1           7h50m


NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/mysql-operator-metrics   ClusterIP   10.96.142.22    <none>        8443/TCP   8h
```

### Verify RadonDB MySQL cluster

Check the CRDs as follows.

```kubectl
$ kubectl get crd | grep mysql.radondb.com
backups.mysql.radondb.com                             2021-11-02T07:00:01Z
mysqlclusters.mysql.radondb.com                       2021-11-02T07:00:01Z
mysqlusers.mysql.radondb.com                          2021-11-02T07:00:01Z
```

For the default deployment, run the following command to check the cluster, and a statefulset of three replicas (RadonDB MySQL nodes) and services used to access the nodes are displayed.

```kubectl
$ kubectl get statefulset,svc
NAME           READY   AGE
sample-mysql   3/3     7h33m

NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/sample-follower          ClusterIP   10.96.131.84    <none>        3306/TCP   7h37m
service/sample-leader            ClusterIP   10.96.111.214   <none>        3306/TCP   7h37m
service/sample-mysql             ClusterIP   None            <none>        3306/TCP   7h37m
```

## Access RadonDB MySQL

> **Note**
> 
> You need to prepare a client used to connect to MySQL.

- If the client is installed in a different Kubernetes cluster, see [Access Applications in a Cluster](https://kubernetes.io/docs/tasks/access-application-cluster/) to configure port forwarding and load balancing.
  
- You can use `service_name` or `clusterIP` to access RadonDB MySQL in the Kubernetes cluster.
  
  > **Note**
  > 
  > RadonDB MySQL provides the leader and follower services to access the leader node and replicas respectively. The leader service always points to the leader node (read/write) and the follower service points to the replicas (read only).
  

If the client and database are in the same Kubernetes cluster, access RadonDB MySQL as follows.

### Use `service_name`

- Access the leader service (RadonDB MySQL leader node).
  
  ```shell
  mysql -h <leader_service_name>.<namespace> -u <user_name> -p
  ```
  
  Access the leader service as follows. The username is `radondb_usr`, the release name is `sample`, and the namespace of RadonDB MySQL is `default`.
  
  ```shell
  mysql -h sample-leader.default -u radondb_usr -p
  ```
  
- Access the follower service (RadonDB MySQL replicas)
  
  ```shell
  mysql -h <follower_service_name>.<namespace> -u <user_name> -p
  ```
  
  Access the replicas as follows. The username is `radondb_usr`, the release name is `sample`, and the namespace of RadonDB MySQL is `default`.

  ```shell
  mysql -h sample-follower.default -u radondb_usr -p  
  ```
  

### Use `clusterIP`

The HA read/write IP address of RadonDB MySQL points to the `clusterIP` of the leader service, and the HA read-only IP address points to the `clusterIP` of the follower services.

```shell
mysql -h <clusterIP> -P <mysql_Port> -u <user_name> -p
```

Access a leader service as follows. The username is radondb_usr, and the clusterIP of the leader service is 10.10.128.136.

```shell
mysql -h 10.10.128.136 -P 3306 -u radondb_usr -p
```

## Uninstallation

### Uninstall Operator

Uninstall RadonDB MySQL Operator with the release name `demo` in the current namespace.

```shell
helm delete demo
```

### Uninstall RadonDB MySQL

Uninstall the RadonDB MySQL cluster with the release name `sample`.

```kubectl
kubectl delete mysqlclusters.mysql.radondb.com sample
```

### Uninstall custom resources

```kubectl
kubectl delete customresourcedefinitions.apiextensions.k8s.io mysqlclusters.mysql.radondb.com
kubectl delete customresourcedefinitions.apiextensions.k8s.io mysqlusers.mysql.radondb.com
kubectl delete customresourcedefinitions.apiextensions.k8s.io backups.mysql.radondb.com
```