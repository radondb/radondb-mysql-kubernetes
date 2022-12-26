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

package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// errorConnectionStates contains the list of the Slave_IO_State message.
var errorConnectionStates = []string{
	"connecting to master",
	"reconnecting after a failed binlog dump request",
	"reconnecting after a failed master event read",
	"waiting to reconnect after a failed binlog dump request",
	"waiting to reconnect after a failed master event read",
}

var internalLog = log.Log.WithName("mysql-internal")

// Config is used to connect to a MysqlCluster.
type Config struct {
	User     string
	Password string
	Host     string
	Port     int32
}

// NewConfigFromClusterKey returns a new Config based on a MySQLCluster key.
func NewConfigFromClusterKey(c client.Client, clusterKey client.ObjectKey, userName, host string) (*Config, error) {
	cluster := &apiv1alpha1.MysqlCluster{}
	if err := c.Get(context.TODO(), clusterKey, cluster); err != nil {
		return nil, err
	}

	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{Name: mysqlcluster.New(cluster).GetNameForResource(utils.Secret), Namespace: cluster.Namespace}

	if err := c.Get(context.TODO(), secretKey, secret); err != nil {
		return nil, err
	}

	if host == utils.LeaderHost {
		host = fmt.Sprintf("%s-leader.%s", cluster.Name, cluster.Namespace)
	}

	switch userName {
	case utils.OperatorUser:
		password, ok := secret.Data["operator-password"]
		if !ok {
			return nil, fmt.Errorf("operator-password cannot be empty")
		}
		return &Config{
			User:     utils.OperatorUser,
			Password: string(password),
			Host:     host,
			Port:     utils.MysqlPort,
		}, nil

	case utils.RootUser:
		password, ok := secret.Data["internal-root-password"]
		if !ok {
			return nil, fmt.Errorf("internal-root-password cannot be empty")
		}
		return &Config{
			User:     utils.RootUser,
			Password: string(password),
			Host:     host,
			Port:     utils.MysqlPort,
		}, nil
	default:
		return nil, fmt.Errorf("MySQL user %s are not supported", userName)
	}

}

// GetMysqlDSN returns a data source name.
func (c *Config) GetMysqlDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=5s&multiStatements=true&interpolateParams=true",
		c.User, c.Password, c.Host, c.Port,
	)
}

type sqlRunner struct {
	db *sql.DB
}

// SQLRunner interface is a subset of mysql.DB.
type SQLRunner interface {
	QueryExec(query Query) error
	QueryRow(query Query, dest ...interface{}) error
	QueryRows(query Query) (*sql.Rows, error)
}

type closeFunc func()

// SQLRunnerFactory a function that generates a new SQLRunner.
type SQLRunnerFactory func(cfg *Config, errs ...error) (SQLRunner, closeFunc, error)

// NewSQLRunner opens a connections using the given DSN.
func NewSQLRunner(cfg *Config, errs ...error) (SQLRunner, closeFunc, error) {
	var db *sql.DB
	var close closeFunc = nil

	// Make this factory accept a functions that tries to generate a config.
	if len(errs) > 0 && errs[0] != nil {
		return nil, close, errs[0]
	}

	db, err := sql.Open("mysql", cfg.GetMysqlDSN())
	if err != nil {
		return nil, close, err
	}
	db.SetConnMaxIdleTime(10 * time.Second)
	db.SetConnMaxLifetime(1 * time.Minute)
	if err := db.Ping(); err != nil {
		internalLog.V(1).Info("failed to ping mysql", "error", err)
		if cErr := db.Close(); cErr != nil {
			internalLog.Error(cErr, "failed closing the database connection")
		}
		return nil, close, err
	}

	// Close connection function.
	close = func() {
		if cErr := db.Close(); cErr != nil {
			internalLog.Error(cErr, "failed closing the database connection")
		}
	}

	return &sqlRunner{db: db}, close, nil
}

// QueryExec used to run the query with args.
func (s sqlRunner) QueryExec(query Query) error {
	if _, err := s.db.Exec(query.String(), query.args...); err != nil {
		return err
	}

	return nil
}

func (s sqlRunner) QueryRow(query Query, dest ...interface{}) error {
	return s.db.QueryRow(query.escapedQuery, query.args...).Scan(dest...)
}

func (s sqlRunner) QueryRows(query Query) (*sql.Rows, error) {
	rows, err := s.db.Query(query.escapedQuery, query.args...)
	if err != nil {
		return nil, err
	}

	return rows, rows.Err()
}

// CheckSlaveStatusWithRetry check the slave status with retry time.
func CheckSlaveStatusWithRetry(sqlRunner SQLRunner, retry uint32) (isLagged, isReplicating corev1.ConditionStatus, err error) {
	for {
		if retry == 0 {
			break
		}

		if isLagged, isReplicating, err = checkSlaveStatus(sqlRunner); err == nil {
			return
		}

		time.Sleep(time.Second * 3)
		retry--
	}

	return
}

// checkSlaveStatus check the slave status.
func checkSlaveStatus(sqlRunner SQLRunner) (isLagged, isReplicating corev1.ConditionStatus, err error) {
	var rows *sql.Rows
	isLagged, isReplicating = corev1.ConditionUnknown, corev1.ConditionUnknown
	rows, err = sqlRunner.QueryRows(NewQuery("show slave status;"))
	if err != nil {
		return
	}

	defer rows.Close()

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return
		}
		return corev1.ConditionFalse, corev1.ConditionFalse, nil
	}

	var cols []string
	cols, err = rows.Columns()
	if err != nil {
		return
	}

	scanArgs := make([]interface{}, len(cols))
	for i := range scanArgs {
		scanArgs[i] = &sql.RawBytes{}
	}

	err = rows.Scan(scanArgs...)
	if err != nil {
		return
	}

	slaveIOState := strings.ToLower(columnValue(scanArgs, cols, "Slave_IO_State"))
	slaveSQLRunning := columnValue(scanArgs, cols, "Slave_SQL_Running")
	lastSQLError := columnValue(scanArgs, cols, "Last_SQL_Error")
	secondsBehindMaster := columnValue(scanArgs, cols, "Seconds_Behind_Master")

	if utils.StringInArray(slaveIOState, errorConnectionStates) {
		return isLagged, corev1.ConditionFalse, fmt.Errorf("Slave_IO_State: %s", slaveIOState)
	}

	if slaveSQLRunning != "Yes" {
		return isLagged, corev1.ConditionFalse, fmt.Errorf("Last_SQL_Error: %s", lastSQLError)
	}

	isReplicating = corev1.ConditionTrue

	var longQueryTime float64
	if err = GetGlobalVariable(sqlRunner, "long_query_time", &longQueryTime); err != nil {
		return
	}

	// Check whether the slave is lagged.
	sec, _ := strconv.ParseFloat(secondsBehindMaster, 64)
	if sec > longQueryTime*100 {
		isLagged = corev1.ConditionTrue
	} else {
		isLagged = corev1.ConditionFalse
	}

	return
}

// CheckReadOnly check whether the mysql is read only.
func CheckReadOnly(sqlRunner SQLRunner) (corev1.ConditionStatus, error) {
	var readOnly uint8
	if err := GetGlobalVariable(sqlRunner, "read_only", &readOnly); err != nil {
		return corev1.ConditionUnknown, err
	}

	if readOnly == 0 {
		return corev1.ConditionFalse, nil
	}

	return corev1.ConditionTrue, nil
}

// GetGlobalVariable used to get the global variable by param.
func GetGlobalVariable(sqlRunner SQLRunner, param string, val interface{}) error {
	return sqlRunner.QueryRow(NewQuery("select @@global.?", param), val)
}

func CheckProcesslist(sqlRunner SQLRunner) (bool, error) {
	var rows *sql.Rows
	rows, err := sqlRunner.QueryRows(NewQuery("show processlist;"))
	if err != nil {
		return false, err
	}

	defer rows.Close()

	var cols []string
	cols, err = rows.Columns()
	if err != nil {
		return false, err
	}

	scanArgs := make([]interface{}, len(cols))
	for i := range scanArgs {
		scanArgs[i] = &sql.RawBytes{}
	}

	for rows.Next() {
		if err = rows.Scan(scanArgs...); err != nil {
			return false, err
		}

		state := columnValue(scanArgs, cols, "State")
		if strings.Contains(state, "Master has sent all binlog to slave") {
			return true, nil
		}
	}
	return false, nil
}

// columnValue get the column value.
func columnValue(scanArgs []interface{}, slaveCols []string, colName string) string {
	columnIndex := -1
	for idx := range slaveCols {
		if slaveCols[idx] == colName {
			columnIndex = idx
			break
		}
	}

	if columnIndex == -1 {
		return ""
	}

	return string(*scanArgs[columnIndex].(*sql.RawBytes))
}

// BuildUserManagementSQL returns a Query that creates a user and grants it permissions.
func BuildUserManagementSQL(user *apiv1alpha1.MysqlUser, pass string) (q Query, err error) {
	userName := user.Spec.User
	hosts := user.Spec.Hosts
	permissions := user.Spec.Permissions
	query := Query{}
	// Throw error if there are no allowed hosts.
	if len(hosts) == 0 {
		return query, errors.New("no allowedHosts specified")
	}

	queries := []Query{
		getCreateUserQuery(userName, pass, hosts, user.Spec.TLSOptions),
		getAlterUserQuery(userName, pass, hosts),
	}

	if len(permissions) > 0 {
		queries = append(queries, permissionsToQuery(permissions, userName, hosts, user.Spec.WithGrantOption))
	}

	query = BuildAtomicQuery(queries...)

	return query, nil
}

func getCreateUserQuery(user, pwd string, allowedHosts []string, tlsOption apiv1alpha1.TLSOptions) Query {
	idsTmpl, idsArgs := getUsersIdentification(user, &pwd, allowedHosts)
	idsTmpl += getUserTLSRequire(tlsOption)

	return NewQuery(fmt.Sprintf("CREATE USER IF NOT EXISTS%s", idsTmpl), idsArgs...)
}

func getUserTLSRequire(tlsOption apiv1alpha1.TLSOptions) string {
	return fmt.Sprintf(" REQUIRE %s", tlsOption.Type)
}

// Only support changing passwords.
func getAlterUserQuery(user, pwd string, allowedHosts []string) Query {
	args := []interface{}{}
	q := "ALTER USER"

	ids, idsArgs := getUsersIdentification(user, &pwd, allowedHosts)
	q += ids
	args = append(args, idsArgs...)

	return NewQuery(q, args...)
}

func getUsersIdentification(user string, pwd *string, allowedHosts []string) (ids string, args []interface{}) {
	for i, host := range allowedHosts {
		// Add comma if more than one allowed hosts are used.
		if i > 0 {
			ids += ","
		}

		if pwd != nil {
			ids += " ?@? IDENTIFIED BY ?"
			args = append(args, user, host, *pwd)
		} else {
			ids += " ?@?"
			args = append(args, user, host)
		}
	}

	return ids, args
}

// DropUser removes a MySQL user if it exists, along with its privileges.
func DropUser(sqlRunner SQLRunner, user, host string) error {
	query := NewQuery("DROP USER IF EXISTS ?@?;", user, host)

	if err := sqlRunner.QueryExec(query); err != nil {
		return fmt.Errorf("failed to delete user, err: %s", err)
	}

	return nil
}

func permissionsToQuery(permissions []apiv1alpha1.UserPermission, user string, allowedHosts []string, withGrant bool) Query {
	permQueries := []Query{}

	for _, perm := range permissions {
		// If you wish to grant permissions on all tables, you should explicitly use "*".
		for _, table := range perm.Tables {
			args := []interface{}{}

			escPerms := []string{}
			for _, perm := range perm.Privileges {
				escPerms = append(escPerms, Escape(perm))
			}

			schemaTable := fmt.Sprintf("%s.%s", escapeID(perm.Database), escapeID(table))

			// Build GRANT query.
			idsTmpl, idsArgs := getUsersIdentification(user, nil, allowedHosts)

			query := "GRANT " + strings.Join(escPerms, ", ") + " ON " + schemaTable + " TO" + idsTmpl
			if withGrant {
				query += " WITH GRANT OPTION"
			}
			args = append(args, idsArgs...)

			permQueries = append(permQueries, NewQuery(query, args...))
		}
	}

	return ConcatenateQueries(permQueries...)
}

func escapeID(id string) string {
	if id == "*" {
		return id
	}

	// don't allow using ` in id name.
	id = strings.ReplaceAll(id, "`", "")

	return fmt.Sprintf("`%s`", id)
}

// Escape escapes a string.
func Escape(sql string) string {
	dest := make([]byte, 0, 2*len(sql))
	var escape byte
	for i := 0; i < len(sql); i++ {
		escape = 0
		switch sql[i] {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
		case '\n': /* Must be escaped for logs */
			escape = 'n'
		case '\r':
			escape = 'r'
		case '\\':
			escape = '\\'
		case '\'':
			escape = '\''
		case '"': /* Better safe than sorry */
			escape = '"'
		case '\032': /* This gives problems on Win32 */
			escape = 'Z'
		}

		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, sql[i])
		}
	}

	return string(dest)
}
