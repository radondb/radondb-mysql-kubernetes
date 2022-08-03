package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

type MySQLChecker interface {
	CheckMySQLHealthy(role string) bool
	CheckMySQLLiveness() error
	Client
}

// mysqlChecker is a implementation of MySQLChecker, it checks the local mysql and updates the POD labels.
type mysqlChecker struct {
	SQLRunner
	Client
}

type Client interface {
	ReadLabels() map[string]string
	PatchLabel(k, v string) error
}

// inClusterClient is a implementation of client, it operates the pod itself through the clientSet.
type inClusterClient struct {
	*kubernetes.Clientset
	ClientOptions
}

type ClientOptions struct {
	NameSpace string
	PodName   string
}

type RaftStatus struct {
	Leader string   `json:"leader"`
	State  string   `json:"state"`
	Nodes  []string `json:"nodes"`
}

func GetInclusterClientSet() *kubernetes.Clientset {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed to create InClusterConfig: %s", err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to get clientset: %s", err.Error())
	}
	return clientset
}

func XenonPingMyself() error {
	args := []string{"xenon", "ping"}
	cmd := exec.Command("xenoncli", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to exec xenoncli xenon ping: %v", err)
	}
	return nil
}

func GetRaftStatus() *RaftStatus {
	args := []string{"raft", "status"}
	cmd := exec.Command("xenoncli", args...)
	res, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to exec xenoncli raft status: %v", err)
	}
	raftStatus := RaftStatus{}
	if err := json.Unmarshal(res, &raftStatus); err != nil {
		log.Fatalf("failed to unmarshal raft status: %v", err)
	}
	return &raftStatus
}

func GetRole() string {
	return GetRaftStatus().State
}

func NewInclusterClient(options *ClientOptions) *inClusterClient {
	return &inClusterClient{
		GetInclusterClientSet(),
		*options,
	}
}

func (c *inClusterClient) PatchLabel(k, v string) error {
	pod, err := c.Clientset.CoreV1().Pods(c.NameSpace).Get(context.TODO(), c.PodName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	pod.Labels[k] = v
	_, err = c.Clientset.CoreV1().Pods(c.NameSpace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	return err
}

func (c *inClusterClient) ReadLabels() map[string]string {
	pod, err := c.Clientset.CoreV1().Pods(c.NameSpace).Get(context.TODO(), c.PodName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("failed to get pod: %s", err.Error())
		return nil
	}
	return pod.Labels
}

func GetMyRole(labels map[string]string) string {
	if role, ok := labels["role"]; ok {
		return role
	}
	return string(utils.Unknown)
}

func GetMyHealthy(labels map[string]string) string {
	if health, ok := labels["healthy"]; ok {
		return health
	}
	return "no"
}

func ConvertHealthy(isHealthy bool) string {
	if isHealthy {
		return "yes"
	}
	return "no"
}

func NewMysqlChecker(s SQLRunner, options *ClientOptions) MySQLChecker {
	return &mysqlChecker{
		s,
		NewInclusterClient(options),
	}
}

func SleepForever() bool {
	_, err := os.Stat("/var/lib/mysql/sleep-forever")
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		log.Errorf("failed to check sleep-forever: %s", err.Error())
		return false
	}
	return true
}

func (m *mysqlChecker) CheckMySQLLiveness() error {
	return m.select1()
}

func (m *mysqlChecker) select1() error {
	if _, err := m.QueryRows(NewQuery("select 1;")); err != nil {
		return err
	}
	return nil
}

func (m *mysqlChecker) CheckMySQLHealthy(role string) bool {
	res := m.checkMySQL()
	switch role {
	case string(utils.Leader):
		return m.checkLeader(res)
	case string(utils.Follower):
		return m.checkFollower(res)
	default:
		log.Infof("check %s", role)
		return m.isReadonly()
	}
}

func (m *mysqlChecker) isReadonly() bool {
	var readOnly uint8
	if err := m.QueryRow(NewQuery("select @@global.read_only;"), &readOnly); err != nil {
		log.Errorf("failed to get read_only: %s", err.Error())
		return false
	}
	return readOnly == 1
}

type checkResult struct {
	readOnly, isLagged, isReplicating bool
}

func (m *mysqlChecker) checkMySQL() checkResult {
	isLagged, isReplicating := m.showSlaveStatus()
	return checkResult{
		readOnly:      m.isReadonly(),
		isLagged:      isLagged,
		isReplicating: isReplicating,
	}
}

func (m *mysqlChecker) checkLeader(res checkResult) bool {
	log.Infof("check leader, readonly: %t, isLagged: %t, isReplicating: %t", res.readOnly, res.isLagged, res.isReplicating)

	if !m.existUpdateFile() && res.readOnly {
		log.Errorf("im leader but read_only is on")
		if err := m.setGlobalReadOnly(false); err != nil {
			log.Errorf("failed to set read_only to off: %s", err.Error())
			return false
		}
	}
	return !res.readOnly && !res.isReplicating
}

func (m *mysqlChecker) checkFollower(res checkResult) bool {
	log.Infof("check follower, readonly: %t, isLagged: %t, isReplicating: %t", res.readOnly, res.isLagged, res.isReplicating)
	return res.readOnly && !res.isLagged && res.isReplicating
}

func (m *mysqlChecker) showSlaveStatus() (bool, bool) {
	lagged, replicating, err := CheckSlaveStatusWithRetry(m.SQLRunner, 3)
	if err != nil {
		log.Errorf("failed to show slave status: %s", err.Error())
		return false, false
	}
	isLagged := lagged == corev1.ConditionTrue
	isReplicating := replicating == corev1.ConditionTrue
	return isLagged, isReplicating
}

func (m *mysqlChecker) setGlobalReadOnly(on bool) error {
	option := "OFF"
	if on {
		option = "ON"
	}
	log.Infof("try to turn %s read_only", option)

	if err := m.SQLRunner.QueryExec(NewQuery("SET GLOBAL read_only==?;", option)); err != nil {
		log.Errorf("failed to set global read_only: %s", err.Error())
		return err
	}
	if err := m.SQLRunner.QueryExec(NewQuery("SET GLOBAL super_read_only==?;", option)); err != nil {
		log.Errorf("failed to set global super_read_only: %s", err.Error())
		return err
	}
	return nil
}

func (m *mysqlChecker) existUpdateFile() bool {
	return utils.ExistUpdateFile()
}
