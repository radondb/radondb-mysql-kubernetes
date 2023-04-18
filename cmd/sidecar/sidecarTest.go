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

package main

import (
	"testing"

	"github.com/radondb/radondb-mysql-kubernetes/sidecar"
	"github.com/stretchr/testify/assert"
)

func TestGetContainerName(t *testing.T) {
	containerName := sidecar.GetContainerType()
	assert.NotNil(t, containerName)
}

// func TestBackupConfig(t *testing.T) {
// 	backupCfg := sidecar.NewBackupConfig()
// 	assert.NotNil(t, backupCfg)
// }

func TestReqBackupConfig(t *testing.T) {
	reqBackupCfg := sidecar.NewReqBackupConfig()
	assert.NotNil(t, reqBackupCfg)
}

func TestInitConfig(t *testing.T) {
	initCfg := sidecar.NewInitConfig()
	assert.NotNil(t, initCfg)
}

// func TestRunHttpServer(t *testing.T) {
// 	backupCfg := sidecar.NewBackupConfig()
// 	stop := make(chan struct{}, 1)
// 	err := sidecar.RunHttpServer(backupCfg, stop)
// 	assert.Error(t, err)
// }

func TestRunRequestBackup(t *testing.T) {
	reqBackupCfg := sidecar.NewReqBackupConfig()
	err := sidecar.RunRequestBackup(reqBackupCfg, "test")
	assert.Error(t, err)
}

func TestInitCommand(t *testing.T) {
	initCfg := sidecar.NewInitConfig()
	initCmd := sidecar.NewInitCommand(initCfg)
	assert.NotNil(t, initCmd)
}
