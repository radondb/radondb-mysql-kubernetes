package sidecar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

func requestABackup(cfg *BackupClientConfig, host string, endpoint string) (*http.Response, error) {

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
