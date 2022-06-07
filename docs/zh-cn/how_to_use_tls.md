[TOC]

# 为MySQL客户端开启加密连接

# `TLS`(传输层加密)简介

RadonDB MySQL Operator 默认采用非加密连接，如果具备网络嗅探及监视的第三方工具可能截获服务端与客户端之间的数据，容易造成信息泄露，因此建议开启加密连接来确保数据安全。

RadonDB MySQL Operator 服务端支持`TLS`,协议为MySQL支持的加密协议，如`5.7`版本支持`TLS 1.0、TLS 1.1、TLS 1.2`、`8.0`版本支持`TLS 1.0、TLS 1.1、TLS 1.2、TLS 1.3`。

使用加密连接需要满足两个条件：

* MySQL Operator 服务端开启加密连接支持
* 客户端使用加密连接

# 配置`MySQL Operator`使用加密连接

## 准备证书

* `ca.crt` - 服务端`CA`证书
* `tls.key` - 服务端证书私钥
* `tls.crt` - 服务端证书

可以用`OpenSSL`生成，也可以用`MySQL`自带的`mysql_ssl_rsa_setup`快捷生成：

`mysql_ssl_rsa_setup --datadir=/tmp/certs`

运行该命令后会生成如下文件：

```shell
certs
├── ca-key.pem
├── ca.pem
├── client-cert.pem
├── client-key.pem
├── private_key.pem
├── public_key.pem
├── server-cert.pem
└── server-key.pem
```



### 根据证书文件创建secret

```shell
kubectl create secret generic sample-ssl --from-file=tls.crt=server.pem --
from-file=tls.key=server-key.pem --from-file=ca.crt=ca.pem --
type=kubernetes.io/tls
```

### 配置RadonDB MySQL 集群使用`TLS`

```shell
kubectl patch mysqlclusters.mysql.radondb.com sample  --type=merge -p '{"spec":{"tlsSecretName":"sample-ssl"}}'
```

> 配置之后会触发`rolling update`即集群会重启

### 验证测试

* 不使用`SSL`连接

  ```shell
  kubectl exec -it sample-mysql-0 -c mysql -- mysql -uradondb_usr -p"RadonDB@123"  -e "\s"
  mysql  Ver 14.14 Distrib 5.7.34-37, for Linux (x86_64) using  7.0
  Connection id:          7940
  Current database:
  Current user:           radondb_usr@localhost
  SSL:                    Not in use
  Current pager:          stdout
  Using outfile:          ''
  Using delimiter:        ;
  Server version:         5.7.34-37-log Percona Server (GPL), Release 37, Revision 7c516e9
  Protocol version:       10
  Connection:             Localhost via UNIX socket
  Server characterset:    utf8mb4
  Db     characterset:    utf8mb4
  Client characterset:    latin1
  Conn.  characterset:    latin1
  UNIX socket:            /var/lib/mysql/mysql.sock
  Uptime:                 21 hours 49 min 36 sec
  
  Threads: 5  Questions: 181006  Slow queries: 0  Opens: 127  Flush tables: 1  Open tables: 120  Queries per second avg: 2.303
  ```

  

* 使用`SSL`连接

```shell
 kubectl exec -it sample-mysql-0 -c mysql -- mysql -uradondb_usr -p"RadonDB@123" --ssl-mode=REQUIRED -e "\s"
mysql: [Warning] Using a password on the command line interface can be insecure.
--------------
mysql  Ver 14.14 Distrib 5.7.34-37, for Linux (x86_64) using  7.0

Connection id:          7938
Current database:
Current user:           radondb_usr@localhost
SSL:                    Cipher in use is ECDHE-RSA-AES128-GCM-SHA256
Current pager:          stdout
Using outfile:          ''
Using delimiter:        ;
Server version:         5.7.34-37-log Percona Server (GPL), Release 37, Revision 7c516e9
Protocol version:       10
Connection:             Localhost via UNIX socket
Server characterset:    utf8mb4
Db     characterset:    utf8mb4
Client characterset:    latin1
Conn.  characterset:    latin1
UNIX socket:            /var/lib/mysql/mysql.sock
Uptime:                 21 hours 49 min 26 sec

Threads: 5  Questions: 180985  Slow queries: 0  Opens: 127  Flush tables: 1  Open tables: 120  Queries per second avg: 2.303
```
