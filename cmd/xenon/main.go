package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/radondb/radondb-mysql-kubernetes/utils"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type KubeAPI struct {
	Client *kubernetes.Clientset
	Config *rest.Config
}

type runRemoteCommandConfig struct {
	container, namespace, podName string
}

const (
	leaderStopCommand  = "kill -9 $(pidof mysqld)"
	mysqlUser          = "root"
	mysqlHost          = "127.0.0.1"
	mysqlPwd           = ""
	raftDisableCommand = "xenoncli raft disable"
	raftEnableCommand  = "xenoncli raft enable"
	raftStatusCommand  = "xenoncli raft status"
	mysqlGtidCommand   = "xenoncli cluster gtid json"
)

var (
	ns          string
	podName     string
	autoRebuild string
)

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

type MySQLServer struct {
	DB *sql.DB
}

func init() {
	ns = os.Getenv("NAMESPACE")
	podName = os.Getenv("POD_NAME")
	autoRebuild = os.Getenv("AUTO_REBUILD")
	debugFlag, _ := strconv.ParseBool(os.Getenv("RADONDB_DEBUG"))
	if debugFlag {
		log.SetLevel(log.DebugLevel)
	}
	log.Infof("debug flag set to %t", debugFlag)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s leaderStart|leaderStop|liveness|readiness|postStart|preStop", os.Args[0])
	}
	switch os.Args[1] {
	case "leaderStart":
		if err := leaderStart(); err != nil {
			log.Fatalf("leaderStart failed: %s", err.Error())
		}
	case "leaderStop":
		if err := leaderStop(); err != nil {
			log.Fatalf("leaderStop failed: %s", err.Error())
		}
	case "liveness":
		if err := liveness(); err != nil {
			log.Fatalf("liveness failed: %s", err.Error())
		}
	case "readiness":
		if err := readiness(); err != nil {
			log.Fatalf("readiness failed: %s", err.Error())
		}
	case "postStart":
		if err := postStart(); err != nil {
			log.Fatalf("postStart failed: %s", err.Error())
		}
	case "preStop":
		if err := preStop(); err != nil {
			log.Fatalf("postStop failed: %s", err.Error())
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
	conn, err := getLocalMySQLConn()
	if err != nil {
		return fmt.Errorf("failed to get the connection of local MySQL: %s", err.Error())
	}
	defer conn.Close()

	if isReadonly(conn) {
		log.Info("I am readonly, skip the leader stop")
		os.Exit(0)
	}
	ch := make(chan error)
	go func() {
		defer func() {
			if err := enableMyRaft(); err != nil {
				log.Error(err)
			}
		}()

		log.Info("Raft disable")
		if err := disableMyRaft(); err != nil {
			log.Errorf("failed to failover: %v", err)
			ch <- err
		}

		// Step 1: disable event scheduler
		log.Info("Disabling event scheduler")
		stmt, err := SetEventScheduler(conn, false)
		if err != nil {
			ch <- err
		}
		log.Infof("set event scheduler to false: %s", stmt)

		// Step 2: set readonly
		log.Info("Setting readonly")
		stmt, err = SetReadOnly(conn, true)
		if err != nil {
			ch <- err
		}
		log.Infof("set readonly to true: %s", stmt)

		// Step 3: check long running writes
		log.Info("Checking long running writes")
		num, stmt, err := CheckLongRunningWrites(conn, 4)
		if err != nil {
			ch <- err
		}
		log.Infof("%d,long running writes: %s", num, stmt)

		// TODO: Step 4: set max connections

		// Step 5: kill threads
		log.Info("Killing threads")
		err = KillThreads(conn)
		if err != nil {
			ch <- err
		}

		// Step 6: FlushTablesWithReadLock
		log.Info("Flushing tables with read lock")
		stmt, err = FlushTablesWithReadLock(conn)
		if err != nil {
			ch <- err
		}
		log.Info("flushed tables with read lock: ", stmt)

		// Step 7: FlushBinaryLogs
		log.Info("Flushing binary logs")
		stmt, err = FlushBinaryLogs(conn)
		if err != nil {
			ch <- err
		}
		log.Info("flushed binary logs:", stmt)
	}()
	select {
	case err := <-ch:
		return err
	case <-time.After(5 * time.Second):
		log.Info("timeout")
		if err := killMysqld(); err != nil {
			return err
		}
		return nil
	}

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

// Step 1: check if auto rebuild is enabled
// Step 2: waiting for mysql connection available
// Step 3: waiting for xenon available
// Step 4: check whether the gtd set of the current node is the subset of the Leader`s
// Step 5: rebuild if necessary
func postStart() error {
	log.Info("postStart starts")
	// Step 1: check if auto rebuild is enabled
	if !enableAutoRebuild() {
		log.Info("auto rebuild is disabled, skip the post start")
		return nil
	}
	log.Info("auto rebuild is enabled")

	// Step 2: waiting for mysql connection available
	if err := WaitMySQLAvailable(time.Second * 30); err != nil {
		return fmt.Errorf("failed to wait for mysql: %s", err.Error())
	}
	// Step 3: waiting for xenon available
	if err := WaitXenonAvailable(time.Minute); err != nil {
		return fmt.Errorf("xenon is not available: %s", err.Error())
	}
	// Get the GTID of the current server and leader server
	var mySet, leaderSet string
	aliveNode := 0
	gtidList := GetGTIDList()
	for _, gtid := range gtidList {
		if gtid.Raft == string(Unknown) || gtid.Raft == "" {
			continue
		}
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
		aliveNode++
	}
	// If there are only two available nodes, rebuild is unsafe
	if aliveNode < 3 || leaderSet == "" {
		log.Infof("cluster is not available, alive node: %d, leader`s gtid set: %s", aliveNode, leaderSet)
		return nil
	}

	db, err := getLocalMySQLConn()
	if err != nil {
		return fmt.Errorf("failed to get the connection of local MySQL: %s", err.Error())
	}
	defer db.Close()

	// Step 4: check whether the gtd set of the current node is the subset of the Leader`s
	isSubset, err := HaveErrantTransactions(db, leaderSet, mySet)
	if err != nil {
		return err
	}
	if !isSubset {
		// Step 5: Rebuild me
		log.Infof("mySet is not a subset of leaderSet, rebuild me, mySet: %s, leaderSet: %s", mySet, leaderSet)
		if err := PatchRebuildLabelTo(myself("")); err != nil {
			return err
		}
		return nil
	}
	log.Infof("postStart is finished, the current node does not need to be rebuilt\nmySet: \n%s,\nleaderSet: \n%s", formatGTIDSet(mySet), formatGTIDSet(leaderSet))
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

func enableAutoRebuild() bool {
	return autoRebuild == "true"
}

func runCommandLocal(cmd []string) (bytes.Buffer, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var stdout, stderr bytes.Buffer
	var err error
	log.Debugf("command to execute is [%s]", strings.Join(cmd, " "))
	cmd_ := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	cmd_.Stdout = &stdout
	cmd_.Stderr = &stderr
	err = cmd_.Run()
	return stdout, stderr.String(), err
}

func localDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=5s&multiStatements=true&interpolateParams=true",
		mysqlUser, mysqlPwd, mysqlHost, MysqlPort)
}

func getLocalMySQLConn() (*sql.DB, error) {
	db, err := sql.Open("mysql", localDSN())
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func SetEventScheduler(db *sql.DB, state bool) (string, error) {
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

func SetReadOnly(db *sql.DB, state bool) (string, error) {
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

func CheckLongRunningWrites(db *sql.DB, thresh int) (int, string, error) {
	var count int
	query := "select SUM(ct) from ( select count(*) as ct from information_schema.processlist  where command = 'Query' and time >= ? and info not like 'select%' union all select count(*) as ct  FROM  INFORMATION_SCHEMA.INNODB_TRX trx WHERE trx.trx_started < CURRENT_TIMESTAMP - INTERVAL ? SECOND) A"
	err := db.QueryRow(query, thresh, thresh).Scan(&count)
	return count, query + "(" + strconv.Itoa(thresh) + ")", err
}

func KillThreads(db *sql.DB) error {
	query := "SELECT Id FROM information_schema.PROCESSLIST WHERE Command != 'Binlog Dump GTID' AND User not in ('root','radondb_repl') AND Id != CONNECTION_ID()"
	logs := query
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to scan rows: %s", err.Error())
		}
		log, err := KillThread(db, strconv.Itoa(id))
		logs += log
		if err != nil {
			return err
		}
	}
	return err
}

// TODO: This is a hack to kill threads. We need to find a better way to do this
func KillThread(db *sql.DB, id string) (string, error) {
	_, err := db.Exec("KILL ?", id)
	return "KILL ? (" + id + ")", err
}

// func SetMaxConnections(db *sql.DB, connections string) (string, error) {
// 	query := "SET GLOBAL max_connections=" + connections
// 	_, err := db.Exec(query)
// 	return query, err
// }

func FlushTablesWithReadLock(db *sql.DB) (string, error) {
	query := "FLUSH NO_WRITE_TO_BINLOG TABLES WITH READ LOCK"
	_, err := db.Exec(query)
	return query, err
}

func FlushBinaryLogs(db *sql.DB) (string, error) {
	_, err := db.Exec("FLUSH  BINARY LOGS")
	return "FLUSH BINARY LOGS", err
}

func isReadonly(db *sql.DB) bool {
	var readOnly int
	err := db.QueryRow("SELECT @@read_only").Scan(&readOnly)
	if err != nil {
		return false
	}
	return readOnly == 1
}

// Returns true if all GTIDs in mySet are also in leaderSet. Returns false otherwise.
// https://dev.mysql.com/doc/refman/5.7/en/gtid-functions.html#function_gtid-subset
func HaveErrantTransactions(db *sql.DB, leaderSet, mySet string) (bool, error) {
	var isSubset int
	query := "select gtid_subset('" + mySet + "','" + leaderSet + "') as slave_is_subset"

	err := db.QueryRow(query).Scan(&isSubset)
	if err != nil {
		return false, fmt.Errorf("failed to execute: %s, err: %s", query, err.Error())
	}
	return isSubset == 1, nil
}

func GetGTID(db *sql.DB) (string, error) {
	var gtid string
	err := db.QueryRow("SELECT @@gtid_executed").Scan(&gtid)
	return gtid, err
}

func DoFailOver() error {
	defer func() {
		if err := enableMyRaft(); err != nil {
			log.Error(err)
		}
	}()
	if err := disableMyRaft(); err != nil {
		return err
	}
	return nil
}

func enableMyRaft() error {
	raftEnableCommand := []string{"bash", "-c", raftEnableCommand}
	if _, stdErr, err := runCommandLocal(raftEnableCommand); err != nil {
		return fmt.Errorf("failed to enable my raft: %s", stdErr)
	}
	return nil
}

func disableMyRaft() error {
	raftDisableCommand := []string{"bash", "-c", raftDisableCommand}
	if _, stdErr, err := runCommandLocal(raftDisableCommand); err != nil {
		log.Errorf("failed to disable my raft: %s", stdErr)
		return err
	}
	return nil
}

func WaitXenonAvailable(timeout time.Duration) error {
	err := wait.PollImmediate(2*time.Second, timeout, func() (bool, error) {
		if err := XenonPingMyself(); err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("failed to start xenon: %s", err.Error())
	}
	return nil
}

func WaitForNewLeader(timeout time.Duration) error {
	err := wait.PollImmediate(2*time.Second, timeout, func() (bool, error) {
		leader := getLeader()
		if leader != "" {
			log.Infof("new leader: %s", leader)
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for new leader: %s", err.Error())
	}
	return nil
}

func WaitMySQLAvailable(timeout time.Duration) error {
	err := wait.PollImmediate(2*time.Second, timeout, func() (done bool, err error) {
		conn, err := getLocalMySQLConn()
		if err != nil {
			return false, nil
		}
		defer conn.Close()
		return true, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func GetGTIDList() []*GTID {
	cmd := []string{"bash", "-c", mysqlGtidCommand}
	res, stderr, err := runCommandLocal(cmd)
	if err != nil {
		log.Fatalf("failed to exec xenoncli cluster gtid json: %s", stderr)
	}
	gtidList := GTIDList{GTID: []*GTID{}}
	if err := json.Unmarshal(res.Bytes(), &gtidList); err != nil {
		log.Fatalf("failed to unmarshal gtid: %s", err.Error())
	}
	return gtidList.GTID
}

func formatGTIDSet(set string) string {
	return strings.ReplaceAll(set, ",", "\n")
}

func getLeader() string {
	gtid := GetGTIDList()
	for _, g := range gtid {
		if g.Raft == string(Leader) {
			return g.ID
		}
	}
	return ""
}

func PatchRebuildLabelTo(n MySQLNode) error {
	patch := `{"metadata":{"labels":{"rebuild":"true"}}}`
	err := patchPodLabel(n, patch)
	if err != nil {
		return fmt.Errorf("failed to patch pod rebuild label: %s", err.Error())
	}
	return nil
}

func patchPodLabel(n MySQLNode, patch string) error {
	clientset, err := GetClientSet()
	if err != nil {
		return fmt.Errorf("failed to create clientset: %s", err.Error())
	}
	_, err = clientset.CoreV1().Pods(n.Namespace).Patch(context.TODO(), n.PodName, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (k *KubeAPI) Exec(namespace, pod, container string, stdin io.Reader, command []string) (string, string, error) {
	var stdout, stderr bytes.Buffer

	var Scheme = runtime.NewScheme()
	if err := corev1.AddToScheme(Scheme); err != nil {
		log.Fatalf("failed to add to scheme: %v", err)
		return "", "", err
	}
	var ParameterCodec = runtime.NewParameterCodec(Scheme)

	request := k.Client.CoreV1().RESTClient().Post().
		Resource("pods").SubResource("exec").
		Namespace(namespace).Name(pod).
		VersionedParams(&corev1.PodExecOptions{
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

func runRemoteCommand(kubeapi *KubeAPI, cfg runRemoteCommandConfig, cmd []string) (string, string, error) {
	bashCmd := []string{"bash"}
	reader := strings.NewReader(strings.Join(cmd, " "))
	return kubeapi.Exec(cfg.namespace, cfg.podName, cfg.container, reader, bashCmd)
}

func NewForConfig(config *rest.Config) (*KubeAPI, error) {
	var api KubeAPI
	var err error

	api.Config = config
	api.Client, err = kubernetes.NewForConfig(api.Config)

	return &api, err
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

func killMysqld() error {
	config, err := NewConfig()
	if err != nil {
		panic(err)
	}
	k, err := NewForConfig(config)
	if err != nil {
		panic(err)
	}
	cfg := runRemoteCommandConfig{
		podName:   podName,
		namespace: ns,
		container: "mysql",
	}

	killMySQLCommand := []string{leaderStopCommand}
	log.Infof("killing mysql command: %s", leaderStopCommand)
	var output, stderr string
	output, stderr, err = runRemoteCommand(k, cfg, killMySQLCommand)
	log.Info("output=[" + output + "]")
	log.Info("stderr=[" + stderr + "]")
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
