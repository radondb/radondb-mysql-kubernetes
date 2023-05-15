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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

const (
	// backupStatus http trailer
	backupStatusTrailer = "X-Backup-Status"

	// success string
	backupSuccessful = "Success"

	// failure string
	backupFailed = "Failed"

	serverPort           = utils.XBackupPort
	serverProbeEndpoint  = "/health"
	serverBackupEndpoint = "/xbackup"
	serverConnectTimeout = 5 * time.Second

	// DownLoad server url.
	serverBackupDownLoadEndpoint = "/download"
)

type server struct {
	cfg *Config
	http.Server
}

// Create new Http Server.
func newServer(cfg *Config, stop <-chan struct{}) *server {
	mux := http.NewServeMux()
	srv := &server{
		cfg: cfg,
		Server: http.Server{
			Addr:    fmt.Sprintf(":%d", serverPort),
			Handler: mux,
		},
	}

	// Add handle functions.
	// Health check
	mux.HandleFunc(serverProbeEndpoint, srv.healthHandler)
	// Backup server
	mux.Handle(serverBackupEndpoint, maxClients(http.HandlerFunc(srv.backupHandler), 1))
	// Backup download server.
	mux.Handle(serverBackupDownLoadEndpoint,
		maxClients(http.HandlerFunc(srv.backupDownloadHandler), 1))

	// Shutdown gracefully the http server.
	go func() {
		<-stop // wait for stop signal
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Error(err, "failed to stop http server")
		}
	}()

	return srv
}

// nolint: unparam
func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Error(err, "failed writing request")
	}
}

func (s *server) backupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("content-type", "text/json")

	// Extract backup name from POST body

	var requestBody BackupClientConfig
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !s.isAuthenticated(r) {
		http.Error(w, "Not authenticated!", http.StatusForbidden)
		return
	}
	// /backup only handle S3 backup
	if requestBody.BackupType == S3 {

		backName, Datetime, backupSize, err := RunTakeS3BackupCommand(&requestBody)
		log.Info("get backup result", "backName", backName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			msg, _ := json.Marshal(utils.JsonResult{Status: backupSuccessful, BackupName: backName, Date: Datetime, BackupSize: backupSize})
			w.Write(msg)
		}
	}
}

func (s *server) backupDownloadHandler(w http.ResponseWriter, r *http.Request) {

	if !s.isAuthenticated(r) {
		http.Error(w, "Not authenticated!", http.StatusForbidden)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "HTTP server does not support streaming!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Trailer", backupStatusTrailer)

	// nolint: gosec
	xtrabackup := exec.Command(xtrabackupCommand, s.cfg.XtrabackupArgs()...)
	xtrabackup.Stderr = os.Stderr

	stdout, err := xtrabackup.StdoutPipe()
	if err != nil {
		log.Error(err, "failed to create stdout pipe")
		http.Error(w, "xtrabackup failed", http.StatusInternalServerError)
		return
	}

	defer func() {
		// don't care
		_ = stdout.Close()
	}()

	if err := xtrabackup.Start(); err != nil {
		log.Error(err, "failed to start xtrabackup command")
		http.Error(w, "xtrabackup failed", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(w, stdout); err != nil {
		log.Error(err, "failed to copy buffer")
		http.Error(w, "buffer copy failed", http.StatusInternalServerError)
		return
	}

	if err := xtrabackup.Wait(); err != nil {
		log.Error(err, "failed waiting for xtrabackup to finish")
		w.Header().Set(backupStatusTrailer, backupFailed)
		http.Error(w, "xtrabackup failed", http.StatusInternalServerError)
		return
	}

	// success
	w.Header().Set(backupStatusTrailer, backupSuccessful)
	flusher.Flush()
}

func (s *server) isAuthenticated(r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	return ok && user == s.cfg.BackupUser && pass == s.cfg.BackupPassword
}

// maxClients limit an http endpoint to allow just n max concurrent connections.
func maxClients(h http.Handler, n int) http.Handler {
	sema := make(chan struct{}, n)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		defer func() {
			<-sema
		}()
		h.ServeHTTP(w, r)
	})
}

func prepareURL(svc string, endpoint string) string {
	if !strings.Contains(svc, ":") {
		svc = fmt.Sprintf("%s:%d", svc, serverPort)
	}
	return fmt.Sprintf("http://%s%s", svc, endpoint)
}

// Set the timeout for HTTP.
func transportWithTimeout(connectTimeout time.Duration) http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   connectTimeout,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
