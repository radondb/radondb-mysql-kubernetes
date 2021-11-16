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
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	Req *http.Request
}

type Response struct {
	Body io.ReadCloser
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpClient struct {
	client *http.Client
}

func NewHttpClient(client *http.Client) HttpClient {
	return &httpClient{
		client: client,
	}
}

func (cli *httpClient) Do(req *http.Request) (*http.Response, error) {
	return cli.client.Do(req)
}

type Executor interface {
	Execute(req *Request) (*Response, error)
}

type httpExecutor struct {
	Client HttpClient
}

func NewHttpExecutor(httpClient HttpClient) Executor {
	return &httpExecutor{
		Client: httpClient,
	}
}

func (executor *httpExecutor) Execute(req *Request) (*Response, error) {
	resp, err := executor.Client.Do(req.Req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code is %d", resp.StatusCode)
	}
	return &Response{
		Body: resp.Body,
	}, nil
}
