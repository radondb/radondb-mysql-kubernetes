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
	"os"
	"os/exec"
	"strings"
)

// RunTakeBackupCommand starts a backup command
func RunTakeBackupCommand(cfg *Config, name string) error {
	// cfg->XtrabackupArgs()
	xtrabackup := exec.Command(xtrabackupCommand, cfg.XtrabackupArgs()...)

	var err error
	xcloud := exec.Command(xcloudCommand, cfg.XCloudArgs()...)
	log.Info("xargs ", "xargs", strings.Join(cfg.XCloudArgs(), " "))
	if xcloud.Stdin, err = xtrabackup.StdoutPipe(); err != nil {
		log.Error(err, "failed to pipline")
		return err
	}
	xtrabackup.Stderr = os.Stderr
	xcloud.Stderr = os.Stderr

	if err := xtrabackup.Start(); err != nil {
		log.Error(err, "failed to start xtrabackup command")
		return err
	}
	if err := xcloud.Start(); err != nil {
		log.Error(err, "fail start xcloud ")
		return err
	}

	// pipe command fail one, whole things fail
	errorChannel := make(chan error, 2)
	go func() {
		errorChannel <- xcloud.Wait()
	}()
	go func() {
		errorChannel <- xtrabackup.Wait()
	}()

	for i := 0; i < 2; i++ {
		if err = <-errorChannel; err != nil {
			return err
		}
	}
	return nil
}
