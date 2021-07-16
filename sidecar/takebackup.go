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
	log.Info("backup mysql", "name", name)
	// cfg->XtrabackupArgs()
	xtrabackup := exec.Command(xtrabackupCommand, cfg.XtrabackupArgs()...)

	var err error
	//if len(cfg.XCloudS3AccessKey) == 0 || len(cfg.XCloudS3Bucket) == 0 || len(cfg.X)
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
	//xbcloud may be fail
	if err := xcloud.Wait(); err != nil {
		log.Error(err, "failed waiting for xcloud to finish")
		return err
	}
	// xtrabackup fail rarely
	if err := xtrabackup.Wait(); err != nil {
		log.Error(err, "failed waiting for xtrabackup to finish")
		return err
	}

	return nil
}
