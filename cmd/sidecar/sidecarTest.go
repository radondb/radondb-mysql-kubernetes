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
