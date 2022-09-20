/*
Copyright 2021 RadonDB.

Licensed under the Apache License, Version 2.0 (the "License")
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
	"bytes"
	"context"
	"io"
	"path/filepath"

	"errors"

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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewInitCommand return a pointer to cobra.Command.
func NewInitCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "do some initialization operations.",
		Run: func(cmd *cobra.Command, args []string) {
			var init bool
			var err error
			if init, err = runCloneAndInit(cfg); err != nil {
				log.Error(err, "clone error")
			}
			if err = runInitCommand(cfg, init); err != nil {
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
func runCloneAndInit(cfg *Config) (bool, error) {
	//check follower is exists?
	serviceURL := ""
	server := ""
	var hasInitialized = false
	var err error
	// Check the rebuildFrom exist?
	if serviceURL, server, err = getPod(cfg); err == nil {
		// rebuild, remove restoreFrom
		cfg.XRestoreFrom = ""
		log.Info("found the rebuild-from pod", "service", serviceURL)
	} else {
		log.Info("found the rebuild-from pod", "error", err.Error())
	}
	if len(serviceURL) == 0 && CheckServiceExist(cfg, "follower") {
		serviceURL = fmt.Sprintf("http://%s-%s:%v", cfg.ClusterName, "follower", utils.XBackupPort)
		server = fmt.Sprintf("%s-%s", cfg.ClusterName, "follower")
	}
	//check leader is exist?
	if len(serviceURL) == 0 && CheckServiceExist(cfg, "leader") {
		serviceURL = fmt.Sprintf("http://%s-%s:%v", cfg.ClusterName, "leader", utils.XBackupPort)
		server = fmt.Sprintf("%s-%s", cfg.ClusterName, "leader")
	}
	// Check has initialized. If so just return.
	hasInitialized, _ = checkIfPathExists(path.Join(dataPath, "mysql"))
	log.Info("mysqld is", "initialize", hasInitialized)
	if hasInitialized {
		if err := UpgradeShGen(cfg); err != nil {
			return hasInitialized, err
		}
	}
	if len(serviceURL) != 0 {
		// is MySQL8, CLone it
		if cfg.MySQLVersion.Major == 8 && len(cfg.XRestoreFrom) == 0 {
			// return clone int
			log.Info("server choose ", "server", server)
			err := DoCLone(cfg, server)
			return hasInitialized, err
		}

		if hasInitialized {
			log.Info("MySQL data directory existing!")
			return hasInitialized, nil
		}

		// backup at first
		Args := fmt.Sprintf("curl --user $BACKUP_USER:$BACKUP_PASSWORD %s/download|xbstream -x -C %s; exit ${PIPESTATUS[0]}",
			serviceURL, utils.DataVolumeMountPath)
		cmd := exec.Command("/bin/bash", "-c", "--", Args)
		log.Info("runCloneAndInit", "cmd", Args)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return hasInitialized, fmt.Errorf("failed to disable the run restore: %s", err)
		}
		cfg.XRestoreFrom = utils.DataVolumeMountPath // just for init clone
		cfg.CloneFlag = true
		return hasInitialized, nil
	}
	log.Info("no leader or follower found")
	return hasInitialized, nil
}

// runInitCommand do some initialization operations.
func runInitCommand(cfg *Config, hasInitialized bool) error {
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
	// check env exist
	if ib_pool := getEnvValue(utils.ROIbPool); len(ib_pool) != 0 {
		//replace sed
		arg := fmt.Sprintf("sed -i 's/^innodb_buffer_pool_size.*/innodb_buffer_pool_size\t=\t%s/' "+path.Join(mysqlConfigPath, "my.cnf"), ib_pool)
		log.Info("run sed", "arg", arg)
		cmd := exec.Command("sh", "-c", arg)
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to sed innodb_buffer_pool_size : %s", err)
		}
	}
	if ib_inst := getEnvValue(utils.ROIbInst); len(ib_inst) != 0 {
		//replace sed
		arg := fmt.Sprintf("sed -i 's/^innodb_buffer_pool_instances.*/innodb_buffer_pool_instances\t=\t%s/' "+path.Join(mysqlConfigPath, "my.cnf"), ib_inst)
		log.Info("run sed", "arg", arg)
		cmd := exec.Command("sh", "-c", arg)
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to sed innodb_buffer_pool_instances : %s", err)
		}
	}
	if ib_log := getEnvValue(utils.ROIbLog); len(ib_log) != 0 {
		//replace sed
		arg := fmt.Sprintf("sed -i 's/^innodb_log_file_size.*/innodb_log_file_size\t=\t%s/' "+path.Join(mysqlConfigPath, "my.cnf"), ib_log)
		log.Info("run sed", "arg", arg)
		cmd := exec.Command("sh", "-c", arg)
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to sed innodb_log_file_size : %s", err)
		}
	}
	// SSL settings.
	if exists, _ := checkIfPathExists(utils.TlsMountPath); exists {
		buildSSLdata()
	}
	// copy mysqlchecker to /opt/radondb/
	if exists, _ := checkIfPathExists(utils.RadonDBBinDir); exists {
		log.Info("copy mysqlchecker to /opt/radondb/")
		copyMySQLchecker()
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

// PITR backup
func (myConf *Config) RunPitrBackupS3(cfg *BackupClientConfig) {
	if len(cfg.XCloudS3AccessKey) == 0 || len(cfg.XCloudS3SecretKey) == 0 ||
		len(cfg.XCloudS3Bucket) == 0 || len(cfg.XCloudS3EndPoint) == 0 {
		log.Error(errors.New("s3: S3 not set"), "Do not set the S3 enviroment!")
		return
	} else {
		// Get Last Backup dir and Gitd
		lastbackup, lastgtid := getLastBackupInfo(myConf)
		// If do not have last backup, do nothing
		if len(lastbackup) != 0 && lastbackup != "null" {
			// Upload the binlog to S3
			log.Info("now begin to upload to S3")
			backupBinLogs(cfg, lastbackup, lastgtid)

		}
		//TODO : truncate by lastGtid
		log.Info("get last gtid", "lastgtid", lastgtid)

	}
}

// List all binlog and upload to S3
func backupBinLogs(cfg *BackupClientConfig, lastBackup, lastgtid string) error {
	// 1. flush all binlog
	arg := "mysql -uroot -h127.0.0.1 -e 'flush logs'"
	cmd := exec.Command("sh", "-c", arg)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to flush logs: %s", err.Error())
	}
	// 2. list all binlog
	root := utils.DataVolumeMountPath

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if err != nil {

			return fmt.Errorf("failed to walk dir: %s", err.Error())
		}

		if !info.IsDir() && //not dir
			strings.HasPrefix(filepath.Base(path), "mysql-bin.") && // mysql-bin.000xxx
			filepath.Base(path) != "mysql-bin.index" { // Don't upload the mysql-bin.index
			// files = append(files, path)
			log.Info("now upload file:", "file", path)
			// 3. upload to lastbackupDir
			if s3, err := NewS3(strings.TrimPrefix(strings.TrimPrefix(cfg.XCloudS3EndPoint, "https://"), "http://"),
				cfg.XCloudS3AccessKey, cfg.XCloudS3SecretKey, cfg.XCloudS3Bucket,
				strings.HasPrefix(cfg.XCloudS3EndPoint, "https")); err != nil {
				return fmt.Errorf("failed to new s3 : %s", err)
			} else {
				s3.S3Upload(cfg, buildBinlogDir(lastBackup), path)
			}

		}

		return nil
	})

	if err != nil {
		log.Error(err, "cannot find the files")
	}

	return nil
}

// Get Last Backup and Gtid
func getLastBackupInfo(cfg *Config) (string, string) {
	str := `curl -sX GET -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" -H "Content-Type: application/json"  --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/apis/mysql.radondb.com/v1alpha1/namespaces/%s/mysqlclusters/%s |jq '.status.%s'`
	arg := fmt.Sprintf(str, cfg.NameSpace, cfg.ClusterName, "lastbackup")
	cmd := exec.Command("sh", "-c", arg)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	log.Info("getLastBackupInfo", "args", arg)
	if err != nil {
		log.Error(err, "failed to create stdout pipe")

		return "", ""
	}
	if err := cmd.Start(); err != nil {
		return "", ""
	}
	defer func() {
		// don't care
		_ = stdout.Close()
	}()

	var buf bytes.Buffer
	io.Copy(&buf, stdout)
	lastbackup := strings.Trim(buf.String(), "\"\n")
	log.Info("getLastBackupInfo", "lastbackup", lastbackup)
	arg = fmt.Sprintf(str, cfg.NameSpace, cfg.ClusterName, "lastbackupGtid")

	cmd = exec.Command("sh", "-c", arg)
	cmd.Stderr = os.Stderr
	if stdout, err = cmd.StdoutPipe(); err != nil {
		log.Info("getLastBackupInfo", "StdoutPipe Error", err.Error())
		return lastbackup, ""
	}
	if err := cmd.Start(); err != nil {
		return lastbackup, ""
	}
	defer func() {
		// don't care
		_ = stdout.Close()
	}()

	var buf2 bytes.Buffer
	io.Copy(&buf2, stdout)
	lastbackupGtid := strings.Trim(buf2.String(), "\"\n")
	log.Info("getLastBackupInfo", "lastgtid", lastbackupGtid)
	return lastbackup, lastbackupGtid
}

// start the backup http server.
func RunHttpServer(cfg *Config, stop <-chan struct{}) error {
	//go RunPitr(cfg)
	srv := newServer(cfg, stop)
	return srv.ListenAndServe()
}

// request a backup command.
func RunRequestBackup(cfg *BackupClientConfig, host string) error {
	if cfg.BackupType == S3 {
		_, err := requestS3Backup(cfg, host, serverBackupEndpoint)
		return err
	}
	if cfg.BackupType == NFS {
		err := requestNFSBackup(cfg, host, serverBackupDownLoadEndpoint)
		return err
	}
	return fmt.Errorf("unknown backup type: %s", cfg.BackupType)
}

// Save plugin.cnf and extra.cnf to specified path.
func saveCnfTo(targetPath string, extraCnf *ini.File) error {
	userId := 1001
	groupId := 1001
	if err := copyFile(path.Join(mysqlCMPath, utils.PluginConfigs), path.Join(targetPath, utils.PluginConfigs)); err != nil {
		return fmt.Errorf("failed to copy plugin.cnf: %s", err)
	}
	if err := os.Chown(path.Join(targetPath, utils.PluginConfigs), userId, groupId); err != nil {
		return fmt.Errorf("failed to change owner of plugin.cnf: %s", err)
	}
	if err := extraCnf.SaveTo(path.Join(targetPath, "extra.cnf")); err != nil {
		return fmt.Errorf("failed to save extra.cnf: %s", err)
	}
	if err := os.Chown(path.Join(targetPath, "extra.cnf"), userId, groupId); err != nil {
		return fmt.Errorf("failed to change owner of plugin.cnf: %s", err)
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

func buildSSLdata() error {
	// cp -rp /tmp/myssl/* /etc/mysql/ssl/ Refer https://stackoverflow.com/questions/31467153/golang-failed-exec-command-that-works-in-terminal
	shellCmd := "cp  /tmp/mysql-ssl/* " + utils.TlsMountPath
	cmd := exec.Command("sh", "-c", shellCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy ssl: %s", err)
	}

	cronCmd := "chown -R mysql.mysql " + utils.TlsMountPath
	cmd = exec.Command("sh", "-c", cronCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy ssl: %s", err)
	}
	return nil
}

func copyMySQLchecker() error {
	cpCmd := "cp /mnt/mysqlchecker " + utils.RadonDBBinDir
	cmd := exec.Command("sh", "-c", cpCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy mysqlchecker: %s", err)
	}
	chownCmd := "chown -R mysql.mysql " + utils.RadonDBBinDir
	cmd = exec.Command("sh", "-c", chownCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to chown mysqlchecker: %s", err)
	}
	return nil
}

func getPod(cfg *Config) (string, string, error) {
	log.Info("Now check the pod which has got rebuild-from")
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", "", err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "", err
	}
	match := map[string]string{
		utils.LabelRebuildFrom: "true",
	}
	podList, err := clientset.CoreV1().Pods(cfg.NameSpace).
		List(context.TODO(), v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(match).String()})
	if err != nil {
		return "", "", err
	}
	if len(podList.Items) == 1 {
		pod := podList.Items[0]
		// Patch remove rebuild-from
		if err := removeRebuildFrom(clientset, cfg, pod.Name); err != nil {
			log.Info("remove rebuild from", "error", err.Error())
		}
		return fmt.Sprintf("%s.%s-mysql.%s:%v", pod.Name, cfg.ClusterName, cfg.NameSpace, utils.XBackupPort),
			fmt.Sprintf("%s.%s-mysql.%s", pod.Name, cfg.ClusterName, cfg.NameSpace), nil
	} else {
		return "", "", fmt.Errorf("not correct pod choose")
	}

}

func removeRebuildFrom(clientset *kubernetes.Clientset, cfg *Config, podName string) error {
	patch := fmt.Sprintf(`[{"op": "remove", "path": "/metadata/labels/%s"}]`, utils.LabelRebuildFrom)
	_, err := clientset.CoreV1().Pods(cfg.NameSpace).Patch(context.TODO(), podName, types.JSONPatchType, []byte(patch), v1.PatchOptions{})

	return err
}

func DoCLone(cfg *Config, server string) error {
	sql := fmt.Sprintf(`RESET SLAVE ALL;SET GLOBAL clone_valid_donor_list = '%s:3306' ;
	CLONE INSTANCE FROM '%s'@'%s':3306
	IDENTIFIED BY '%s';
`, server, cfg.DonorClone, server, cfg.DonorClonePassword)
	cloneSh := fmt.Sprintf(`#!/bin/bash
echo 'now is doing clone.'
mysqld &
mysql=( mysql -uroot -hlocalhost  --password="" )

for i in {120..0}; do
	if echo 'SELECT 1' | "${mysql[@]}" &> /dev/null; then
		break
	fi
	echo 'MySQL run process in progress...'
	sleep 1
done
mysql -uroot -hlocalhost  --password="" -e "%s"
echo "now delete socks file"
rm -rf /var/lib/mysql/*.sock
rm -rf /var/lib/mysql/mysql.sock.lock
`, sql)
	log.Info("write clone shell", "shell", cloneSh)
	//write to docker.
	if err := ioutil.WriteFile(initFilePath+"/clone.sh", []byte(cloneSh), 0755); err != nil {
		return fmt.Errorf("failed to write clone.sh: %s", err)
	}
	return nil
}

func UpgradeShGen(cfg *Config) error {
	if !cfg.NeedUpgrade {
		log.Info("do not upgrade mysqld")
		return nil
	}
	upgradesh := `#!/bin/bash
	echo 'now is doing upgrade'
	mysqld --upgrade=force &
	mysql=( mysql -uroot -hlocalhost  --password="" )
	
	for i in {120..0}; do
		if echo 'SELECT 1' | "${mysql[@]}" &> /dev/null; then
			break
		fi
		echo 'MySQL run process in progress...'
		sleep 1
	done
	mysql -uroot -hlocalhost  --password="" -e "shutdown"
	# waiting it stop
	for i in {120..0}; do
		if echo 'SELECT 1' | "${mysql[@]}" &> /dev/null; then
			echo 'MySQLD is shutdown...'
			sleep 1
		else
			echo 'MySQLD is shutted.'
			pkill mysql
			rm -rf /var/lib/mysql/*.sock
			rm -rf /var/lib/mysql/mysql.sock.lock
			break
		fi
	done
	`
	log.Info("write upgrade shell", "shell", upgradesh)
	//write to docker.
	if err := ioutil.WriteFile(initFilePath+"/upgrade.sh", []byte(upgradesh), 0755); err != nil {
		return fmt.Errorf("failed to write upgrade.sh: %s", err)
	}
	return nil
}
