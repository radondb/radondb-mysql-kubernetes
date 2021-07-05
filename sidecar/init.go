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

package sidecar

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// NewInitCommand return a pointer to cobra.Command.
func NewInitCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "do some initialization operations.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runInitCommand(cfg); err != nil {
				log.Error(err, "init command failed")
				os.Exit(1)
			}
		},
	}

	return cmd
}

// runInitCommand do some initialization operations.
func runInitCommand(cfg *Config) error {
	var err error

	if exists, _ := checkIfPathExists(dataPath); exists {
		// remove lost+found.
		if err := os.RemoveAll(dataPath + "/lost+found"); err != nil {
			return fmt.Errorf("removing lost+found: %s", err)
		}

		// Get the mysql user.
		user, err := user.Lookup("mysql")
		if err != nil {
			return fmt.Errorf("failed to get mysql user: %s", err)
		}
		uid, err := strconv.Atoi(user.Uid)
		if err != nil {
			return fmt.Errorf("failed to get mysql user uid: %s", err)
		}
		gid, err := strconv.Atoi(user.Gid)
		if err != nil {
			return fmt.Errorf("failed to get mysql user gid: %s", err)
		}
		// chown -R mysql:mysql /var/lib/mysql.
		if err = os.Chown(dataPath, uid, gid); err != nil {
			return fmt.Errorf("failed to chown %s: %s", dataPath, err)
		}
	}

	if err = os.Mkdir(configPath, os.FileMode(0755)); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("error mkdir %s: %s", configPath, err)
		}
	}

	// Run reset master in init-mysql container.
	if err = ioutil.WriteFile(initFilePath+"/reset.sql", []byte("reset master;"), 0644); err != nil {
		return fmt.Errorf("failed to write reset.sql: %s", err)
	}

	// build init.sql.
	initSqlPath := path.Join(configPath, "init.sql")
	if err = ioutil.WriteFile(initSqlPath, buildInitSql(cfg), 0644); err != nil {
		return fmt.Errorf("failed to write init.sql: %s", err)
	}

	// build extra.cnf.
	extraConfig, err := cfg.buildExtraConfig(initSqlPath)
	if err != nil {
		return fmt.Errorf("failed to build extra.cnf: %s", err)
	}
	// save extra.cnf to conf.d.
	if err := extraConfig.SaveTo(path.Join(configPath, "extra.cnf")); err != nil {
		return fmt.Errorf("failed to save extra.cnf: %s", err)
	}

	// copy leader-start.sh from config-map to scripts mount.
	leaderStartPath := path.Join(scriptsPath, "leader-start.sh")
	if err = copyFile(path.Join(configMapPath, "leader-start.sh"), leaderStartPath); err != nil {
		return fmt.Errorf("failed to copy scripts: %s", err)
	}
	if err = os.Chmod(leaderStartPath, os.FileMode(0755)); err != nil {
		return fmt.Errorf("failed to chmod scripts: %s", err)
	}

	// copy leader-stop.sh from config-map to scripts mount.
	leaderStopPath := path.Join(scriptsPath, "leader-stop.sh")
	if err = copyFile(path.Join(configMapPath, "leader-stop.sh"), leaderStopPath); err != nil {
		return fmt.Errorf("failed to copy scripts: %s", err)
	}
	if err = os.Chmod(leaderStopPath, os.FileMode(0755)); err != nil {
		return fmt.Errorf("failed to chmod scripts: %s", err)
	}

	// for install tokudb.
	if cfg.InitTokuDB {
		arg := fmt.Sprintf("echo never > %s/enabled", sysPath)
		cmd := exec.Command("sh", "-c", arg)
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to disable the transparent_hugepage: %s", err)
		}
	}

	// build xenon.json.
	xenonFilePath := path.Join(xenonPath, "xenon.json")
	if err = ioutil.WriteFile(xenonFilePath, buildXenonConf(cfg), 0644); err != nil {
		return fmt.Errorf("failed to write xenon.json: %s", err)
	}

	log.Info("init command success")
	return nil
}

// checkIfPathExists check if the path exists.
func checkIfPathExists(path string) (bool, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		log.Error(err, "failed to open file", "file", path)
		return false, err
	}

	err = f.Close()
	return true, err
}

// buildXenonConf build a config file for xenon.
func buildXenonConf(cfg *Config) []byte {
	pingTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount
	heartbeatTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount
	requestTimeout := cfg.ElectionTimeout / cfg.AdmitDefeatHearbeatCount

	version := "mysql80"
	if cfg.MySQLVersion.Major == 5 {
		if cfg.MySQLVersion.Minor == 6 {
			version = "mysql56"
		} else {
			version = "mysql57"
		}
	}

	var masterSysVars, slaveSysVars string
	if cfg.InitTokuDB {
		masterSysVars = "tokudb_fsync_log_period=default;sync_binlog=default;innodb_flush_log_at_trx_commit=default"
		slaveSysVars = "tokudb_fsync_log_period=1000;sync_binlog=1000;innodb_flush_log_at_trx_commit=1"
	} else {
		masterSysVars = "sync_binlog=default;innodb_flush_log_at_trx_commit=default"
		slaveSysVars = "sync_binlog=1000;innodb_flush_log_at_trx_commit=1"
	}

	hostName := fmt.Sprintf("%s.%s.%s", cfg.HostName, cfg.ServiceName, cfg.NameSpace)

	str := fmt.Sprintf(`{
    "log": {
        "level": "INFO"
    },
    "server": {
        "endpoint": "%s:%d",
        "peer-address": "%s:%d",
        "enable-apis": true
    },
    "replication": {
        "passwd": "%s",
        "user": "%s"
    },
    "rpc": {
        "request-timeout": %d
    },
    "mysql": {
        "admit-defeat-ping-count": 3,
        "admin": "root",
        "ping-timeout": %d,
        "passwd": "%s",
        "host": "localhost",
        "version": "%s",
        "master-sysvars": "%s",
        "slave-sysvars": "%s",
        "port": 3306,
        "monitor-disabled": true
    },
    "raft": {
        "election-timeout": %d,
        "admit-defeat-hearbeat-count": %d,
        "heartbeat-timeout": %d,
        "meta-datadir": "/var/lib/xenon/",
        "leader-start-command": "/scripts/leader-start.sh",
        "leader-stop-command": "/scripts/leader-stop.sh",
        "semi-sync-degrade": true,
        "purge-binlog-disabled": true,
        "super-idle": false
    }
}
`, hostName, utils.XenonPort, hostName, utils.XenonPeerPort, cfg.ReplicationPassword, cfg.ReplicationUser, requestTimeout,
		pingTimeout, cfg.RootPassword, version, masterSysVars, slaveSysVars, cfg.ElectionTimeout,
		cfg.AdmitDefeatHearbeatCount, heartbeatTimeout)
	return utils.StringToBytes(str)
}

// buildInitSql used to build init.sql. The file run after the mysql init.
func buildInitSql(cfg *Config) []byte {
	sql := fmt.Sprintf(`SET @@SESSION.SQL_LOG_BIN=0;
DELETE FROM mysql.user WHERE user='%s';
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* to '%s'@'%%' IDENTIFIED BY '%s';
DELETE FROM mysql.user WHERE user='%s';
GRANT SELECT, PROCESS, REPLICATION CLIENT ON *.* to '%s'@'%%' IDENTIFIED BY '%s';
DELETE FROM mysql.user WHERE user='%s';
GRANT SUPER, PROCESS, RELOAD, CREATE, SELECT ON *.* to '%s'@'%%' IDENTIFIED BY '%s';
FLUSH PRIVILEGES;
`, cfg.ReplicationUser, cfg.ReplicationUser, cfg.ReplicationPassword, cfg.MetricsUser, cfg.MetricsUser,
		cfg.MetricsPassword, cfg.OperatorUser, cfg.OperatorUser, cfg.OperatorPassword)

	return utils.StringToBytes(sql)
}
