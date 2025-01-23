Contents
=================

   * [简介](#简介)
   * [限制](#限制)
   * [用法](#用法)
     * [Check](#check)
     * [Upgrade](#upgrade)

# 简介
从 MySQL5.7 升级到 MySQL8.0 有两种方式, 一种是直接从 MySQL5.7 的数据文件中升级, 这种方式被称为 In-place Upgrade, 另一种是通过`msqldump` 进行逻辑备份与恢复, 这种方式被称为 `logic upgrade` , `updrade80.sh` 是采用第一种方式进行的.

# 限制
采用`in-place upgrade` 速度较快,但是必须要谨慎, 严格参考 [mysql手册](https://dev.mysql.com/doc/refman/8.0/en/upgrade-binary-package.html#upgrade-procedure-inplace) 要求进行检查, 并且事先要做好备份.

# 用法
 建议实现执行 Check, 如果check 没有任何信息, 然后执行 Update 命令,
## Check 
用法：
```
upgrade80.sh check [集群名称]
```
示例:
```
./upgrade80.sh check sample 
```

## Upgrade
用法
```
./upgrade80.sh update [集群名称] [mysql80的sidecar镜像]
```
示例:
```
 ./upgrade80.sh update sample "radondb/mysql80-sidecar:v2.2.0"
 ```