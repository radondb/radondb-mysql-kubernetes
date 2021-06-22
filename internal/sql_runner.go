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
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	corev1 "k8s.io/api/core/v1"
)

var (
	// errorConnectionStates contains the list of the Slave_IO_State message.
	errorConnectionStates = []string{
		"connecting to master",
		"reconnecting after a failed binlog dump request",
		"reconnecting after a failed master event read",
		"waiting to reconnect after a failed binlog dump request",
		"waiting to reconnect after a failed master event read",
	}
)

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

	if stringInArray(slaveIOState, errorConnectionStates) {
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
func (sr *SQLRunner) GetGlobalVariable(param string, val interface{}) error {
	query := fmt.Sprintf("select @@global.%s", param)
	return sr.db.QueryRow(query).Scan(val)
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

// stringInArray check whether the str is in the strArray.
func stringInArray(str string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, str)
	return index < len(strArray) && strArray[index] == str
}
