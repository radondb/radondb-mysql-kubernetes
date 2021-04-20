
# 简介  

[MySQL](https://hub.docker.com/repository/docker/xenondb/percona) 镜像已经发布到 docker hub 中， 当前可用版本为：

    xenondb/percona (tag: 5.7.33)

发布新版本时会在此更新。

# 环境变量

## `MYSQL_ROOT_PASSWORD`

指定 root 超级账户密码，可以为空。

**注意**: 在命令行上设置 MySQL root 用户密码是不安全的。

## `MYSQL_REPL_PASSWORD`

指定复制账户密码，默认为 `Repl_123`。

## `INIT_TOKUDB`

设置为 `1` 表示启用 TOKUDB 引擎。

## `MYSQL_INITDB_SKIP_TZINFO`

设置此变量后，将跳过 MySQL 的时区信息设置。

## `MYSQL_DATABASE`

此变量可选，允许用户指定镜像启动时创建的数据库名称。如果提供了用户/密码（请参阅下文），则将授予该用户对该数据库的超级用户访问权限（GRANT ALL）。

## `MYSQL_USER`, `MYSQL_PASSWORD`

变量可选，结合使用以创建一个新用户并设置该用户的密码。将为该用户授予 `MYSQL_DATABASE` 变量指定的数据库的超级用户权限（请参见上文）。这两个变量都是创建用户所必需的。

# 构建镜像

    docker build -t mysql:v1.0 .
