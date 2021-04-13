
# 简介

[Xenon](https://hub.docker.com/repository/docker/kryptondb/xenon) 镜像已经发布在 docker hub，当前可用版本：

    kryptondb/xenon (tag: 1.1.5)

发布新版本时会在此更新。

# 环境变量

## `MYSQL_ROOT_PASSWORD`

指定将在配置文件 `xenon.json` 中设置的 root 账户密码。

## `MYSQL_REPL_PASSWORD`

指定将在配置文件 `xenon.json` 中设置的复制账户密码，默认值为 `Repl_123`。

## `HOST`

指定 kubenetes 集群中的 endpoint。

## `Master_SysVars`

在 master 节点上执行的 MySQL 配置参数.

## `Slave_SysVars`

在 slave 节点上执行的 MySQL 配置参数.

## `LEADER_START_CMD`

在 leader 节点启动时执行的 shell 命令。

## `LEADER_STOP_CMD`

在 leader 节点停止时执行的 shell 命令。

# 构建镜像

```bash
docker build -t xenon:v1.0 .
```
