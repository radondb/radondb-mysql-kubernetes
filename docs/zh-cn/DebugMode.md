Contents
=============
 * [Debug模式](#debug模式)
 * [移除debug模式](#移除debug模式)


# Debug 模式

在运维阶段, 如果你想避免restart-on-fail循环的 mysql 容器，你应该使用 Debug 模式。只要创建一个空文件 `/var/lib/mysql/sleep-forever` 即可.

示例:

```bash
kubectl exec -it sample-mysql-0 -c mysql -- sh -c 'touch /var/lib/mysql/sleep-forever'
```
该命令会让 Pod `sample-mysql-0` 的 mysql 容器在 mysqld 已经crash的情况下, 永远不会重启.

# 移除 Debug 模式

```bash
kubectl exec -it sample-mysql-0 -c mysql -- sh -c 'rm /var/lib/mysql/sleep-forever'
```
该命令会让 Pod `sample-mysql-0` 的 mysql 恢复正常情况, 即 mysqld 退出后, mysql 容器会重启.