
English | [简体中文](../zh-cn/deploy_radondb-mysql_operator_on_k8s.md)

Contents
===

* [Install a RadonDB MySQL cluster on Kubernetes](#install-a-radondb-mysql-cluster-on-kubernetes)
  * [Overview](#Overview)
  * [Prerequisites](#prerequisites)
  * [Procedure](#procedure)
    * [Step 1: Add a Helm repository](#step-1-add-a-helm-repository)
    * [Step 2: Install Operator](#step-2-install-operator)
    * [Step 3: Deploy a RadonDB MySQL cluster](#step-3-Deploy-a-radondb-mysql-cluster)
  * [Verification](#verification)
    * [Verify RadonDB MySQL Operator](#verify-radondb-mysql-operator)
    * [Verify the RadonDB MySQL cluster](#verify-the-radondb-mysql-cluster)
  * [Access RadonDB MySQL](#access-radondb-mysql)
  * [Uninstallation](#uninstallation)
    * [Uninstall Operator](#uninstall-operator)
    * [Uninstall RadonDB MySQL](#uninstall-radondb-mysql)
    * [Uninstall the custom resources](#uninstall-the-custom-resources)

# Install a RadonDB MySQL cluster on Kubernetes

## Overview

RadonDB MySQL is an open-source, high-availability, and cloud-native database cluster solution based on MySQL. It supports the high-availability architecture of one leader node and multiple replicas, with a set of management functions for security, automatic backup, monitoring and alarming, automatic scaling, and so on. RadonDB MySQL has been widely used in production by **banks, insurance enterprises, and other traditional large enterprises**.

RadonDB MySQL can be deployed and managed on Kubernetes to automate tasks relevant to RadonDB MySQL clusters.

This tutorial demonstrates how to deploy a RadonDB MySQL cluster (Operator) on Kubernetes.

## Prerequisites

* You need to prepare a Kubernetes cluster.

## Procedure

### Step 1: Add a Helm repository.

```shell
helm repo add radondb https://radondb.github.io/radondb-mysql-kubernetes/
```

Check that a chart named `radondb/mysql-operator` exists in the repository.

```shell
$ helm search repo
NAME                            CHART VERSION   APP VERSION                     DESCRIPTION                 
radondb/mysql-operator          0.1.0           v2.1.0                          Open Source，High Availability Cluster，based on MySQL                     
```

### Step 2: Install Operator

The following sets the release name to `demo` and creates a [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) named `demo-mysql-operator`.

```
helm install demo radondb/mysql-operator
```

> **Note**
> 
> This step also creates the [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) required by the cluster.

### Step 3: Deploy a RadonDB MySQL cluster

Create an instance of the `mysqlclusters.mysql.radondb.com` CRD and thereby create a RadonDB MySQL cluster with default parameters as follows. To set cluster parameters, see [Configuration Parameters](../zh-cn/config_para.md).

```kubectl
kubectl apply -f https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqlcluster.yaml
```

## Verification

### Verify RadonDB MySQL Operator

Check the demo Deployment and its monitoring service. The deployment is successful if the following information is displayed.

```kubectl
$ kubectl get deployment,svc
NAME                  READY   UP-TO-DATE   AVAILABLE   AGE
demo-mysql-operator   1/1     1            1           7h50m


NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/mysql-operator-metrics   ClusterIP   10.96.142.22    <none>        8443/TCP   8h
```

### Verify the RadonDB MySQL cluster

Run the following command to check the CRDs:

```kubectl
$ kubectl get crd | grep mysql.radondb.com
backups.mysql.radondb.com                             2021-11-02T07:00:01Z
mysqlclusters.mysql.radondb.com                       2021-11-02T07:00:01Z
mysqlusers.mysql.radondb.com                          2021-11-02T07:00:01Z
```

Run the following command to check the cluster. If a statefulset of three replicas (RadonDB MySQL nodes) and services used to access the nodes are displayed, the installation is successful.

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

- You can use service_name or clusterIP to access RadonDB MySQL in the Kubernetes cluster.
  
  > **Note**
  > 
  > RadonDB MySQL provides a leader service and a follower service, which are used to access the leader node and follower nodes respectively. The leader service always points to the leader node that supports read and write, and the follower service always points to the follower nodes that are read-only.

The following demonstrates how to access RadonDB MySQL by using a client in the same Kubernetes cluster as the database.

### Use ClusterIP

The HA read/write IP address of RadonDB MySQL points to the `clusterIP` of the leader service, and the HA read-only IP address points to the `clusterIP` of the follower services.

```shell
mysql -h <clusterIP> -P <mysql_Port> -u <user_name> -p
```

For example, run the following command to access a leader service. The username is `radondb_usr`, and the clusterIP of the leader service is `10.10.128.136`.

```shell
mysql -h 10.10.128.136 -P 3306 -u radondb_usr -p
```

### Use service_name

Pods in the Kubernetes cluster can access RadonDB MySQL by using service_name.

> **Note**
> 
> `service_name` cannot be used to access database pods from the host machines in the Kubernetes cluster.

* Access the leader service (RadonDB MySQL leader node)
  
  ```shell
  mysql -h <leader_service_name>.<namespace> -u <user_name> -p
  ```
  
  For example, run the following command to access the leader service. The username is `radondb_usr`, the release name is `sample`, and the namespace of RadonDB MySQL is `default`.
  
  ```shell
  mysql -h sample-leader.default -u radondb_usr -p
  ```

* Access the follower service (RadonDB MySQL replicas)
  
  ```shell
  mysql -h <follower_service_name>.<namespace> -u <user_name> -p
  ```

  For example, run the following command to access the follower service. The username is `radondb_usr`, the release name is `sample`, and the namespace of RadonDB MySQL is `default`.
  
  ```shell
  mysql -h sample-follower.default -u radondb_usr -p  
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

### Uninstall the custom resources

```kubectl
kubectl delete customresourcedefinitions.apiextensions.k8s.io mysqlclusters.mysql.radondb.com
kubectl delete customresourcedefinitions.apiextensions.k8s.io mysqlusers.mysql.radondb.com
kubectl delete customresourcedefinitions.apiextensions.k8s.io backups.mysql.radondb.com
```