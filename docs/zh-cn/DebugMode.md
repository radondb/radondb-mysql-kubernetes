[English](../en-us/DebugMode.md) | 简体中文

目录
=============
 * [Debug 模式](#debug-模式)
 * [移除 Debug 模式](#移除-debug-模式)


# 使用 Debug 模式

在运维阶段, 如果想避免 MySQL 容器出现 restart-on-failure 循环，你应该使用 Debug 模式。只需创建一个空文件 `/var/lib/mysql/sleep-forever` 即可。

示例:

```bash
kubectl exec -it sample-mysql-0 -c mysql -- sh -c 'touch /var/lib/mysql/sleep-forever'
```
该命令会让 Pod `sample-mysql-0` 的 MySQL 容器在 mysqld 已经崩溃的情况下，永远不会重启。

# 移除 Debug 模式

```bash
kubectl exec -it sample-mysql-0 -c mysql -- sh -c 'rm /var/lib/mysql/sleep-forever'
```
该命令会让 Pod `sample-mysql-0` 的 MySQL 容器恢复正常，即 mysqld 退出后，MySQL 容器会重启。