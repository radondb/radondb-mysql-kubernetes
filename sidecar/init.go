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
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	"github.com/spf13/cobra"
)

// NewInitCommand return a pointer to cobra.Command.
func NewInitCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "do some initialization operations.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCloneAndInit(cfg); err != nil {
				log.Error(err, "clone error")
			}
			if err := runInitCommand(cfg); err != nil {
				log.Error(err, "init command failed")
				os.Exit(1)
			}
		},
	}

	return cmd
}

// Check leader or follower backup status is ok.
func CheckServiceExist(cfg *Config, service string) bool {
	serviceURL := fmt.Sprintf("http://%s-%s:%v%s", cfg.ClusterName, service, utils.XBackupPort, "/health")
	req, err := http.NewRequest("GET", serviceURL, nil)
	if err != nil {
		log.Info("failed to check available service", "service", serviceURL, "error", err)
		return false
	}

	client := &http.Client{}
	client.Transport = transportWithTimeout(serverConnectTimeout)
	resp, err := client.Do(req)
	if err != nil {
		log.Info("service was not available", "service", serviceURL, "error", err)
		return false
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		log.Info("service not available", "service", serviceURL, "HTTP status code", resp.StatusCode)
		return false
	}

	return true
}

// Clone from leader or follower.
func runCloneAndInit(cfg *Config) error {
	//check follower is exists?
	serviceURL := ""
	if len(serviceURL) == 0 && CheckServiceExist(cfg, "follower") {
		serviceURL = fmt.Sprintf("http://%s-%s:%v", cfg.ClusterName, "follower", utils.XBackupPort)
	}
	//check leader is exist?
	if len(serviceURL) == 0 && CheckServiceExist(cfg, "leader") {
		serviceURL = fmt.Sprintf("http://%s-%s:%v", cfg.ClusterName, "leader", utils.XBackupPort)
	}

	if len(serviceURL) != 0 {
		// backup at first
		Args := fmt.Sprintf("rm -rf /backup/initbackup;mkdir -p /backup/initbackup;curl --user $BACKUP_USER:$BACKUP_PASSWORD %s/download|xbstream -x -C /backup/initbackup; exit ${PIPESTATUS[0]}",
			serviceURL)
		cmd := exec.Command("/bin/bash", "-c", "--", Args)
		log.Info("runCloneAndInit", "cmd", Args)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to disable the run restore: %s", err)
		}
		cfg.XRestoreFrom = backupInitDirectory
		cfg.CloneFlag = true
		return nil
	}
	log.Info("no leader or follower found")
	return nil
}

// runInitCommand do some initialization operations.
func runInitCommand(cfg *Config) error {
	var err error
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

	if exists, _ := checkIfPathExists(dataPath); exists {
		// remove lost+found.
		if err := os.RemoveAll(dataPath + "/lost+found"); err != nil {
			return fmt.Errorf("removing lost+found: %s", err)
		}
		// chown -R mysql:mysql /var/lib/mysql.
		if err = os.Chown(dataPath, uid, gid); err != nil {
			return fmt.Errorf("failed to chown %s: %s", dataPath, err)
		}
	}

	// copy appropriate my.cnf from config-map to config mount.
	if err = copyFile(path.Join(mysqlCMPath, "my.cnf"), path.Join(mysqlConfigPath, "my.cnf")); err != nil {
		return fmt.Errorf("failed to copy my.cnf: %s", err)
	}

	buildDefaultXenonMeta(uid, gid)

	// build client.conf.
	clientConfig, err := cfg.buildClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build client.conf: %s", err)
	}
	// save client.conf to /etc/mysql.
	if err := clientConfig.SaveTo(path.Join(clientConfPath)); err != nil {
		return fmt.Errorf("failed to save client.conf: %s", err)
	}

	if err = os.Mkdir(extraConfPath, os.FileMode(0755)); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("error mkdir %s: %s", extraConfPath, err)
		}
	}

	// chown -R mysql:mysql /var/lib/mysql.
	if err = os.Chown(extraConfPath, uid, gid); err != nil {
		return fmt.Errorf("failed to chown %s: %s", dataPath, err)
	}

	// Run reset master in init-mysql container.
	if err = ioutil.WriteFile(initFilePath+"/reset.sql", []byte("reset master;"), 0644); err != nil {
		return fmt.Errorf("failed to write reset.sql: %s", err)
	}
	hasInitialized, _ := checkIfPathExists(path.Join(dataPath, "mysql"))
	// build init.sql.
	initSqlPath := path.Join(extraConfPath, "init.sql")

	// build extra.cnf.
	extraConfig, err := cfg.buildExtraConfig(initSqlPath)
	if err != nil {
		return fmt.Errorf("failed to build extra.cnf: %s", err)
	}

	// Notice: plugin.cnf cannot be copied to /etc/mysql/conf.d when initialized.
	// Check /var/lib/mysql/mysql exists. if exists it means that been initialized.
	if hasInitialized || strings.HasPrefix(getEnvValue("MYSQL_VERSION"), "5") {
		// Save plugin.cnf and extra.cnf to /etc/mysql/conf.d.
		saveCnfTo(extraConfPath, extraConfig)
	} else {
		log.Info("mysql is not initialized, use shell script copying plugin.cnf")
		// Save plugin.cnf and extra.cnf to /docker-entrypoint-initdb.d.
		saveCnfTo(initFilePath, extraConfig)

		src := PluginConfigsSh()
		// Write plugin.sh to docker-entrypoint-initdb.d/plugin.sh.
		// In this way, plugin.sh will be performed automatically when Percona docker-entrypoint.sh is executed.
		// plugin.sh will copy plugin.cnf and extra.cnf to /etc/mysql/conf.d.
		if err = ioutil.WriteFile(initFilePath+"/plugin.sh", []byte(src), 0755); err != nil {
			return fmt.Errorf("failed to write plugin.sh: %s", err)
		}
	}

	// // build leader-start.sh.
	// bashLeaderStart := cfg.buildLeaderStart()
	// leaderStartPath := path.Join(scriptsPath, "leader-start.sh")
	// if err = ioutil.WriteFile(leaderStartPath, bashLeaderStart, os.FileMode(0755)); err != nil {
	// 	return fmt.Errorf("failed to write leader-start.sh: %s", err)
	// }

	// // build leader-stop.sh.
	// bashLeaderStop := cfg.buildLeaderStop()
	// leaderStopPath := path.Join(scriptsPath, "leader-stop.sh")
	// if err = ioutil.WriteFile(leaderStopPath, bashLeaderStop, os.FileMode(0755)); err != nil {
	// 	return fmt.Errorf("failed to write leader-stop.sh: %s", err)
	// }

	// for install tokudb.
	if cfg.InitTokuDB {
		arg := fmt.Sprintf("echo never > %s/enabled", sysPath)
		cmd := exec.Command("sh", "-c", arg)
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to disable the transparent_hugepage: %s", err)
		}
	}

	// run the restore.
	// Check datadir is empty.
	// if /var/lib/mysql/mysql is empty, then run the restore.
	// otherwise , it must be has data, then do nothing.
	if !hasInitialized {
		if len(cfg.XRestoreFrom) != 0 {
			var err_f error
			if cfg.CloneFlag {
				err_f = cfg.executeCloneRestore()
				if err_f != nil {
					return fmt.Errorf("failed to execute Clone Restore : %s", err_f)
				}
			} else {
				if err_f = cfg.ExecuteNFSRestore(); err_f != nil {
					// No nfs , do s3 restore.
					if err_f = cfg.executeS3Restore(cfg.XRestoreFrom); err_f != nil {
						return fmt.Errorf("failed to restore from %s: %s", cfg.XRestoreFrom, err_f)
					}
				}
			}
			// Check has initialized again.
			hasInitialized, _ = checkIfPathExists(path.Join(dataPath, "mysql"))
		}
	}
	// Build init.sql after restore
	if err = ioutil.WriteFile(initSqlPath, cfg.buildInitSql(hasInitialized), 0644); err != nil {
		return fmt.Errorf("failed to write init.sql: %s", err)
	}
	// build xenon.json.
	xenonFilePath := path.Join(xenonPath, "xenon.json")
	if err = ioutil.WriteFile(xenonFilePath, cfg.buildXenonConf(), 0644); err != nil {
		return fmt.Errorf("failed to write xenon.json: %s", err)
	}

	log.Info("init command success")
	return nil
}

// start the backup http server.
func RunHttpServer(cfg *Config, stop <-chan struct{}) error {
	srv := newServer(cfg, stop)
	return srv.ListenAndServe()
}

// request a backup command.
func RunRequestBackup(cfg *Config, host string) error {
	_, err := requestABackup(cfg, host, serverBackupEndpoint)
	return err
}

// Save plugin.cnf and extra.cnf to specified path.
func saveCnfTo(targetPath string, extraCnf *ini.File) error {
	if err := copyFile(path.Join(mysqlCMPath, utils.PluginConfigs), path.Join(targetPath, utils.PluginConfigs)); err != nil {
		return fmt.Errorf("failed to copy plugin.cnf: %s", err)
	}
	if err := extraCnf.SaveTo(path.Join(targetPath, "extra.cnf")); err != nil {
		return fmt.Errorf("failed to save extra.cnf: %s", err)
	}
	return nil
}

func PluginConfigsSh() string {
	return fmt.Sprintf(`#!/bin/bash
cp %s %s
cp %s %s
chown -R mysql.mysql %s
chown -R mysql.mysql %s`,
		// cp plugin.cnf to /etc/mysql/conf.d/
		path.Join(initFilePath, utils.PluginConfigs), path.Join(extraConfPath, utils.PluginConfigs),
		// cp extra.cnf to /etc/mysql/conf.d/
		path.Join(initFilePath, "extra.cnf"), path.Join(extraConfPath, "extra.cnf"),
		// chown -R mysql.mysql plugin.cnf
		path.Join(extraConfPath, utils.PluginConfigs),
		// chown -R mysql.mysql extra.cnf
		path.Join(extraConfPath, "extra.cnf"))
}

func buildDefaultXenonMeta(uid, gid int) error {
	metaFile := fmt.Sprintf("%s/peers.json", xenonConfigPath)
	// mkdir var/lib/xenon.
	// https://github.com/radondb/xenon/blob/master/src/raft/raft.go#L118
	if err := os.MkdirAll(xenonConfigPath, 0777); err != nil {
		return fmt.Errorf("failed to mkdir %s: %s", xenonConfigPath, err)
	}
	// copy appropriate peers.json from config-map to config mount.
	if err := copyFile(path.Join(xenonCMPath, "peers.json"), path.Join(xenonConfigPath, "peers.json")); err != nil {
		return fmt.Errorf("failed to copy peers.json: %s", err)
	}
	// chown -R mysql:mysql /var/lib/xenon/peers.json.
	if err := os.Chown(metaFile, uid, gid); err != nil {
		return fmt.Errorf("failed to chown %s: %s", metaFile, err)
	}
	return nil
}
