#!/bin/bash

# usage: file_env VAR [DEFAULT]
# ie: file_env 'XYZ_DB_PASSWORD' 'example'
# (will allow for "$XYZ_DB_PASSWORD_FILE" to fill in the value of
#  "$XYZ_DB_PASSWORD" from a file, especially for Docker's secrets feature)
file_env() {
	local var="$1"
	local fileVar="${var}_FILE"
	local def="${2:-}"
	if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
		echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
		exit 1
	fi
	local val="$def"
	if [ "${!var:-}" ]; then
		val="${!var}"
	elif [ "${!fileVar:-}" ]; then
		val="$(< "${!fileVar}")"
	fi
	export "$var"="$val"
	unset "$fileVar"
}

host=$(hostname)

file_env 'HOST_SUFFIX' ''

if [ -n "$HOST_SUFFIX" ]; then
	host=$host.$HOST_SUFFIX
fi

file_env 'MYSQL_REPL_PASSWORD' 'Repl_123'

printf '{
 "log": {
  "level": "INFO"
 },
 "server": {
  "endpoint": "%s:8801"
 },
 "replication": {
  "passwd": "%s",
  "user": "qc_repl"
 },
 "rpc": {
  "request-timeout": 2000
 },
 "mysql": {
  "admit-defeat-ping-count": 3,
  "admin": "root",
  "basedir": "/usr",
  "defaults-file": "/etc/mysql/my.cnf",
  "ping-timeout": 2000,
  "passwd": "",
  "host": "localhost",
  "version": "mysql57",
  "master-sysvars": "tokudb_fsync_log_period=default;sync_binlog=default;innodb_flush_log_at_trx_commit=default",
  "slave-sysvars": "tokudb_fsync_log_period=1000;sync_binlog=1000;innodb_flush_log_at_trx_commit=1",
  "port": 3306
 },
 "raft": {
  "election-timeout": 10000,
  "admit-defeat-hearbeat-count": 5,
  "heartbeat-timeout": 2000,
  "meta-datadir": "/var/lib/xenon/",
  "leader-start-command": "",
  "leader-stop-command": "",
  "semi-sync-degrade": true,
  "purge-binlog-disabled": true,
  "super-idle": false
 }
}' $host $MYSQL_REPL_PASSWORD > /etc/xenon/xenon.json

chown -R mysql:mysql /etc/xenon/xenon.json

exec "$@"