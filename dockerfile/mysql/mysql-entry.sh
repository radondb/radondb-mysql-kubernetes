#!/bin/bash

# usage: file_env VAR [DEFAULT]
#	ie: file_env 'XYZ_DB_PASSWORD' 'example'
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

# usage: process_init_file FILENAME MYSQLCOMMAND...
#	ie: process_init_file foo.sh mysql -uroot
# (process a single initializer file, based on its extension. we define this
# function here, so that initializer scripts (*.sh) can use the same logic,
# potentially recursively, or override the logic used in subsequent calls)
process_init_file() {
	local f="$1"; shift
	local mysql=( "$@" )

	case "$f" in
		*.sh)	 echo "$0: running $f"; . "$f" ;;
		*.sql)	echo "$0: running $f"; "${mysql[@]}" < "$f"; echo ;;
		*.sql.gz) echo "$0: running $f"; gunzip -c "$f" | "${mysql[@]}"; echo ;;
		*)		echo "$0: ignoring $f" ;;
	esac
	echo
}

# Fetch value from server config
# We use mysqld --verbose --help instead of my_print_defaults because the
# latter only show values present in config files, and not server defaults
_get_config() {
	local conf="$1";
	"mysqld" --verbose --help --log-bin-index="$(mktemp -u)" 2>/dev/null \
		| awk '$1 == "'"$conf"'" && /^[^ \t]/ { sub(/^[^ \t]+[ \t]+/, ""); print; exit }'
	# match "datadir	  /some/path with/spaces in/it here" but not "--xyz=abc\n	 datadir (xyz)"
}

if [ -n "$INIT_TOKUDB" ]; then
	export LD_PRELOAD=/usr/lib/x86_64-linux-gnu/libjemalloc.so.1
fi
# Get config
DATADIR="$(_get_config 'datadir')"

if [ ! -d "$DATADIR/mysql" ]; then
	mkdir -p "$DATADIR"

	echo 'Initializing database'
	mysqld --initialize-insecure --skip-ssl
	echo 'Database initialized'

	if command -v mysql_ssl_rsa_setup > /dev/null && [ ! -e "$DATADIR/server-key.pem" ]; then
		# https://github.com/mysql/mysql-server/blob/23032807537d8dd8ee4ec1c4d40f0633cd4e12f9/packaging/deb-in/extra/mysql-systemd-start#L81-L84
		echo 'Initializing certificates'
		mysql_ssl_rsa_setup --datadir="$DATADIR"
		echo 'Certificates initialized'
	fi

	SOCKET="$(_get_config 'socket')"
	"mysqld" --skip-networking --socket="${SOCKET}" &
	pid="$!"

	mysql=( mysql --protocol=socket -uroot -hlocalhost --socket="${SOCKET}" --password="" )

	for i in {120..0}; do
		if echo 'SELECT 1' | "${mysql[@]}" &> /dev/null; then
			break
		fi
		echo 'MySQL init process in progress...'
		sleep 1
	done
	if [ "$i" = 0 ]; then
		echo >&2 'MySQL init process failed.'
		exit 1
	fi

	if [ -z "$MYSQL_INITDB_SKIP_TZINFO" ]; then
		# sed is for https://bugs.mysql.com/bug.php?id=20545
		mysql_tzinfo_to_sql /usr/share/zoneinfo | sed 's/Local time zone must be set--see zic manual page/FCTY/' | "${mysql[@]}" mysql
	fi

	# install TokuDB engine
	if [ -n "$INIT_TOKUDB" ]; then
		ps-admin --docker --enable-tokudb -u root
	fi

	"${mysql[@]}" <<-EOSQL
		-- What's done in this file shouldn't be replicated
		-- or products like mysql-fabric won't work
		SET @@SESSION.SQL_LOG_BIN=0;
		DELETE FROM mysql.user WHERE user NOT IN ('mysql.sys', 'root') OR host NOT IN ('localhost') ;
		CREATE USER 'root'@'127.0.0.1' IDENTIFIED BY '' ;
		GRANT ALL ON *.* TO 'root'@'127.0.0.1' WITH GRANT OPTION ;
		DROP DATABASE IF EXISTS test ;
		FLUSH PRIVILEGES ;
	EOSQL

	file_env 'MYSQL_REPL_PASSWORD' 'Repl_123'
	echo "GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* to 'qc_repl'@'%' IDENTIFIED BY '$MYSQL_REPL_PASSWORD' ;" | "${mysql[@]}"
	echo 'FLUSH PRIVILEGES ;' | "${mysql[@]}"

	file_env 'MYSQL_USER' 'qingcloud'
	echo "MySQL USER: $MYSQL_USER"
	if [ $MYSQL_USER = "root" -o $MYSQL_USER = "qc_repl" ]; then
		echo >&2 'Donot set MYSQL_USER as root or qc_repl.'
		exit 1
	fi

	file_env 'MYSQL_PASSWORD' 'qingcloud'

	file_env 'MYSQL_DATABASE'
	if [ "$MYSQL_DATABASE" ]; then
		echo "CREATE DATABASE IF NOT EXISTS \`$MYSQL_DATABASE\` ;" | "${mysql[@]}"
		mysql+=( "$MYSQL_DATABASE" )
	fi

	echo "CREATE USER '$MYSQL_USER'@'%' IDENTIFIED BY '$MYSQL_PASSWORD' ;" | "${mysql[@]}"
	if [ "$MYSQL_DATABASE" ]; then
		echo "GRANT ALL ON \`$MYSQL_DATABASE\`.* TO '$MYSQL_USER'@'%' ;" | "${mysql[@]}"
	fi
	echo 'FLUSH PRIVILEGES ;' | "${mysql[@]}"
	echo 'reset master;' | "${mysql[@]}"

	if ! kill -s TERM "$pid" || ! wait "$pid"; then
		echo >&2 'MySQL init process failed.'
		exit 1
	fi

	sed -i '/server-id/d' /etc/mysql/my.cnf
	chown -R mysql:mysql "$DATADIR"
fi

rm -f /var/log/mysql/error.log
rm -f /var/lib/mysql/auto.cnf

uuid=$(cat /proc/sys/kernel/random/uuid)
printf '[auto]\nserver_uuid=%s' $uuid > /var/lib/mysql/auto.cnf

echo
echo 'MySQL init process done.'
echo

exec "$@"
