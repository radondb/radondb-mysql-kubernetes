#!/bin/bash
#
# gh-mysql-rewind -m <master-host> [-x] [-r]
#   Rewind errant transactions on local server and rewire to replicate from master-host
#   -m master-host, serves as GTID baseline
#   -x execute (default is noop)
#   -r auto-start replication upon rewiring
#
# The objective of this tool is to fix a MySQL server which has errant
# transactions by reverting it far enough into the past such that errant trnsactions
# are reverted (rewinded), and run the correct GTID bookkeeping so as to be able to
# connect the server as a replica and consume replication events from the correct position.
#
# The tool solves two use cases:
# - A replica which as accidentally contaminated by DML, e.g. rows were deleted directly on replica.
# - A split brain master: a scenario where a failover process demoted master A and promoted master B,
#   even as demoted master A continued to receive some traffic.
#
# gh-mysql-rewind:
# - Operates on a single MySQL server
# - Requires identify of a master-host, a server from which the operated server will end up replicating.
# - Auto-detects what needs to be fixed
# - Reverts entire binlog files: either the last binlog file, or the last two binlog files, or... the last n binlog files
# - Uses MariaDB's `mysqlbinlog --flashback` functionality.
# - Probably rewinds farther into the past than strictly needed
# - Keeps record of the GTID entries which have been rewinded/reverted
# - After de-applying the binlog files, issues `reset master` and fixes `executed_gtid_set`, `gtid_purged` on operated server
# - Cleans up relay log files and refreshes replication
# - Potentially auto-starts replication.

# Users: look for the 'IMPLEMENTATION' keyword in the code to find specific implementation hints.


# IMPLEMENTATION
# Provide the following credentials for remote MySQL connection (user only need to query @@global.gtid_executed)
master_user="$GH_MYSQL_REWIND_MASTER_RO_USER"
master_password="$GH_MYSQL_REWIND_MASTER_RO_PASSWORD"
# Provide the following credentials for local MySQL connection (user needs many privileges, ideally this would be a ALL PRIVILEGES account)
local_user="$GH_MYSQL_REWIND_LOCAL_POWER_USER"
local_password="$GH_MYSQL_REWIND_LOCAL_POWER_PASSWORD"
# where binary logs are found on your MySQL server
binlog_path="/var/lib/mysql"
# Where the MariaDB `mysqlbinlog` binary is found on your MySQL server.
# Not to be confused with the MySQL `mysqlbinlog` binary, which is also expected to be found on your server.
mariadb_mysqlbinlog="mariadb-mysqlbinlog"
# /IMPLEMENTATION

hostname="$(hostname)"
logpath="/var/log/gh-mysql-rewind/$(date +%s)"
logfile="$logpath/gh-mysql-rewind.log"

master_host=
execute=
resume_replication=
usage_message="gh-mysql-rewind -m <master-host> [-r] [-x]"

tmp_file_prefix="/tmp/gh-mysql-rewind"
local_binary_logs_file="${tmp_file_prefix}.binary-logs"

errant_gtid=""
reverted_gtid=""
reverted_binlogs=""

set -o pipefail

while getopts "m:rx" flag; do
  case $flag in
    m)
      master_host=$OPTARG
      ;;
    r)
      resume_replication="true"
      ;;
    x) # execute, not noop
      execute="true"
      ;;
    *) # unknown
      echo "unknown flag $flag"
      fail "$usage_message"
      ;;
  esac
done


fail() {
  export error_message="${1}"
  (>&2 echo "[ERROR] $(date "+%F %T") $error_message")
  exit 1
}

dependency_checks() {
  for cmd in awk sed jq cat grep date tail tee orchestrator-client mysqlbinlog $mariadb_mysqlbinlog ; do
    command -v "$cmd" >/dev/null 2>&1 || fail "dependency failure: cannot find '$cmd' command"
  done
}

mysql_master() {
  mysql -h "$master_host" -u"$master_user" -p"$master_password" "$@"
}

mysql_local() {
  mysql -h localhost -u"$local_user" -p"$local_password" "$@"
}

reverse() {
  (tac 2> /dev/null || tail -r)
}

sanity_checks() {
  if [ -z "$master_host" ] ; then
    fail "master_host is empty. Provide via -m <master-host>"
  fi

  # IMPLEMENTATION
  # Verify the two hosts (this host and master_host) belong to same logical cluster.
  # This will be implicitly also verified when looking for shared GTID entries,
  # but a check here is clearer.
  # /IMPLEMENTATION

  master_cluster="TODO"
  if [ -z "$master_cluster" ] ; then
    fail "Cannot determine master's cluster"
  fi
  local_cluster="TODO"
  if [ -z "$local_cluster" ] ; then
    fail "Cannot determine local cluster"
  fi
  if [ "$master_cluster" != "$local_cluster" ] ; then
    fail "master_cluster: $master_cluster <> local_cluster: $local_cluster"
  fi

  if ! mysql_local -ss -e "show slave status\G" | grep Slave_IO_Running | awk '{print $2}' | grep -q "No" ; then
    fail "This server ($hostname) is replicating. Stop replication first."
  fi
  if ! mysql_local -ss -e "show slave status\G" | grep Slave_SQL_Running | awk '{print $2}' | grep -q "No" ; then
    fail "This server ($hostname) is replicating. Stop replication first."
  fi

  if mysql_local -ss -e 'select @@global.read_only' | grep 0 ; then
    fail "As safety mechanism, this tool may only run on a @@read_only=1 server"
  fi

  if mysql_local -ss -e 'show processlist' | grep -q 'pt-archive' ; then
    fail "pt-archiver found to be running. Please terminate it first"
  fi

  local count_replicas=$(orchestrator-client -c api -path "instance-replicas/${hostname}/3306" | jq 'length')
  if (( count_replicas != 0 )) ; then
    fail "sanity_checks failed: $hostname has $count_replicas replicas. Must have no replicas"
  fi

  local sql_delay="$(orchestrator-client -c api -path "instance/${hostname}/3306" | jq -r .SQLDelay)"
  if (( sql_delay != 0 )) ; then
    fail "sanity_checks failed: $hostname has sql_delay=$sql_delay. Cannot rewind on delayed servers. Un-delay the server first"
  fi

  # IMPLEMENTATION
  # any other sanity checks and  blockers specific to your environments go here...
  # /IMPLEMENTATION

  # Sanity checks satisfied
  return 0
}

gtid_subtract() {
  subtract="$(mysql_local -ss -e "select gtid_subtract('${1}', '${2}')")"
  echo "$subtract"
}

get_gtid_executed_local() {
  mysql_local -ss -e "select @@global.gtid_executed"
}

get_gtid_executed_on_master() {
  mysql_master -ss -e "select @@global.gtid_executed"
}

dump_local_binary_logs_file() {
  mysql_local -ss -e "show binary logs" | awk '{print $1}' > "$local_binary_logs_file"
}

extract_previous_gtid_entries_from_binlog() {
  local binlog_file="$1"
  # Utilizing the MySQL mysqlbinlog tool (which supports GTID)
  sudo mysqlbinlog "$binlog_path/$binlog_file" | sed '/^SET @@SESSION.GTID_NEXT/q' | sed -n '/Previous-GTIDs/,/^# at /p' | sed '1d;$d' | awk '{print $2}'
}


rewind() {
  echo "rewind"
  echo "------"

  if [ -z "$execute" ] ; then
    echo "-x not specified. noop execution"
  fi

  echo "Completely stopping replication"
  orchestrator-client -c stop-replica -i "$hostname"
  echo "- done"

  gtid_executed_local="$(get_gtid_executed_local)"
  gtid_executed_on_master="$(get_gtid_executed_on_master)"
  errant_gtid="$(gtid_subtract "${gtid_executed_local}" "${gtid_executed_on_master}")"

  if [ -z "$errant_gtid" ] ; then
    fail "No errant GTID detected. Master and replica have identical gtid_executed. Aborting operation."
  fi
  echo "errant_gtid: $errant_gtid"

  shared_gtid="$(gtid_subtract "${gtid_executed_on_master}" "${gtid_executed_local}")"
  if [ -z "$shared_gtid" ] ; then
    fail "Sanity check: there are no shared GTID entries between the two servers. Seems like they were never part of the same cluster."
  fi

  echo "Flushing binary logs to evaluate which binary logs contain the errant transactions"
  # This FLUSH needs to be done even on 'noop', or else we can't reliably evaluate which binary logs contain the errant GTIDs.
  # What we do is to parse the Previous-GTID header in the binary logs. For each binary log, this tells us about the known GTID set
  # in all pevious binary logs. But that leaves the last binary log in a mystery. So what we do is flush binary logs, to generate an
  # additional, empty binary log. But in its header, we will find the infromation about the "mystery" binary log.
  expire_logs_days="$(mysql_local -ss -e "select @@global.expire_logs_days")"
  mysql_local -e "set global expire_logs_days=0"
  mysql_local -e "flush NO_WRITE_TO_BINLOG binary logs"
  echo "+ done"
  echo "Restoring expire_logs_days to $expire_logs_days"
  mysql_local -e "set global expire_logs_days=$expire_logs_days"
  echo "+ done"

  dump_local_binary_logs_file

  # We will rewind the errant GTIDs, but we will rewind more than just those. There will most likely be "normal" transactions rewinded.
  # 'reverted_gtid' is the total set of GTIDs rewinded in this operation. Obviously 'reverted_gtid' contains 'errant_gtid'.
  reverted_gtid=""
  # We will rewind the last n binary logs. 'reverted_binlogs' is the listing of the binary logs we will rewind in this operation.
  reverted_binlogs=""
  found_binlogs_match=""
  # 'initial_previous_gtids' is the Previous-GTIDs we parse out of the newly flushed binary log. This value sums up all the GTIDs contained
  # in all of the binary logs.
  initial_previous_gtids=""
  for binlog_file in $(cat "$local_binary_logs_file" | reverse) ; do
    previous_gtids="$(extract_previous_gtid_entries_from_binlog "$binlog_file")"
    # use gtid_subtract() against the empty string to clean up the gtid string
    previous_gtids="$(gtid_subtract "$previous_gtids" "")"

    subtract="$(gtid_subtract "$errant_gtid" "$previous_gtids")"
    if [ "$reverted_binlogs" == "" ] ; then
      # This is the first binray log that we iterate. ie this is the last binary
      # log generated (we iterate backwards).
      # Also, this must be the binary log we FLUSHed into. It should be empty of any interesting
      # transactions. We expect its "Previous-GTIDs" entry to contain errant GTIDs, or else
      # our binary logs don't have the entries.
      echo "Sanity check: expecting newly rotated binlog $binlog_file to contain errant GTID in its 'Previous-GTIDs' header"
      if [ "$subtract" != "" ] ; then
        fail "Sanity check failed: expecting newly rotated binlog $binlog_file to contain errant GTID in its 'Previous-GTIDs' header"
      fi
      echo "- ok"
      initial_previous_gtids="$previous_gtids"
    fi
    # 'reverted_gtid' is re-evaluated for each binary log. It is the sum of all GTIDs in all last-n binary logs.
    # Instead of parsing the contents of those binary logs and summing up the GTIDs contained in them, we
    # subtract initial_previous_gtids-previous_gtids (all known GTIDs - all GTIDs in preceeding binary logs)
    reverted_gtid="$(gtid_subtract "$initial_previous_gtids" "$previous_gtids")"
    reverted_binlogs="$reverted_binlogs $binlog_file"

    echo "Checking if errant transactions are contained by $reverted_binlogs"
    if [ "$subtract" == "$errant_gtid" ] ; then
      found_binlogs_match="true"
      echo "- yes. Search for binary logs is over."
      break
    fi
    echo "- no"
  done
  if [ -z "$found_binlogs_match" ] ; then
    fail "Unable to find binlogs match"
  fi

  echo "projecting gtid_executed once we rewind $reverted_binlogs and revert $reverted_gtid"
  projected_reverted_gtid_executed="$(gtid_subtract "$gtid_executed_local" "$reverted_gtid")"

  # For auditing purposes:
  echo "$errant_gtid" | sudo tee "$logpath/errant_gtid" > /dev/null
  echo "$reverted_gtid" | sudo tee "$logpath/reverted_gtid" > /dev/null
  (cd "$binlog_path" ; sudo cp $reverted_binlogs "$logpath/")

  echo "Will generate and apply flashback"

  # We could choose to revert all binary logs at once, but reverting them one by one (from newest, going backwards)
  # can make it simpler to later audit the steps taken.
  for binlog_file in $reverted_binlogs ; do
    binlog_bare_flashback_file="${tmp_file_prefix}.flashback.$binlog_file"
    binlog_rich_flashback_file="${tmp_file_prefix}.flashback-with-gtid.$binlog_file"

    echo "- generating flashback for $binlog_file"
    sudo $mariadb_mysqlbinlog --flashback "$binlog_path/$binlog_file" > "$binlog_bare_flashback_file"
    echo "  - done. See $binlog_bare_flashback_file"
    # MariaDB strips out any MySQL GTID information from the binary logs.
    # To apply the flashbacked content into mysql, we must generate GTID entries. We create a random UUID via `uuidgen -r`
    # and inject a 'SET @@SESSION.GTID_NEXT' for each binlog event
    echo "- Enriching flashback file with dummy GTID statements"
    cat "$binlog_bare_flashback_file" | awk 'BEGIN {"uuidgen -r" |& getline u} /^BEGIN/ {c += 1 ; print "SET @@SESSION.GTID_NEXT= \x27" u ":" c "\x27/*!*/;"} {print}' | sed -e s/', @@session.check_constraint_checks=1//g' > "$binlog_rich_flashback_file"
    echo "  - done. See $binlog_rich_flashback_file"
    if [ "$execute" == "true" ] ; then
      echo "- Applying flashback"
      cat "$binlog_rich_flashback_file" | mysql_local --force
      echo "  - done"
    else
      echo "- noop: not applying flashback file"
    fi
  done

  echo "Summary of flashback GTID content (GTID entries injected onto server, expect different UUID for each binary log):"
  grep "SESSION.GTID_NEXT" ${tmp_file_prefix}.flashback-with-gtid.* | sed -r -e 's/^.*SESSION.GTID_NEXT= .(.*:).*$/\1?/p' | sort | uniq -c

  if [ "$execute" == "true" ] ; then
    echo "resetting master"
    mysql_local -e "reset master"
    echo "- done"
    echo "setting gtid_purged to $projected_reverted_gtid_executed"
    mysql_local -e "set global gtid_purged = '${projected_reverted_gtid_executed}'"
    echo "- done"
    echo "Running CHANGE MASTER TO to get rid of relay logs and to point to $master_host. Relay logs will be deleted because replication is stopped"
    mysql_local -e "change master to master_host='$master_host', master_auto_position=1"
    echo "- done"

    if [ "$resume_replication" == "true" ] ; then
      echo "staring replication"
      orchestrator-client -c start-replica -i "${hostname}"
      echo "- done"
    fi
  else
    echo "- noop: not resetting master"
  fi

  echo "all done"
  return 0
}

main() {
  echo "logfile: $logfile"
  sanity_checks
  rm -f ${tmp_file_prefix:-/tmp/no-such-file}.*
  rewind
}

dependency_checks
sudo mkdir -p "$logpath"
main 2>&1 | ts | sudo tee -a $logfile