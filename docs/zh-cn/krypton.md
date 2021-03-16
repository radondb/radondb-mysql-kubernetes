
# 简介

[krypton](https://hub.docker.com/repository/docker/zhyass/krypton) 镜像已经发布在 docker hub，当前可用版本：

```bash
zhyass/krypton (tag: beta0.1.0)
```

发布新版本时会在此更新。

# 环境变量

## `MYSQL_REPL_PASSWORD`

指定将在配置文件 `krypton.json` 中设置的复制账户密码，默认值为 `Repl_123`。

## `HOST_SUFFIX`

指定 kubenetes 集群中的 endpoint ，默认为 nil 。

# 构建镜像

```bash
docker build -t krypton:v1.0 .
```
