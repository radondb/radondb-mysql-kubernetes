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
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

//
var (
	xenonHttpUser = utils.RootUser + ":"
)

// RequestConfig is the configuration required to get xenon http requests.
type RequestConfig struct {
	rootPassword string
	host         string
	xenonHttpUrl utils.XenonHttpUrl
	data         interface{}
}

func NewRequestConfig(host, rootPasswd string, xenonHttpUrl utils.XenonHttpUrl, data interface{}) RequestConfig {
	return RequestConfig{
		rootPassword: rootPasswd,
		host:         host,
		xenonHttpUrl: xenonHttpUrl,
		data:         data,
	}
}

// NewXenonHttpRequest returns the http request of the corresponding type according to the incoming configuration.
func NewXenonHttpRequest(cfg RequestConfig) (*Request, error) {
	requestType, ok := utils.XenonHttpUrls[cfg.xenonHttpUrl]
	if !ok {
		return nil, fmt.Errorf("xenon http url[%s] does not exist or does not support", cfg.xenonHttpUrl)
	}

	switch requestType {
	case http.MethodGet:
		return newHttpGetRequest(cfg)
	case http.MethodPost:
		return newHttpPostRequest(cfg)
	default:
		return nil, fmt.Errorf("request type[%s] does not support", requestType)
	}
}

func newHttpGetRequest(cfg RequestConfig) (*Request, error) {
	req, err := http.NewRequest(http.MethodGet, newXenonRequestUrl(cfg.host, cfg.xenonHttpUrl), nil)
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(append([]byte(xenonHttpUser), cfg.rootPassword...))
	req.Header.Set("Authorization", "Basic "+encoded)
	return &Request{Req: req}, nil
}

func newHttpPostRequest(cfg RequestConfig) (*Request, error) {
	reqBody := ""
	if cfg.data != nil {
		reqBody = cfg.data.(string)
	}

	req, err := http.NewRequest(http.MethodPost, newXenonRequestUrl(cfg.host, cfg.xenonHttpUrl), bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	encoded := base64.StdEncoding.EncodeToString(append([]byte(xenonHttpUser), cfg.rootPassword...))
	req.Header.Set("Authorization", "Basic "+encoded)
	return &Request{Req: req}, nil
}

func newXenonRequestUrl(host string, xenonHttpUrl utils.XenonHttpUrl) string {
	return fmt.Sprintf("http://%s:%d%s", host, utils.XenonPeerPort, xenonHttpUrl)
}
