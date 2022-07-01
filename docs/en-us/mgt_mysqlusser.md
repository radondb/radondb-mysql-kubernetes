English | [简体中文](../zh-cn/mgt_mysqluser.md)

Contents
==========
   * [1. Prerequisites](#1.-Prerequisites)
   * [2. Create user account](#2.-Create user account)
      * [2.1.1 Validate CRD](#2.1.1-Validate-CRD)
      * [2.2.2 Create Secret](#2.2.2-Create-Secret)
      * [2.3 Create user](#2.3-Create-user)
   * [3. Modify user account](#3.-Modify-user-account)
      * [3.1 Authorize IP address](#3.1-Authorize-IP-address)
      * [3.2 User privilege](#3.2-User-privilege)
   * [4. Delete user account](#4.-Delete-user-account)
   * [5. Sample configuration](#5.-Sample-configuration)
      * [5.1 Secret](#5.1-Secret)
      * [5.2 MysqlUser](#5.2-MysqlUser)

# Use MysqlUser CRD to manage MySQL users

##  1. Prerequisites

* The [RadonDB MySQL cluster](deploy_radondb-mysql_operator_on_k8s.md) has been deployed.

##  2. Create user account

###  2.1.1 Validate CRD

Run the following command, and the `mysqlusers.mysql.radondb.com` CRD is displayed.

```plain
kubectl get crd | grep mysqluser
mysqlusers.mysql.radondb.com                          2021-09-21T09:15:08Z
```

###  2.2.2 Create Secret

RadonDB MySQL uses the [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) object in Kubernetes to save the user password.

Run the following command to create a Secret named `sample-user-password` using the [sample configuration](#5.1-Secret).

```plain
kubectl apply -f https://raw.githubusercontent.com/radondb/radondb-mysql-kubernetes/main/config/samples/mysqluser_secret.yaml
```

###  2.3 Create user

Run the following command to create a user named `sample_user` using the [sample configuration](#52-mysqluser).

```plain
kubectl apply -f https://raw.githubusercontent.com/radondb/radondb-mysql-kubernetes/main/config/samples/mysql_v1alpha1_mysqluser.yaml 
```

> **Note:** Modifying `spec.user` (username) directly creates a new user with the username. To create multiple users, make sure that metadata.name (CR instance name) corresponds to spec.user.

##  3. Modify user account

The user account is defined by the parameters in the spec field. Currently, the following operations are supported:

* Modify the `hosts` parameter.
*	Add the `permissions` parameter.


###  3.1 Authorize IP address

You are allowed to authorize the IP address of the user account by defining the `hosts` parameter:

* `%` indicates all IP addresses are authorized.
*	You can modify one or more IP addresses.

```plain
  hosts: 
    - "%"
```

###  3.2 User privilege

You can define the database access permission for the user account with the `permissions` field in `mysqlUser`, and add user rights by adding parameters in the `permissions` field.

```plain
permissions:
    - database: "*"
      tables:
        - "*"
      privileges:
        - SELECT
```

* The `database` parameter indicates the database that the user account is allowed to access. `*` indicates the user account is allowed to access all databases in the cluster.
* The `tables` parameter indicates the database tables that the user account is allowed to access. `*` indicates the user account is allowed to access all tables in the database.
* The `privileges` parameter indicates the database permissions granted for the user account. For more information, see [privileges supported by MySQL](https://dev.mysql.com/doc/refman/5.7/en/grant.html).

##  4. Delete user account

Delete the MysqlUser CR created with the [sample configuration](#52-mysqluser) as follows.

```plain
kubectl delete mysqluser sample-user-cr
```

>**Note:** Deleting the MysqlUser CR automatically deletes the corresponding MySQL user.

## 5. Sample configuration

### 5.1 Secret

```plain
apiVersion: v1
kind: Secret
metadata:
  name: sample-user-password   # Secret name, applied to the secretSelector  
data:
  pwdForSample: UmFkb25EQkAxMjMKIA==  
  # secret key, applied to secretSelector.secretKey. The example password is base64-encoded RadonDB@123.
  # pwdForSample2:
  # pwdForSample3:
```

###  5.2 MysqlUser

```plain
apiVersion: mysql.radondb.com/v1alpha1
kind: MysqlUser
metadata:
 
  name: sample-user-cr  # User CR name. It is recommended that you manage one user with one user CR.
spec:
  user: sample_user  # The name of the user to be created/updated
  hosts:            # Hosts can be accessed. You can specify multiple hosts. % represents all hosts.
       - "%"
  permissions:
    - database: "*"  # Database name. * indicates all databases. 
      tables:        # Table name. * indicates all tables
         - "*"
      privileges:     # Privileges. See https://dev.mysql.com/doc/refman/5.7/en/grant.html for more details.
         - SELECT
  
  userOwner:  # Specify the cluster where the user is located. It cannot be modified.
    clusterName: sample
    nameSpace: default # The namespace of the RadonDB MySQL cluster
  
  secretSelector:  # The secret key specifying the user and storing the user password
    secretName: sample-user-password  # Password name
    secretKey: pwdForSample  # The passwords of multiple users can be stored in a secret and distinguished by keys.
```