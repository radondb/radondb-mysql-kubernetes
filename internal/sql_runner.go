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
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	corev1 "k8s.io/api/core/v1"

	mysqlv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
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

// SQLRunner is a runner for run the sql.
type SQLRunner struct {
	db *sql.DB
}

// NewSQLRunner return a pointer to SQLRunner.
func NewSQLRunner(user, password, host string, port int) (*SQLRunner, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=5s&interpolateParams=true&multiStatements=true",
		user, password, host, port,
	)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &SQLRunner{db}, nil
}

// CheckSlaveStatusWithRetry check the slave status with retry time.
func (s *SQLRunner) CheckSlaveStatusWithRetry(retry uint32) (isLagged, isReplicating corev1.ConditionStatus, err error) {
	for {
		if retry == 0 {
			break
		}

		if isLagged, isReplicating, err = s.checkSlaveStatus(); err == nil {
			return
		}

		time.Sleep(time.Second * 3)
		retry--
	}

	return
}

// checkSlaveStatus check the slave status.
func (s *SQLRunner) checkSlaveStatus() (isLagged, isReplicating corev1.ConditionStatus, err error) {
	var rows *sql.Rows
	isLagged, isReplicating = corev1.ConditionUnknown, corev1.ConditionUnknown
	rows, err = s.db.Query("show slave status;")
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
	if err = s.GetGlobalVariable("long_query_time", &longQueryTime); err != nil {
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
func (s *SQLRunner) CheckReadOnly() (corev1.ConditionStatus, error) {
	var readOnly uint8
	if err := s.GetGlobalVariable("read_only", &readOnly); err != nil {
		return corev1.ConditionUnknown, err
	}

	if readOnly == 0 {
		return corev1.ConditionFalse, nil
	}

	return corev1.ConditionTrue, nil
}

// RunQuery used to run the query with args.
func (s *SQLRunner) RunQuery(query string, args ...interface{}) error {
	if _, err := s.db.Exec(query, args...); err != nil {
		return err
	}

	return nil
}

// GetGlobalVariable used to get the global variable by param.
func (s *SQLRunner) GetGlobalVariable(param string, val interface{}) error {
	query := fmt.Sprintf("select @@global.%s", param)
	return s.db.QueryRow(query).Scan(val)
}

func (s *SQLRunner) CheckProcesslist() (bool, error) {
	var rows *sql.Rows
	rows, err := s.db.Query("show processlist;")
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

// Close closes the database and prevents new queries from starting.
func (sr *SQLRunner) Close() error {
	return sr.db.Close()
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

// CheckUserQuery check if the user exist.
func (s *SQLRunner) CheckUserQuery(query string, args ...interface{}) error {
	var result string
	err := s.db.QueryRow(query, args...).Scan(&result)
	if err != nil {
		return fmt.Errorf("check user faild, err:%s", err)
	}

	if result == "" {
		return fmt.Errorf("user is not exist")
	}

	return nil
}

// GetCreateQuery get the query of create users.
func GetCreateQuery(ctx context.Context, hosts []string, userName, password string) Query {
	idsTmpl, idsArgs := getUsersIdentification(userName, &password, hosts)

	return NewQuery(fmt.Sprintf("CREATE USER IF NOT EXISTS%s", idsTmpl), idsArgs...)
}

// GetDeleteQuery get the query of delete users.
func GetDeleteQuery(ctx context.Context, hosts []string, userName string) Query {
	queries := []Query{}
	for _, host := range hosts {
		queries = append(queries, NewQuery("DROP USER IF EXISTS ?@?;", userName, host))
	}

	return BuildAtomicQuery(queries...)
}

// GetGrantQuery get the query for grant.
func GetGrantQuery(permissions []mysqlv1alpha1.UserPermission, user string, allowedHosts []string) Query {
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
			args = append(args, idsArgs...)

			permQueries = append(permQueries, NewQuery(query, args...))
		}
	}

	return ConcatenateQueries(permQueries...)
}

// SetPasswordQuery set mysql user password.
func SetPasswordQuery(ctx context.Context, name, host, password string) Query {
	return NewQuery("ALTER USER ?@? IDENTIFIED BY ?;", name, host, password)
}

// CheckUserQuery get the query of check if the user exist.
func CheckUserQuery(ctx context.Context, name, host string) Query {
	return NewQuery("SELECT USER FROM mysql.user WHERE user = ? AND host = ?;", name, host)
}

// getUsersIdentification return the identification(name,host,password(if exist)) of user.
func getUsersIdentification(user string, pwd *string, hosts []string) (ids string, args []interface{}) {
	for i, host := range hosts {
		// add comma if more than one allowed hosts are used.
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

// escapeID remove the ` in id.
func escapeID(id string) string {
	if id == "*" {
		return id
	}

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
