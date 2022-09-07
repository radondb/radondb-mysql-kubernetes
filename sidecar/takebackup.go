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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// RunTakeBackupCommand starts a backup command
func RunTakeBackupCommand(cfg *Config) (string, string, string, error) {
	// cfg->XtrabackupArgs()
	xtrabackup := exec.Command(xtrabackupCommand, cfg.XtrabackupArgs()...)

	var err error
	backupName, DateTime := cfg.XBackupName()
	Gtid := ""
	xcloud := exec.Command(xcloudCommand, cfg.XCloudArgs(backupName)...)
	log.Info("xargs ", "xargs", strings.Join(cfg.XCloudArgs(backupName), " "))
	if xcloud.Stdin, err = xtrabackup.StdoutPipe(); err != nil {
		log.Error(err, "failed to pipline")
		return "", "", "", err
	}
	//xtrabackup.Stderr = os.Stderr
	xcloud.Stderr = os.Stderr

	var wg sync.WaitGroup
	Stderr, err := xtrabackup.StderrPipe()
	if err != nil {
		return "", "", "", fmt.Errorf("RunCommand: cmd.StderrPipe(): %v", err)
	}
	if err := xtrabackup.Start(); err != nil {
		log.Error(err, "failed to start xtrabackup command")
		return "", "", "", err
	}
	if err := xcloud.Start(); err != nil {
		log.Error(err, "fail start xcloud ")
		return "", "", "", err
	}
	scanner := bufio.NewScanner(Stderr)
	//scanner.Split(ScanLinesR)
	wg.Add(1)
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			fmt.Println(text)
			if index := strings.Index(text, "GTID"); index != -1 {
				// Mysql5.7 examples: MySQL binlog position: filename 'mysql-bin.000002', position '588', GTID of the last change '319bd6eb-2ea2-11ed-bf40-7e1ef582b427:1-2'
				// MySQL8.0 no gtid:  MySQL binlog position: filename 'mysql-bin.000025', position '156'
				length := len("GTID of the last change")
				Gtid = strings.Trim(text[index+length:], " '") // trim space and \'
				if len(Gtid) != 0 {
					log.Info("Catch gtid: " + Gtid)
				}

			}
		}
		wg.Done()
	}()

	wg.Wait()
	// pipe command fail one, whole things fail
	errorChannel := make(chan error, 2)
	go func() {
		errorChannel <- xcloud.Wait()
	}()
	go func() {
		errorChannel <- xtrabackup.Wait()
	}()
	defer xtrabackup.Wait()
	defer xcloud.Wait()

	for i := 0; i < 2; i++ {
		if err = <-errorChannel; err != nil {
			log.Info("catch error , need to stop")
			_ = xtrabackup.Process.Kill()
			_ = xcloud.Process.Kill()

			return "", "", "", err
		}
	}

	return backupName, DateTime, Gtid, nil
}
