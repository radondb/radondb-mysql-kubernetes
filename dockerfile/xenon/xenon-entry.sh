#!/bin/sh
set -e

# Indirect expansion (ie) is not supported in bourne shell. That's why we are using this "magic" here.
ie_gv() {
	eval "echo \$$1"
}

# usage: file_env VAR [DEFAULT]
# ie: file_env 'XYZ_DB_PASSWORD' 'example'
# (will allow for "$XYZ_DB_PASSWORD_FILE" to fill in the value of
#  "$XYZ_DB_PASSWORD" from a file, especially for Docker's secrets feature)
file_env() {
	local var="$1"
	local fileVar="${var}_FILE"
	local def="${2:-}"

	if [ "$(ie_gv ${var})" != ""  ] && [ "$(ie_gv ${fileVar})" != "" ]; then
		in_error "Both $var and $fileVar are set (but are exclusive)"
	fi

	local val="$def"
	if [ "$(ie_gv ${var})" != "" ]; then
		val=$(ie_gv ${var})
	elif [ "$(ie_gv ${fileVar})" != "" ]; then
		val=`cat $(ie_gv ${fileVar})`
	fi

	export "$var"="$val"
	unset "$fileVar"
}

build_conf() {
	if [ -f /etc/xenon/xenon.json ]; then
		return
	fi

	file_env 'HOST' $(hostname)
	file_env 'MYSQL_REPL_PASSWORD' 'Repl_123'
	file_env 'MYSQL_ROOT_PASSWORD' ''
	file_env 'Master_SysVars' ''
	file_env 'Slave_SysVars' ''
	file_env 'LEADER_START_CMD' ':'
	file_env 'LEADER_STOP_CMD' ':'

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
		"ping-timeout": 2000,
		"passwd": "%s",
		"host": "localhost",
		"version": "mysql57",
		"master-sysvars": "%s",
		"slave-sysvars": "%s",
		"port": 3306,
		"monitor-disabled": true
	},
	"raft": {
		"election-timeout": 10000,
		"admit-defeat-hearbeat-count": 5,
		"heartbeat-timeout": 2000,
		"meta-datadir": "/var/lib/xenon/",
		"leader-start-command": "%s",
		"leader-stop-command": "%s",
		"semi-sync-degrade": true,
		"purge-binlog-disabled": true,
		"super-idle": false
	}
}
' "$HOST" "$MYSQL_REPL_PASSWORD" "$MYSQL_ROOT_PASSWORD" "$Master_SysVars" "$Slave_SysVars" "$LEADER_START_CMD" "$LEADER_STOP_CMD" > /etc/xenon/xenon.json
}

ping_host(){
	# wait for the host ready.
	# see https://github.com/kubernetes/kubernetes/issues/92559
	max=30
	for i in `seq 1 $max`
	do
		if ping -c 1 -W 1 $HOST > /dev/null
		then
			echo "ping $HOST success."
			break
		else
			if [ $i -eq $max ]
			then
				echo >&2 "cannot ping $HOST."
				exit 1
			fi
			sleep 1
		fi
	done
}

build_conf
ping_host
exec "$@"
