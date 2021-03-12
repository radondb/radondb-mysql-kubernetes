# Introduction

The [krypton](https://hub.docker.com/repository/docker/zhyass/krypton) image has been pushed into docker hub. The available versions are:

    zhyass/krypton (tag: beta0.1.0)

Images are updated when new releases are published. 

# Environment Variables

## `MYSQL_REPL_PASSWORD`

This variable specifies a replication password that will be set in the configuration file `krypton.json`, the default is `Repl_123`.

## `HOST_SUFFIX`

This variable is used to specify the endpoint in the kubenetes cluster, default is nil.

# Build Image

```
docker build -t krypton:v1.0 .
```
