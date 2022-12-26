package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-ini/ini"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	clientConfDir         = "/etc/my.cnf.d/client.conf"
	connectionMaxIdleTime = 30 * time.Second
	connectionTimeout     = 30 * time.Second
	raftStatusCmd         = "xenoncli raft status"
)

type RaftStatus struct {
	State  string   `json:"state"`
	Leader string   `json:"leader"`
	Nodes  []string `json:"nodes"`
}

type SlaveStatus struct {
	LastIOErrno               int           `db:"Last_IO_Errno"`
	LastIOError               string        `db:"Last_IO_Error"`
	LastSQLErrno              int           `db:"Last_SQL_Errno"`
	LastSQLError              string        `db:"Last_SQL_Error"`
	MasterHost                string        `db:"Master_Host"`
	RetrievedGtidSet          string        `db:"Retrieved_Gtid_Set"`
	ExecutedGtidSet           string        `db:"Executed_Gtid_Set"`
	SlaveIORunning            string        `db:"Slave_IO_Running"`
	SlaveSQLRunning           string        `db:"Slave_SQL_Running"`
	SlaveIOState              string        `db:"Slave_IO_State"`
	MasterUser                string        `db:"Master_User"`
	MasterPort                int           `db:"Master_Port"`
	ConnectRetry              int           `db:"Connect_Retry"`
	MasterLogFile             string        `db:"Master_Log_File"`
	ReadMasterLogPos          int           `db:"Read_Master_Log_Pos"`
	RelayLogFile              string        `db:"Relay_Log_File"`
	RelayLogPos               int           `db:"Relay_Log_Pos"`
	RelayMasterLogFile        string        `db:"Relay_Master_Log_File"`
	ReplicateDoDB             string        `db:"Replicate_Do_DB"`
	ReplicateIgnoreDB         string        `db:"Replicate_Ignore_DB"`
	ReplicateDoTable          string        `db:"Replicate_Do_Table"`
	ReplicateIgnoreTable      string        `db:"Replicate_Ignore_Table"`
	ReplicateWildDoTable      string        `db:"Replicate_Wild_Do_Table"`
	ReplicateWildIgnoreTable  string        `db:"Replicate_Wild_Ignore_Table"`
	LastErrno                 int           `db:"Last_Errno"`
	LastError                 string        `db:"Last_Error"`
	SkipCounter               int           `db:"Skip_Counter"`
	ExecMasterLogPos          int           `db:"Exec_Master_Log_Pos"`
	RelayLogSpace             int           `db:"Relay_Log_Space"`
	UntilCondition            string        `db:"Until_Condition"`
	UntilLogFile              string        `db:"Until_Log_File"`
	UntilLogPos               int           `db:"Until_Log_Pos"`
	MasterSSLAllowed          string        `db:"Master_SSL_Allowed"`
	MasterSSLCAFile           string        `db:"Master_SSL_CA_File"`
	MasterSSLCAPath           string        `db:"Master_SSL_CA_Path"`
	MasterSSLCert             string        `db:"Master_SSL_Cert"`
	MasterSSLCipher           string        `db:"Master_SSL_Cipher"`
	MasterSSLKey              string        `db:"Master_SSL_Key"`
	SecondsBehindMaster       sql.NullInt64 `db:"Seconds_Behind_Master"`
	MasterSSLVerifyServerCert string        `db:"Master_SSL_Verify_Server_Cert"`
	ReplicateIgnoreServerIds  string        `db:"Replicate_Ignore_Server_Ids"`
	MasterServerID            int           `db:"Master_Server_Id"`
	MasterUUID                string        `db:"Master_UUID"`
	MasterInfoFile            string        `db:"Master_Info_File"`
	SQLDelay                  int           `db:"SQL_Delay"`
	SQLRemainingDelay         sql.NullInt64 `db:"SQL_Remaining_Delay"`
	SlaveSQLRunningState      string        `db:"Slave_SQL_Running_State"`
	MasterRetryCount          int           `db:"Master_Retry_Count"`
	MasterBind                string        `db:"Master_Bind"`
	LastIOErrorTimestamp      string        `db:"Last_IO_Error_Timestamp"`
	LastSQLErrorTimestamp     string        `db:"Last_SQL_Error_Timestamp"`
	MasterSSLCrl              string        `db:"Master_SSL_Crl"`
	MasterSSLCrlpath          string        `db:"Master_SSL_Crlpath"`
	AutoPosition              string        `db:"Auto_Position"`
	ReplicateRewriteDB        string        `db:"Replicate_Rewrite_DB"`
	ChannelName               string        `db:"Channel_Name"`
	MasterTLSVersion          string        `db:"Master_TLS_Version"`
	Masterpublickeypath       string        `db:"Master_public_key_path"`
	Getmasterpublickey        string        `db:"Get_master_public_key"`
	NetworkNamespace          string        `db:"Network_Namespace"`
}

type Agent struct {
	conf      MySQLConfig
	db        *sqlx.DB
	maxDelay  time.Duration
	ksClient  *kubernetes.Clientset
	podName   string
	nameSpace string
}

type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}
type GTID struct {
	ID               string `json:"id"`
	Raft             string `json:"raft"`
	Mysql            string `json:"mysql"`
	ExecutedGTIDSet  string `json:"executed-gtid-set"`
	RetrievedGTIDSet string `json:"retrieved-gtid-set"`
}

func New() *Agent {
	debugFlag, _ := strconv.ParseBool(os.Getenv("RADONDB_DEBUG"))
	if debugFlag {
		log.SetLevel(log.DebugLevel)
	}
	conf := getMySQLclientConf()
	db, err := getMySQLConn(conf)
	if err != nil {
		log.Fatalf("get mysql connection failed: %v", err)
	}
	maxDelay, _ := strconv.Atoi(os.Getenv("MAX_DELAY"))
	ksCgent, err := utils.GetClientSet()
	if err != nil {
		log.Fatalf("get kubernetes clientset failed: %v", err)
	}
	podName := os.Getenv("POD_NAME")
	nameSpace := os.Getenv("NAMESPACE")

	return &Agent{
		conf:      *conf,
		db:        db,
		maxDelay:  time.Duration(maxDelay) * time.Second,
		ksClient:  ksCgent,
		podName:   podName,
		nameSpace: nameSpace,
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s leaderStart|leaderStop|liveness|readiness|postStart|preStop", os.Args[0])
	}
	agent := New()
	defer agent.CloseDB()
	switch os.Args[1] {
	case "liveness":
		if err := agent.liveness(); err != nil {
			log.Fatalf("liveness failed: %s", err.Error())
		}
	case "readiness":
		if err := agent.readiness(); err != nil {
			log.Fatalf("readiness failed: %s", err.Error())
		}
	case "postStart":
		if err := agent.postStart(); err != nil {
			log.Fatalf("postStart failed: %s", err.Error())
		}
	case "preStop":
		if err := agent.preStop(); err != nil {
			log.Fatalf("postStop failed: %s", err.Error())
		}
	default:
		log.Fatalf("Usage: %s leaderStart|leaderStop|liveness|readiness|postStart|preStop", os.Args[0])
	}
}
func (c *Agent) liveness() error {
	// if sleep-forever is set, then skip
	if utils.SleepFlag() {
		return nil
	}
	// get mysql pid from pid file
	if utils.IsMySQLRunning() {
		return nil
	} else {
		return fmt.Errorf("mysql is not running")
	}
}

func (c *Agent) readiness() error {
	// Check the instance works primary or not
	rows, err := c.db.Query("select @@read_only")
	if err != nil {
		return err
	}
	defer rows.Close()
	var readOnly bool
	for rows.Next() {
		if err := rows.Scan(&readOnly); err != nil {
			return err
		}
	}
	role, err := c.getRoleBylabel()
	if err != nil {
		return err
	}
	// slave check delay
	switch role {
	case "FOLLOWER":
		{
			// check if the cluster has leader
			raftStatus, err := c.getRaftStatus()
			if err != nil {
				return err
			}
			if raftStatus.Leader == "" {
				log.Warning("no leader found,skip readiness check")
				return nil
			}
			status := &SlaveStatus{}
			err1 := c.db.GetContext(context.Background(), status, `show slave status`)
			if err1 != nil {
				return err
			}

			if status.SlaveIORunning != "Yes" || status.SlaveSQLRunning != "Yes" {
				return fmt.Errorf("replication threads are stopped")
			}
			if status.LastError != "" {
				return fmt.Errorf("slave has error: %s", status.LastError)
			}
			if status.SecondsBehindMaster.Int64 > int64(c.maxDelay.Seconds()) {
				return fmt.Errorf("slave is too far behind master")
			}
		}
	case "LEADER":
		{
			if !utils.ExistUpdateFile() && readOnly {
				log.Errorf("am leader but read_only is on")
				if err := c.setGlobalReadOnlyOff(); err != nil {
					return err
				}
			}

		}

	}
	return nil
}

func (c *Agent) postStart() error {
	return nil
}

func (c *Agent) preStop() error {
	return nil
}
func (c *Agent) CloseDB() error {
	return c.db.Close()
}

func getMySQLclientConf() *MySQLConfig {
	// read config file
	cfg, err := ini.Load(clientConfDir)
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}
	// read section
	section, err := cfg.GetSection("client")
	if err != nil {
		log.Fatalf("Fail to get section: %v", err)
	}
	// read key
	host := section.Key("host").String()
	port, err := section.Key("port").Int()
	if err != nil {
		log.Fatalf("Fail to get port: %v", err)
	}
	password := section.Key("password").String()
	user := section.Key("user").String()
	return &MySQLConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}

}

func getMySQLConn(conf *MySQLConfig) (*sqlx.DB, error) {
	c := mysql.NewConfig()
	c.User = conf.User
	c.Passwd = conf.Password
	c.Net = "tcp"
	c.Addr = fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	c.Timeout = connectionTimeout
	c.ReadTimeout = connectionTimeout
	c.InterpolateParams = true
	c.ParseTime = true
	db, err := sqlx.Open("mysql", c.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(connectionMaxIdleTime)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetMaxIdleConns(1)
	return db, nil
}

func (c *Agent) getRoleBylabel() (string, error) {
	podMeta, err := c.ksClient.CoreV1().Pods(c.nameSpace).Get(context.TODO(), c.podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if role, ok := podMeta.Labels["role"]; ok {
		return role, nil
	}
	return "", fmt.Errorf("role label not found")
}

func (c *Agent) setGlobalReadOnlyOff() error {
	_, err := c.db.Exec("set global read_only=0")
	if err != nil {
		return fmt.Errorf("failed to disable super_read_only: %w", err)
	}
	return nil
}

func (c *Agent) getRaftStatus() (*RaftStatus, error) {
	config, err := utils.NewConfig()
	if err != nil {
		panic(err)
	}
	k, err := utils.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	cfg := utils.RunRemoteCommandConfig{
		PodName:   c.podName,
		Namespace: c.nameSpace,
		Container: "xenon",
	}

	raftStatusCmd := []string{raftStatusCmd}
	var output, stderr string
	output, stderr, err = utils.RunRemoteCommand(k, cfg, raftStatusCmd)
	log.Info("output=[" + output + "]")
	log.Info("stderr=[" + stderr + "]")
	if err != nil {
		log.Fatal(err)
	}
	status := &RaftStatus{}
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return status, nil

}
