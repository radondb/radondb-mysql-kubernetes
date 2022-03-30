Contents
===

* [Install a RadonDB MySQL Cluster on Rancher](#install-a-radondb-mysql-cluster-on-rancher)
  * [Introduction](#introduction)
  * [Prerequisites](#prerequisites)
  * [Procedure](#procedure)
    * [Step 1: Add a Helm Repository](#step-1-add-a-helm-repository)
    * [Step 2: Install Operator](#step-2-install-operator)
    * [Step 3: Install a RadonDB MySQL Cluster](#step-3-install-a-radondb-mysql-cluster)
  * [Verification](#verification)
  * [Access RadonDB MySQL](#access-radondb-mysql)

# Install a RadonDB MySQL Cluster on Rancher

## Introduction

RadonDB MySQL is an open-source, highly available, and cloud-native database cluster solution based on MySQL. It supports a high availability (HA) architecture with one leader node and multiple follower nodes, and supports a full set of management features such as security, automatic backup, monitoring and alerting, and automatic scaling. RadonDB MySQL has been widely used in production by enterprises such as banks, insurance enterprises, and other traditional large enterprises.

RadonDB MySQL can be installed and managed on Rancher so that tasks relevant to RadonDB MySQL clusters are run automatically.

This tutorial demonstrates how to install RadonDB MySQL Operator and a RadonDB MySQL cluster on Rancher.

## Prerequisites

- You need to [install a Rancher cluster](https://rancher.com/docs/rancher/v2.6/en/quick-start-guide/deployment/quickstart-manual-setup/) and obtain the username and password for logging in to Rancher.

## Procedure

### Step 1: Add a Helm Repository

1. Log in to the Rancher console.

2. Select a cluster to open the cluster management page.

3. Select **App&Marketplace** > **Repositories** to go to the application repository management page.

4. Click **Create** to create a repository for RadonDB MySQL.
   
   - **Name**: User-defined repository name.
   
   - **Target**: Select the HTTP(S) mode and set **Index URL** to `https://radondb.github.io/radondb-mysql-kubernetes/`.

5. Click **Create** to return to the repository management page.
   
   When **State** of the repository changes to `Active`, the repository is running properly.

### Step 2: Install Operator

You only need to install RadonDB MySQL Operator once for a Rancher cluster.

1. On the cluster management page, select **App\&Marketplace** > **Charts** to go to the chart list page.

2. Locate **mysql-operator** to install RadonDB MySQL Operator.
   
   You can select a version of the mysql-operator chart.
   
   ![Charts](_images/mysql_operator_chart.png)

3. Click **Install** and configure the application basic information.
   
   You can select **Customize Helm options before install**.
   
   ![Metadata](_images/mysql_operator_metadata.png)

4. (Optional) Click **Next** to customize the YAML settings of the application.
   
   ![Values](_images/mysql_operator_values.png)

5. Click **Next** to configure the deployment options.
   
   ![Helm Options](_images/helm_options.png)

6. Click **Install** to go to the **Installed App** page.
   
   You can view the installation progress and status in the pane below the list. After the application installation process is complete, you can view the installed application in the list.

### Step 3: Install a RadonDB MySQL cluster

#### Use the CLI

1. On the cluster management page, click the kubectl command icon in the upper-right corner.
   
   ![kubectl command pane](_images/kubectl_command.png)

2. In the command pane, enter the command for installing a cluster.
   
   The following command installs a three-node cluster as an example.
   
   ```shell
   # Run kubectl commands inside here
   # e.g. kubectl get all
   $ cat <<EOF | kubectl apply -f-
   apiVersion: mysql.radondb.com/v1alpha1
   kind: MysqlCluster
   metadata:
      name: sample
   spec:
      replicas: 3
   EOF
   ```

3. Press **Enter**. If `created` is displayed in the command output, the installation is successful.
   
   The following is an example of the command output:
   
   ```shell
   mysqlcluster.mysql.radondb.com/sample created
   ```

#### Import a YAML File

1. Download the [RadonDB MySQL Cluster Configuration Sample](/config/samples/mysql_v1alpha1_mysqlcluster.yaml) and modify the parameter values in the YAML file.
   
   For details about the parameters, see [Configuration Parameters](../zh-cn/config_para.md).

2. On the cluster management page of Rancher, click the YAML import icon in the upper-right corner. In the displayed dialog box, import the modified YAML file.
   
   ![Import YAML](_images/import_yaml.png)

## Verification

1. On the cluster management page, select **Service Discovery** > **Service** to go to the service list page.

2. Locate the installed cluster and check the service status.
   
   If the status of a service is `Active`, the service is running properly.

3. Click the service name to open the service details page and check the pod status.
   
   If the status of a pod is `Active`, the pod is running properly.

4. On the right of an active pod, click **Execute Shell** to open the command pane of the pod.
   
   Run the following command and enter the password to verify the database connection status.
   
   ```shell
   $ mysql -u root -p
   ```
   
   The following figure shows the command output of a database with normal connections:
   
   ![Pod Running](_images/pod_running.png)

## Access RadonDB MySQL

> **Note**
> 
> You need to prepare a client used to connect to MySQL.

- If the client is installed in a different Rancher cluster from the database, you need to [set the load balancer and ingress controller on Rancher](https://rancher.com/docs/rancher/v2.6/en/k8s-in-rancher/load-balancers-and-ingress/).
  
  For more information about how to access a database from a different cluster, see [Access Applications in a Cluster](https://kubernetes.io/docs/tasks/access-application-cluster/).

- If the client is installed in the same Rancher cluster as the database, you can access RadonDB MySQL by using a service name or cluster IP address.
  
  > **Note**
  > 
  > RadonDB MySQL provides a leader service and a follower service, which are used to access the leader node and follower nodes respectively. The leader service always points to the leader node that supports read and write, and the follower service always points to the follower nodes that are read-only.

The following demonstrates how to access RadonDB MySQL by using a client in the same Rancher cluster as the database.

### Use a Cluster IP Address

The HA read-and-write IP address of RadonDB MySQL points to the cluster IP address of the leader service, and the HA read-only IP address points to the IP address of the follower service.

```shell
$ mysql -h <clusterIP> -P <mysql_Port> -u <user_name> -p
```

For example, run the following command to access the leader service, where the username is `radondb_usr` and the cluster IP address of the leader service is `10.10.128.136`:

```shell
$ mysql -h 10.10.128.136 -P 3306 -u radondb_usr -p
```

### Use a Service Name

Pods in the Rancher cluster can access RadonDB MySQL by using a service name.

> **Note**
> 
> Service names cannot be used to access RadonDB MySQL from the host machines of the Rancher cluster.

* Access the leader service (RadonDB MySQL leader node)
  
  ```shell
  mysql -h <leader_service_name>.<namespace> -u <user_name> -p
  ```
  
  For example, run the following command to access the leader service, where the username is `radondb_usr`, the release name is `sample`, and the namespace of RadonDB MySQL is `default`:
  
  ```shell
  mysql -h sample-leader.default -u radondb_usr -p
  ```

* Access the follower service (RadonDB MySQL follower nodes)
  
  ```shell
  mysql -h <follower_service_name>.<namespace> -u <user_name> -p
  ```
  
  For example, run the following command to access the leader service, where the username is `radondb_usr`, the release name is `sample`, and the namespace of RadonDB MySQL is `default`:
  
  ```shell
  mysql -h sample-follower.default -u radondb_usr -p  
  ```