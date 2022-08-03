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
	"net/http"

	// apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

type xenonExecutor struct {
	httpExecutor
	rootPassword string
}

type XenonExecutor interface {
	GetRootPassword() string
	SetRootPassword(rootPassword string)
	RaftStatus(host string) (*RaftStatus, error)
	XenonPing(host string) error
	RaftTryToLeader(host string) error
	ClusterAdd(host string, toAdd string) error
	ClusterRemove(host string, toRemove string) error
}

func NewXenonExecutor() XenonExecutor {
	return &xenonExecutor{httpExecutor: httpExecutor{Client: NewHttpClient(&http.Client{})}}
}

func (executor *xenonExecutor) GetRootPassword() string {
	return executor.rootPassword
}

func (executor *xenonExecutor) SetRootPassword(rootPassword string) {
	executor.rootPassword = rootPassword
}

// RaftStatus gets the raft status of incoming host through http.
func (executor *xenonExecutor) RaftStatus(host string) (*RaftStatus, error) {
	req, err := NewXenonHttpRequest(NewRequestConfig(host, executor.rootPassword, utils.RaftStatus, nil))
	if err != nil {
		return nil, err
	}

	response, err := executor.httpExecutor.Execute(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get raft status, err: %s", err)
	}

	var out map[string]interface{}
	if err = utils.UnmarshalJSON(response.Body, &out); err != nil {
		return nil, err
	}

	nodesJson := out["nodes"].([]interface{})
	nodes := []string{}
	for _, node := range nodesJson {
		nodes = append(nodes, node.(string))
	}
	return &RaftStatus{State: out["state"].(string), Leader: out["leader"].(string), Nodes: nodes}, nil
}

// RaftTryToLeader try setting up incoming host to the leader node.
func (executor *xenonExecutor) RaftTryToLeader(host string) error {
	req, err := NewXenonHttpRequest(NewRequestConfig(host, executor.rootPassword, utils.RaftTryToLeader, nil))
	if err != nil {
		return err
	}

	_, err = executor.httpExecutor.Execute(req)
	if err != nil {
		return fmt.Errorf("failed to execute raft/trytoleader at host[%s], err: %s", req.Req.URL, err)
	}
	return nil
}

func (executor *xenonExecutor) XenonPing(host string) error {
	req, err := NewXenonHttpRequest(NewRequestConfig(host, executor.GetRootPassword(), utils.XenonPing, nil))
	if err != nil {
		return err
	}
	_, err = executor.httpExecutor.Execute(req)
	if err != nil {
		return fmt.Errorf("failed to ping host[%s], err: %s", req.Req.URL, err)
	}
	return nil
}

func (executor *xenonExecutor) ClusterAdd(host string, toAdd string) error {
	addHost := fmt.Sprintf("{\"address\": \"%s\"}", toAdd)
	req, err := NewXenonHttpRequest(NewRequestConfig(host, executor.GetRootPassword(), utils.ClusterAdd, addHost))
	if err != nil {
		return err
	}
	_, err = executor.httpExecutor.Execute(req)
	if err != nil {
		return fmt.Errorf("failed to add host[%s] to host[%s], err: %s", addHost, req.Req.URL, err)
	}
	return nil
}

func (executor *xenonExecutor) ClusterRemove(host string, toRemove string) error {
	removeHost := fmt.Sprintf("{\"address\": \"%s\"}", toRemove)
	req, err := NewXenonHttpRequest(NewRequestConfig(host, executor.GetRootPassword(), utils.ClusterRemove, removeHost))
	if err != nil {
		return err
	}
	_, err = executor.httpExecutor.Execute(req)
	if err != nil {
		return fmt.Errorf("failed to remove host[%s] from host[%s], err: %s", removeHost, req.Req.URL, err)
	}
	return nil
}
