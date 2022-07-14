package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	. "github.com/radondb/radondb-mysql-kubernetes/utils"
	log "github.com/sirupsen/logrus"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type config struct {
	container, namespace, podName, password string
	autoRebuild                             bool
}
type GTID struct {
	ID               string `json:"id"`
	Raft             string `json:"raft"`
	Mysql            string `json:"mysql"`
	ExecutedGTIDSet  string `json:"executed-gtid-set"`
	RetrievedGTIDSet string `json:"retrieved-gtid-set"`
}

type GTIDList struct {
	GTID []*GTID `json:"gtid"`
}

var (
	ns      string
	podName string
)

type KubeAPI struct {
	Client *kubernetes.Clientset
	Config *rest.Config
}
type MySQLServer struct {
	Conn *sqlx.DB
	DSN  string
}

const (
	leaderStopCommand    = "kill -9 $(pidof mysqld)"
	containerNameDefault = "mysql"
	xenonConf            = "/etc/xenon/xenon.json"
	mysqlUser            = "root"
	mysqlHost            = "127.0.0.1"
	raftDisableCommand   = "xenoncli raft disable"
	raftEnableCommand    = "xenoncli raft enable"
	raftStatusCommand    = "xenoncli raft status"
	mysqlGtidCommand     = "xenoncli cluster gtid json"
)

func init() {
	ns = os.Getenv("NAMESPACE")
	podName = os.Getenv("POD_NAME")
	debugFlag, _ := strconv.ParseBool(os.Getenv("RADONDB_DEBUG"))
	if debugFlag {
		log.SetLevel(log.DebugLevel)
	}
	log.Infof("debug flag set to %t", debugFlag)
}

func main() {
	switch os.Args[1] {
	case "leaderStart":
		if err := leaderStart(); err != nil {
			log.Fatalf("leaderStart failed: %v", err)
		}
	case "leaderStop":
		if err := leaderStop(); err != nil {
			log.Fatalf("leaderStop failed: %v", err)
		}
	case "liveness":
		if err := liveness(); err != nil {
			log.Fatalf("liveness failed: %v", err)
		}
	case "readiness":
		if err := readiness(); err != nil {
			log.Fatalf("readiness failed: %v", err)
		}
	case "postStart":
		if err := postStart(); err != nil {
			log.Fatalf("postStart failed: %v", err)
		}
	case "preStop":
		if err := preStop(); err != nil {
			log.Fatalf("postStop failed: %v", err)
		}
	default:
		log.Fatalf("Usage: %s leaderStart|leaderStop|liveness|readiness|postStart|preStop", os.Args[0])
	}
}

// TODO
func leaderStart() error {
	return nil
}

func leaderStop() error {
	log.Infof("leader stop started")
	// first load any configuration provided (e.g. via environment variables)
	cfg := loadConfiguration()
	dsn := mysqlUser + ":" + cfg.password + "@tcp(" + mysqlHost + ")/"
	server := MySQLServer{DSN: dsn}
	Conn, err := server.GetNewDBConn()
	if err != nil {
		log.Error(err)
	}
	IsLeader := IsLeader(Conn)
	if !IsLeader {
		log.Info("I am not the leader")
		os.Exit(0)
	}
	// Step 1: disable event scheduler
	log.Info("Disabling event scheduler")
	stmt, err := SetEventScheduler(Conn, false)
	if err != nil {
		log.Error(err)
	}
	log.Infof("set event scheduler to false: %s", stmt)
	// Step 2: set readonly
	log.Info("Setting readonly")
	stmt, err = SetReadOnly(Conn, true)
	if err != nil {
		log.Error(err)
	}
	log.Infof("set readonly to true: %s", stmt)
	// Step 3: TODO: check long running writes
	log.Info("Checking long running writes")
	num, stmt, err := CheckLongRunningWrites(Conn, 4)
	if err != nil {
		log.Error(err)
	}
	log.Infof("%d,long running writes: %s", num, stmt)
	// Step 4: set max connections
	// log.Info("Setting max connections")
	// stmt, err = SetMaxConnections(Conn, "1")
	// if err != nil {
	// 	log.Error(err)
	// }
	// log.Info("set max connections to 1: ", stmt)
	// Step 5: kill threads
	log.Info("Killing threads")
	stmt, err = KillThreads(Conn)
	if err != nil {
		log.Error(err)
	}
	log.Info("killed threads: ", stmt)
	// Step 6: FlushTablesWithReadLock
	log.Info("Flushing tables with read lock")
	stmt, err = FlushTablesWithReadLock(Conn)
	if err != nil {
		log.Error(err)
	}
	log.Info("flushed tables with read lock: ", stmt)
	// Step 7: FlushBinaryLogs
	log.Info("Flushing binary logs")
	stmt, err = FlushBinaryLogs(Conn)
	if err != nil {
		log.Error(err)
	}
	log.Info("flushed binary logs:", stmt)
	time.Sleep(2 * time.Second)
	// cmd := createLeaderStopCommand(cfg)
	// log.Infof("command to execute is [%s]", strings.Join(cmd, " "))
	// Step 8: do a clean failover
	log.Info("Doing a clean failover")
	DoFailOver()
	// var output, stderr string
	// output, stderr, err = runCommand(k, cfg, cmd)
	// // log any output and check for errors
	// log.Info("output=[" + output + "]")
	// log.Info("stderr=[" + stderr + "]")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	log.Info("leader stop ends")
	return nil
}

func liveness() error {
	return XenonPingMyself()
}

func readiness() error {
	role := GetRole()
	if role != string(Leader) {
		return PatchRoleLabelTo(myself(role))
	}
	return nil
}

// Step 1: waitting for  mysql is in a healthy state
// Step 2: check if auto rebuild is enabled
// Step 3: check if GTID has subtract
// Step 4: rebuild if necessary
func postStart() error {
	log.Info("postStart starts")
	cfg := loadConfiguration()
	dsn := mysqlUser + ":" + cfg.password + "@tcp(" + mysqlHost + ")/"
	server := MySQLServer{DSN: dsn}
	if helth, err := server.WaitforMySQL(time.Second * 10); err != nil {
		log.Error(err)
	} else {
		if !helth {
			log.Error("mysql is not in a healthy state")
			return err
		}
	}
	if cfg.autoRebuild {
		log.Info("auto rebuild is enabled")
		var mySet, leaderSet string
		gtidList := GetGTIDList()
		for _, gtid := range gtidList {
			if strings.Split(gtid.ID, ".")[0] == podName {
				if gtid.Raft == string(Leader) {
					log.Infof("I am the leader, skip GTID_SUBSET")
					return nil
				}
				mySet = gtid.ExecutedGTIDSet
			}
			if gtid.Raft == string(Leader) {
				leaderSet = gtid.ExecutedGTIDSet
			}
		}
		db, err := server.GetNewDBConn()
		if err != nil {
			log.Error(err)
		}
		gtidSubSet, query, err := HaveErrantTransactions(db, leaderSet, mySet)
		if err != nil {
			log.Errorf("can not get errant transactions: %s, %s", err, query)
			return err
		}
		if !gtidSubSet {
			// rebuild
			log.Infof("mySet is not a subset of leaderSet, rebuild me, mySet: %s, leaderSet: %s", mySet, leaderSet)
			if err := PatchRebuildLabelTo(myself("")); err != nil {
				log.Error(err)
				return err
			}
			return nil
		}
		log.Infof("need not rebuid, mySet: %s, leaderSet: %s", mySet, leaderSet)
		return nil

	} else {
		log.Info("auto rebuild is disabled")
	}
	return nil
}

// TODO
func preStop() error {
	if err := leaderStop(); err != nil {
		return err
	}
	return nil
}

func myself(role string) MySQLNode {
	return MySQLNode{
		PodName:   podName,
		Namespace: ns,
		Role:      role,
	}
}
func (k *KubeAPI) Exec(namespace, pod, container string, stdin io.Reader, command []string) (string, string, error) {
	var stdout, stderr bytes.Buffer

	var Scheme = runtime.NewScheme()
	if err := core_v1.AddToScheme(Scheme); err != nil {
		log.Error(err)
		return "", "", err
	}
	var ParameterCodec = runtime.NewParameterCodec(Scheme)

	request := k.Client.CoreV1().RESTClient().Post().
		Resource("pods").SubResource("exec").
		Namespace(namespace).Name(pod).
		VersionedParams(&core_v1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     stdin != nil,
			Stdout:    true,
			Stderr:    true,
		}, ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.Config, "POST", request.URL())

	if err == nil {
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  stdin,
			Stdout: &stdout,
			Stderr: &stderr,
		})
	}

	return stdout.String(), stderr.String(), err
}

func NewConfig() (*rest.Config, error) {
	// The default loading rules try to read from the files specified in the
	// environment or from the home directory.
	loader := clientcmd.NewDefaultClientConfigLoadingRules()

	// The deferred loader tries an in-cluster config if the default loading
	// rules produce no results.
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader, &clientcmd.ConfigOverrides{}).ClientConfig()
}

func NewForConfig(config *rest.Config) (*KubeAPI, error) {
	var api KubeAPI
	var err error

	api.Config = config
	api.Client, err = kubernetes.NewForConfig(api.Config)

	return &api, err
}

// getEnvRequired attempts to get an environmental variable that is required
// by this program. If this cannot happen, we fatally exit
func getEnvRequired(envVar string) string {
	val := strings.TrimSpace(os.Getenv(envVar))

	if val == "" {
		log.Fatalf("required environmental variable %q not set, exiting.", envVar)
	}

	log.Debugf("%s set to: %s", envVar, val)

	return val
}

// loadConfiguration loads configuration from the environment as needed to run a pgBackRest
// command
func loadConfiguration() config {
	cfg := config{}
	flag := os.Getenv("AUTO_REBUILD")
	if flag == "" {
		cfg.autoRebuild = false
	} else {
		autoRebuild, err := strconv.ParseBool(flag)
		cfg.autoRebuild = autoRebuild
		if err != nil {
			log.Error(err)
		}
	}
	cfg.namespace = getEnvRequired("NAMESPACE")
	cfg.container = os.Getenv("CONTAINER")
	if cfg.container == "" {
		cfg.container = containerNameDefault
	}
	log.Debugf("CONTAINER set to: %s", cfg.container)
	cfg.podName = os.Getenv("PODNAME")
	if cfg.podName == "" {
		cfg.podName, _ = os.Hostname()
	}
	log.Debugf("PODNAME set to: %s", cfg.podName)
	xenonConfFile, err := os.Open(xenonConf)
	if err != nil {
		log.Fatal(err)
	}
	contents, err2 := ioutil.ReadAll(xenonConfFile)
	if err2 != nil {
		log.Fatal(err2)
	}
	defer xenonConfFile.Close()
	js, js_err := simplejson.NewJson(contents)
	if js_err != nil {
		log.Fatal(js_err)
	}
	cfg.password, _ = js.Get("mysql").Get("passwd").String()
	return cfg

}

// createPGBackRestCommand form the proper pgBackRest command based on the configuration provided
// func createLeaderStopCommand() []string {
// 	cmd := []string{leaderStopCommand}
// 	return cmd
// }

func runCommandLocal(cmd []string) (bytes.Buffer, string, error) {
	var stdout, stderr bytes.Buffer
	var err error
	log.Debugf("command to execute is [%s]", strings.Join(cmd, " "))
	cmd_ := exec.Command(cmd[0], cmd[1:]...)
	cmd_.Stdout = &stdout
	cmd_.Stderr = &stderr
	err = cmd_.Run()
	return stdout, stderr.String(), err
}

// func runCommand(kubeapi *KubeAPI, cfg config, cmd []string) (string, string, error) {
// 	bashCmd := []string{"bash"}
// 	reader := strings.NewReader(strings.Join(cmd, " "))
// 	return kubeapi.Exec(cfg.namespace, cfg.podName, cfg.container, reader, bashCmd)
// }
func (server *MySQLServer) GetNewDBConn() (*sqlx.DB, error) {
	conn, err := sqlx.Connect("mysql", server.DSN)
	if err != nil {
		log.Errorf("Error connecting to MySQL: %v", err)
		return nil, err
	}
	return conn, nil
}
func SetEventScheduler(db *sqlx.DB, state bool) (string, error) {
	var err error
	stmt := ""
	if state {
		stmt = "SET GLOBAL event_scheduler=1"
	} else {
		stmt = "SET GLOBAL event_scheduler=0"
	}
	_, err = db.Exec(stmt)
	return stmt, err
}
func SetReadOnly(db *sqlx.DB, state bool) (string, error) {
	var err error
	stmt := ""
	if state {
		stmt = "SET GLOBAL read_only=1"
	} else {
		stmt = "SET GLOBAL read_only=0"
	}
	_, err = db.Exec(stmt)
	return stmt, err
}
func CheckLongRunningWrites(db *sqlx.DB, thresh int) (int, string, error) {
	var count int
	query := "select SUM(ct) from ( select count(*) as ct from information_schema.processlist  where command = 'Query' and time >= ? and info not like 'select%' union all select count(*) as ct  FROM  INFORMATION_SCHEMA.INNODB_TRX trx WHERE trx.trx_started < CURRENT_TIMESTAMP - INTERVAL ? SECOND) A"
	err := db.QueryRowx(query, thresh, thresh).Scan(&count)
	return count, query + "(" + strconv.Itoa(thresh) + ")", err
}
func KillThreads(db *sqlx.DB) (string, error) {
	var ids []int
	query := "SELECT Id FROM information_schema.PROCESSLIST WHERE Command != 'Binlog Dump GTID' AND User not in ('root','radondb_repl') AND Id != CONNECTION_ID()"
	logs := query
	err := db.Select(&ids, query)
	if err != nil {
		return logs, err
	}
	for _, id := range ids {
		log, err := KillThread(db, strconv.Itoa(id))
		logs += log
		//Should we exit in case of error ?
		if err != nil {
			return logs, err
		}
	}
	return logs, err

}

// TODO: This is a hack to kill threads. We need to find a better way to do this
func KillThread(db *sqlx.DB, id string) (string, error) {
	_, err := db.Exec("KILL ?", id)
	return "KILL ? (" + id + ")", err

}

func SetMaxConnections(db *sqlx.DB, connections string) (string, error) {

	query := "SET GLOBAL max_connections=" + connections
	_, err := db.Exec(query)
	return query, err
}
func FlushTablesWithReadLock(db *sqlx.DB) (string, error) {
	query := "FLUSH NO_WRITE_TO_BINLOG TABLES WITH READ LOCK"
	_, err := db.Exec(query)
	return query, err
}
func FlushBinaryLogs(db *sqlx.DB) (string, error) {
	_, err := db.Exec("FLUSH  BINARY LOGS")
	return "FLUSH BINARY LOGS", err
}
func IsLeader(db *sqlx.DB) bool {
	var n int
	var isLeader bool
	err := db.QueryRowx("SELECT @@read_only").Scan(&n)
	if err != nil {
		return false
	}
	if n == 0 {
		isLeader = true
	} else {
		isLeader = false
	}
	return isLeader
}
func HaveErrantTransactions(db *sqlx.DB, gtidMaster string, gtidSlave string) (bool, string, error) {

	count := 0
	query := "select gtid_subset('" + gtidMaster + "','" + gtidSlave + "') as slave_is_subset"

	err := db.QueryRowx(query).Scan(&count)
	if err != nil {
		return false, query, err
	}

	if count == 0 {
		return true, query, nil
	}
	return false, query, nil
}
func GetGTID(db *sqlx.DB) (string, error) {
	var gtid string
	err := db.QueryRowx("SELECT @@gtid_executed").Scan(&gtid)
	return gtid, err
}

// func GetPreferredMaster(db *sqlx.DB) (string, error) {
// 	var out map[string]interface{}
// 	raftStatus, stdErr, err := runCommandLocal([]string{raftStatusCommand})
// 	if err != nil {
// 		panic(err)
// 	}
// 	if stdErr != "" {
// 		panic(stdErr)
// 	}
// 	if err = json.Unmarshal(raftStatus.Bytes(), &out); err != nil {
// 		return "", err
// 	}
// 	nodesJson := out["nodes"].([]interface{})
// 	nodes := []string{}
// 	for _, node := range nodesJson {
// 		nodeStr := strings.Split(node.(string), ":")
// 		if os.Hostname() != nodeStr[0] {
// 		nodes = append(nodes, nodeStr[0])
// 	}

// }
func DoFailOver() {
	raftEnableCommand := []string{"bash", "-c", raftEnableCommand}
	defer func() {
		if _, stdErr, err := runCommandLocal(raftEnableCommand); err != nil {
			log.Errorf("Error deferr enableing  raft: %v", stdErr)
		}
	}()
	raftDisableCommand := []string{"bash", "-c", raftDisableCommand}
	if _, stdErr, err := runCommandLocal(raftDisableCommand); err != nil {
		log.Errorf("Error disabling raft: %v", stdErr)
	}
	WaitForNewLeader(time.Second * 200)

}

func WaitForNewLeader(timeout time.Duration) {
	raftStatusCommand := []string{"bash", "-c", raftStatusCommand}
	count := 0

	for {
		raftStatus, stdErr, err := runCommandLocal(raftStatusCommand)
		if err != nil {
			log.Errorf("Error getting raft status: %v", stdErr)
			break
		}
		var out map[string]interface{}
		if err := json.Unmarshal(raftStatus.Bytes(), &out); err != nil {
			log.Errorf("Error unmarshalling raft status: %v", err)
		}
		if out["leader"] != "" && out["leader"] != nil {
			log.Debugf("I am not the leader")
			break

		} else {
			log.Debug("I am the leader: ", out["leader"])
		}
		time.Sleep(time.Second * 2)
		count += 1
		if count*2 > int(timeout) {
			log.Errorf("Timeout waiting for im not leader")
			break
		}

	}
}

func (server *MySQLServer) WaitforMySQL(timeout time.Duration) (bool, error) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	timeoutExceeded := time.After(timeout)
	for {
		select {
		case <-timeoutExceeded:
			return false, fmt.Errorf("db connection failed after %s timeout", timeout)
		case <-ticker.C:
			_, err := sqlx.Connect("mysql", server.DSN)
			if err == nil {
				return true, nil
			}
			log.Errorf("failed to connect to db %s", server.DSN)
		}
	}
}
func GetGTIDList() []*GTID {
	cmd := []string{"bash", "-c", mysqlGtidCommand}
	res, stderr, err := runCommandLocal(cmd)
	if err != nil {
		log.Fatalf("failed to exec xenoncli cluster gtid json: %s", stderr)
	}
	gtidList := GTIDList{GTID: []*GTID{}}
	if err := json.Unmarshal(res.Bytes(), &gtidList); err != nil {
		log.Fatalf("failed to unmarshal gtid: %v", err)
	}
	return gtidList.GTID
}

func PatchRebuildLabelTo(n MySQLNode) error {
	patch := `{"metadata":{"labels":{"rebuild":"true"}}}`
	err := patchPodLabel(n, patch)
	if err != nil {
		return fmt.Errorf("failed to patch pod rebuild label: %v", err)
	}
	return nil
}
func patchPodLabel(n MySQLNode, patch string) error {
	clientset, err := GetClientSet()
	if err != nil {
		return fmt.Errorf("failed to create clientset: %v", err)
	}
	_, err = clientset.CoreV1().Pods(n.Namespace).Patch(context.TODO(), n.PodName, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}
