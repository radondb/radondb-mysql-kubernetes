/*
Copyright 2021 RadonDB.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package syncer

import (
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// log is for logging in this package.
// var log = logf.Log.WithName("mysqlcluster.syncer")

// mysqlSysConfigs is the map of mysql system configs.
var mysqlSysConfigs = map[string]string{
	"default-time-zone":   "+08:00",
	"slow_query_log_file": "/var/log/mysql/mysql-slow.log",
	"read_only":           "ON",
	"binlog_format":       "row",
	"log-bin":             "/var/lib/mysql/mysql-bin",
	"log-timestamps":      "SYSTEM",
	"innodb_open_files":   "655360",
	"open_files_limit":    "655360",

	"gtid-mode":                "ON",
	"enforce-gtid-consistency": "ON",
	"slave_parallel_type":      "LOGICAL_CLOCK",
	"relay_log":                "/var/lib/mysql/mysql-relay-bin",
	"relay_log_index":          "/var/lib/mysql/mysql-relay-bin.index",
	"slow_query_log":           "1",
	"tmp_table_size":           "32M",
	"tmpdir":                   "/var/lib/mysql",
}

var pluginConfigs = map[string]string{
	"plugin-load": "\"semisync_master.so;semisync_slave.so;audit_log.so;connection_control.so\"",

	"rpl_semi_sync_master_enabled":       "OFF",
	"rpl_semi_sync_slave_enabled":        "ON",
	"rpl_semi_sync_master_wait_no_slave": "ON",
	"rpl_semi_sync_master_timeout":       "1000000000000000000",

	"audit_log_file":             "/var/log/mysql/mysql-audit.log",
	"audit_log_exclude_accounts": "\"root@localhost,root@127.0.0.1," + utils.ReplicationUser + "@%," + utils.MetricsUser + "@%\"",
	"audit_log_buffer_size":      "16M",
	"audit_log_policy":           "NONE",
	"audit_log_rotate_on_size":   "104857600",
	"audit_log_rotations":        "6",
	"audit_log_format":           "OLD",

	"connection_control_failed_connections_threshold": "3",
	"connection_control_min_connection_delay":         "1000",
	"connection_control_max_connection_delay":         "2147483647",
}

var mysql57Configs = map[string]string{
	"query-cache-type": "0",
	"query-cache-size": "0",
	"sql-mode": "STRICT_TRANS_TABLES,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER," +
		"NO_AUTO_VALUE_ON_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE,ONLY_FULL_GROUP_BY",

	"expire-logs-days": "7",

	"master_info_repository":       "TABLE",
	"relay_log_info_repository":    "TABLE",
	"slave_rows_search_algorithms": "INDEX_SCAN,HASH_SCAN",
}

var mysql80Configs = map[string]string{
	"sql-mode": "STRICT_TRANS_TABLES,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_VALUE_ON_ZERO,NO_ENGINE_SUBSTITUTION," +
		"NO_ZERO_DATE,NO_ZERO_IN_DATE,ONLY_FULL_GROUP_BY",
	// 7 days = 7 * 24 * 60 * 60
	"binlog_expire_logs_seconds": "604800",
	// use 5.7 auth plugin to be backward compatible
	"default-authentication-plugin": "mysql_native_password",
}

// mysqlCommonConfigs is the map of the mysql common configs.
var mysqlCommonConfigs = map[string]string{
	"character_set_server":            "utf8mb4",
	"interactive_timeout":             "3600",
	"default-time-zone":               "+08:00",
	"key_buffer_size":                 "33554432",
	"log_bin_trust_function_creators": "1",
	"long_query_time":                 "3",
	"binlog_cache_size":               "32768",
	"binlog_stmt_cache_size":          "32768",
	"max_connections":                 "1024",
	"max_connect_errors":              "655360",
	"sync_master_info":                "1000",
	"sync_relay_log":                  "1000",
	"sync_relay_log_info":             "1000",
	"table_open_cache":                "2000",
	"thread_cache_size":               "128",
	"wait_timeout":                    "3600",
	"group_concat_max_len":            "1024",
	"max_allowed_packet":              "1073741824",
	"event_scheduler":                 "OFF",
	"innodb_print_all_deadlocks":      "0",
	"autocommit":                      "1",
	"transaction-isolation":           "READ-COMMITTED",

	"explicit_defaults_for_timestamp": "0",
	"innodb_adaptive_hash_index":      "0",
}

// mysqlStaticConfigs is the map of the mysql static configs.
// The mysql need restart, if modify the config.
var mysqlStaticConfigs = map[string]string{
	"default-storage-engine":      "InnoDB",
	"back_log":                    "2048",
	"ft_min_word_len":             "4",
	"lower_case_table_names":      "0",
	"innodb_ft_max_token_size":    "84",
	"innodb_ft_min_token_size":    "3",
	"sql_mode":                    "STRICT_TRANS_TABLES,NO_ENGINE_SUBSTITUTION",
	"slave_parallel_workers":      "8",
	"slave_pending_jobs_size_max": "1073741824",
	"innodb_log_buffer_size":      "16777216",
	"innodb_log_file_size":        "1073741824",
	"innodb_log_files_in_group":   "2",
	"innodb_flush_method":         "O_DIRECT",
	"innodb_use_native_aio":       "1",
	"innodb_autoinc_lock_mode":    "2",
	"performance_schema":          "1",
}

// mysqlTokudbConfigs is the map of the mysql tokudb configs.
var mysqlTokudbConfigs = map[string]string{
	"loose_tokudb_directio": "ON",
}

// mysqlBooleanConfigs is the list of the mysql boolean configs.
var mysqlBooleanConfigs = []string{
	"federated",
	"skip-host-cache",
	"skip-name-resolve",
	"core-file",
	"skip-slave-start",
	"log-slave-updates",
	"!includedir /etc/mysql/conf.d",
}

// mysqlSSLConfigs is the ist of the mysql ssl configs.
var mysqlSSLConfigs = map[string]string{
	"ssl_ca":   "/etc/mysql-ssl/ca.crt",
	"ssl_cert": "/etc/mysql-ssl/tls.crt",
	"ssl_key":  "/etc/mysql-ssl/tls.key",
}
