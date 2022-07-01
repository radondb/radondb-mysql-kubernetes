English | [简体中文](../zh-cn/mysql.md)

Contents
=================

   * [Synopsis](#Synopsis)
   * [Environment variables](#Environment-variables)
      * [MYSQL_ROOT_PASSWORD](#mysql_root_password)
      * [MYSQL_REPL_PASSWORD](#mysql_repl_password)
      * [INIT_TOKUDB](#init_tokudb)
      * [MYSQL_INITDB_SKIP_TZINFO](#mysql_initdb_skip_tzinfo)
      * [MYSQL_DATABASE](#mysql_database)
      * [MYSQL_USER, <code>MYSQL_PASSWORD</code>](#mysql_user-mysql_password)
   * [Build image](#Build-image)

# Synopsis

The [MySQL](https://hub.docker.com/repository/docker/radondb/percona) image has been released on Docker Hub. Currently available version:

    radondb/percona (tag: 5.7.34)

It will be updated when a new version is released.

# Environment variables

## `MYSQL_ROOT_PASSWORD`

It specifies the root password, which can be empty.

>**Note**: It is not safe to set the root password by the command line.

## `MYSQL_REPL_PASSWORD`

It specifies the replica password, which is `Repl_123` by default.

## `INIT_TOKUDB`

Setting to `1` enables the TOKUDB engine.

## `MYSQL_INITDB_SKIP_TZINFO`

Setting this variable skips setting MySQL time zone.

## `MYSQL_DATABASE`

It is optional, and the user is allowed to specify a name for the database created when the image is started. If a user/password is provided (see below), the user is granted the superuser access to the database (GRANT ALL).

## `MYSQL_USER`, `MYSQL_PASSWORD`

They are optional and used to create new users and set passwords. The user will be granted the superuser access to the database specified by the MYSQL_DATABASE variable (see above). Both variables are required to create a user.

# Build image

    docker build -t mysql:v1.0 .