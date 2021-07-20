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
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

const (
	serverPort           = utils.XBackupPort
	serverProbeEndpoint  = "/health"
	serverBackupEndpoint = "/xbackup"
	serverConnectTimeout = 5 * time.Second
)

type server struct {
	cfg *Config
	http.Server
}

func newServer(cfg *Config, stop <-chan struct{}) *server {
	mux := http.NewServeMux()
	srv := &server{
		cfg: cfg,
		Server: http.Server{
			Addr:    fmt.Sprintf(":%d", serverPort),
			Handler: mux,
		},
	}

	// Add handle functions
	mux.HandleFunc(serverProbeEndpoint, srv.healthHandler)
	mux.Handle(serverBackupEndpoint, maxClients(http.HandlerFunc(srv.backupHandler), 1))

	// Shutdown gracefully the http server
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
	if !s.isAuthenticated(r) {
		http.Error(w, "Not authenticated!", http.StatusForbidden)
		return
	}
	err := RunTakeBackupCommand(s.cfg, "hello")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Write([]byte("OK"))
	}
}

func (s *server) isAuthenticated(r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	return ok && user == s.cfg.BackupUser && pass == s.cfg.BackupPassword
}

// maxClients limit an http endpoint to allow just n max concurrent connections
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

// requestABackup connects to specified host and endpoint and gets the backup
func requestABackup(cfg *Config, host string, endpoint string) (*http.Response, error) {
	log.Info("initialize a backup", "host", host, "endpoint", endpoint)

	req, err := http.NewRequest("GET", prepareURL(host, endpoint), nil)
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

	return resp, nil
}
