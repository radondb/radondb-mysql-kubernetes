# Introduction

The [Xenon](https://hub.docker.com/repository/docker/kryptondb/xenon) image has been pushed into docker hub. The available versions are:

    kryptondb/xenon (tag: 1.1.5-alpha)

Images are updated when new releases are published. 

# Environment Variables

## `MYSQL_ROOT_PASSWORD`

This variable specifies a root password that will be set in the configuration file `xenon.json`.

## `MYSQL_REPL_PASSWORD`

This variable specifies a replication password that will be set in the configuration file `xenon.json`, the default is `Repl_123`.

## `HOST`

This variable is used to specify the endpoint in the kubenetes cluster.

## `LEADER_START_CMD`

This variable is the shell command when leader start.

## `LEADER_STOP_CMD`

This variable is the shell command when leader stop.

# Build Image

```
docker build -t xenon:v1.0 .
```
