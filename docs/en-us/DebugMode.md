English | [简体中文](../zh-cn/DebugMode.md)

Contents
===========
 * [Debug mode](#debug-mode)
 * [Remove the debug mode](#remove-the-debug-mode)

# Enable the debug mode
To avoid the restart-on-fail loop of mysql container in O&M, enable the debug mode. You need to create an empty file `/var/lib/mysql/sleep-forever` as follows.

```bash
kubectl exec -it sample-mysql-0 -c mysql -- sh -c 'touch /var/lib/mysql/sleep-forever'
```
As a result, the MySQL container in the `sample-mysql-0` pod will never restart when the mysqld crashes.

# Remove the debug mode

```bash
kubectl exec -it sample-mysql-0 -c mysql -- sh -c 'rm /var/lib/mysql/sleep-forever'
```

As a result, the MySQL container in the `sample-mysql-0` pod will restart after mysqld exits.