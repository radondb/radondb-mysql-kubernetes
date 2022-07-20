English | [简体中文](../zh-cn/manage_mysql_user.md)

Contents
==========
   * [1. Prerequisite](#1-prerequisite)
   * [2. Create user accounts](#2-create-user-accounts)
      * [2.1 Validate CRD](#21-validate-crd)
      * [2.2 Create users](#22-create-users)
      * [2.3 View users](#23-view-users)
   * [3. Log on as a user](#3-log-on-as-a-user)
   * [4. Delete users](#4-delete-users)
   * [5. Parameters](#5-parameters)

# Manage MySQL users with MysqlUser CRD

##  1. Prerequisite

* The [RadonDB MySQL cluster](deploy_radondb-mysql_operator_on_k8s.md) has been deployed.

##  2. Create user accounts

###  2.1 Validate CRD

Run the following command, and the `mysqlusers.mysql.radondb.com` CRD is displayed.

```plain
kubectl get crd | grep mysqluser
mysqlusers.mysql.radondb.com                          2021-09-21T09:15:08Z
```

###  2.2 Create users

Run the following command to create a normal user named `normal_user` and a superuser named `super_user`. The user password is saved in the `sample-user-password` [Secret](https://kubernetes.io/docs/concepts/configuration/secret/).

```plain
kubectl apply -f https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqluser.yaml
```

> **Note:** In the example, the passwords for the normal user and superuser are both `RadonDB@123`.

### 2.3 View users

```plain
kubectl get mysqluser -o wide                                                                                      
NAME          USERNAME      SUPERUSER   HOSTS   TLSTYPE   CLUSTER   NAMESPACE   AVAILABLE   SECRETNAME             SECRETKEY
normal-user   normal_user   false       ["%"]   NONE      sample    default     True        sample-user-password   normalUser
super-user    super_user    true        ["%"]   NONE      sample    default     True        sample-user-password   superUser
```

## 3. Log on as a user

Run the following command to connect to the primary node of the MySQL cluster as `super_user`.

```plain
kubectl exec -it svc/sample-leader -c mysql -- mysql -usuper_user -pRadonDB@123
```

## 4. Delete users
Run the following command to delete the `MysqlUser` CRD and the users created in the example.
```plain
kubectl delete mysqluser normal-user super-user
```

## 5. Parameters
| Parameters                | Description                                                        |
| ------------------------- | ------------------------------------------------------------------ |
| user                      | User name                                                          |
| hosts                     | Hosts allowed to access; `%` indicates all hosts can be accessed.  |
| withGrantOption           | Whether a user can authorize other users; default value: `false`   |
| tlsOptions.type           | TLS type; valid values: `NONE`/`SSL`/`X509`; default value: `NONE` |
| permissions.database      | Authorized databases; `*` indicates all databases are authorized.  |
| permissions.tables        | Authorized tables; `*` indicates all tables are authorized.        |
| permissions.privileges    | Privileges                                                         |
| userOwner.clusterName     | Name of the cluster that the user is in                            |
| userOwner.nameSpace       | Namespace of the cluster that the user is in                       |
| secretSelector.secretName | Name of the Secret saving the user password                        |
| secretSelector.secretKey  | Key of the Secret saving the user password                         |

> For more details, see [Account Management Statements](https://dev.mysql.com/doc/refman/5.7/en/account-management-statements.html).

> Note: Modifying `spec.user` directly will create a user with the new username. To create multiple users, ensure that `metadata.name` (CRD instance name) is consistent with `spec.user` (username).