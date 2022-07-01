English | [简体中文](../zh-cn/how_to_use_tls.md)

Contents
=============
   * [Enable encrypted connection for MySQL client](#Enable-encrypted-connection-for-MySQL-client)
      * [`TLS` overview](#TLS-overview)
      * [Configure `MySQL Operator` for encrypted connection](#Configure-MySQL-Operator-for-encrypted-connection)
         * [Prepare certificates](#Prepare-certificates)
         * [Create Secret with the certificate files](#Create-Secret-with-the-certificate-files)
         * [Configure the RadonDB MySQL cluster to use TLS](#Configure-the-RadonDB-MySQL-cluster-to-use-TLS)
         * [Verification](#Verification)

# Enable encrypted connection for MySQL client

## `TLS` overview

RadonDB MySQL Operator uses non-encrypted connections by default. Third-party tools able to sniff and monitor the network may intercept data transferred between the server and client, leading to the leakage of information. Therefore, you are advised to enable the encrypted connection for security.

The RadonDB MySQL Operator server supports connections based on the TLS (Transport Layer Security). The protocol is supported by MySQL. RadonDB version `5.7` supports `TLS 1.0`, `1.1`, and `1.2`. Version `8.0` supports `TLS 1.0`, `1.1`,` 1.2`, and `1.3`.

Two requirements for encrypted connection:

* Encrypted connection is enabled on the MySQL Operator server.
* Encrypted connection is used by the client.

## Configure `MySQL Operator` for encrypted connection

### Prepare certificates

* `ca.crt` - Server `CA` certificate
* `tls.key` - Private key of server certificate
* `tls.crt` - Server certificate

The certificates and keys can be generated with `OpenSSL`, or simply with `mysql_ssl_rsa_setup` included in `MySQL`.

`mysql_ssl_rsa_setup --datadir=/tmp/certs`

The following files are generated:

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

### Create Secret with the certificate files

```shell
kubectl create secret generic sample-ssl --from-file=tls.crt=server.pem --
from-file=tls.key=server-key.pem --from-file=ca.crt=ca.pem --
type=kubernetes.io/tls
```

### Configure the RadonDB MySQL cluster to use TLS

```shell
kubectl patch mysqlclusters.mysql.radondb.com sample  --type=merge -p '{"spec":{"tlsSecretName":"sample-ssl"}}'
```

> The configuration will trigger `rolling update` and the cluster will restart.

### Verification

* Non-`SSL` connection

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

  

* `SSL` connection

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