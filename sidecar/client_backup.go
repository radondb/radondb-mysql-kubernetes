package sidecar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

func requestS3Backup(cfg *BackupClientConfig, host string, endpoint string) (*http.Response, error) {

	log.Info("initialize a backup", "host", host, "endpoint", endpoint)
	reqBody, err := json.Marshal(cfg)
	if err != nil {
		log.Error(err, "fail to marshal request body")
		return nil, fmt.Errorf("fail to marshal request body: %s", err)
	}

	req, err := http.NewRequest("POST", prepareURL(host, endpoint), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("fail to create request: %s", err)
	}

	// set authentication user and password
	req.SetBasicAuth(cfg.BackupUser, cfg.BackupPassword)

	client := &http.Client{}
	client.Transport = transportWithTimeout(serverConnectTimeout)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		status := "unknown"
		if resp != nil {
			status = resp.Status
		}
		return nil, fmt.Errorf("fail to get backup: %s, code: %s", err, status)
	}
	defer resp.Body.Close()
	var result utils.JsonResult
	json.NewDecoder(resp.Body).Decode(&result)

	err = setAnnonations(cfg, result.BackupName, result.Date, "S3", result.BackupSize) // set annotation
	if err != nil {
		return nil, fmt.Errorf("fail to set annotation: %s", err)
	}
	return resp, nil
}

func requestNFSBackup(cfg *BackupClientConfig, host string, endpoint string) error {
	log.Info("initializing a NFS backup", "host", host, "endpoint", endpoint)

	backupName, DateTime := cfg.XBackupName()

	reqBody, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("GET", prepareURL(host, endpoint), bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(cfg.BackupUser, cfg.BackupPassword)

	client := &http.Client{
		Transport: transportWithTimeout(serverConnectTimeout),
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get backup: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get backup: HTTP status %s", resp.Status)
	}
	defer resp.Body.Close()

	// Create the backup dir
	// Backupdir is the name of the backup
	backupPath := fmt.Sprintf("%s/%s", "/backup", backupName)
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	// Create a pipe for the xbstream command to read from.
	rc, wc := io.Pipe()
	cmd := exec.Command("xbstream", "-x", "-C", backupPath)
	cmd.Stdin = rc
	cmd.Stderr = os.Stderr

	// Start xbstream command.
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start xbstream cmd: %w", err)
	}

	// Write the response body to the pipe.
	copyErr := make(chan error)
	var n int64
	go func() {
		n, err = io.Copy(wc, resp.Body)
		wc.Close()
		copyErr <- err
	}()

	// Wait for the xbstream command to finish.
	cmdErr := make(chan error)
	go func() {
		cmdErr <- cmd.Wait()
	}()

	if err := <-copyErr; err != nil {
		return fmt.Errorf("failed to write to pipe: %w", err)
	}

	if err := <-cmdErr; err != nil {
		return fmt.Errorf("xbstream command failed: %w", err)
	}

	if err := setAnnonations(cfg, backupName, DateTime, "nfs", n); err != nil {
		return fmt.Errorf("failed to set annotation: %w", err)
	}
	log.Info("backup completed", "backupName", backupName, "backupSize", n)

	return nil
}
