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
	"bufio"
	"bytes"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type PodExecutor struct {
	client corev1client.CoreV1Interface
	config *rest.Config
}

func NewPodExecutor() (*PodExecutor, error) {
	// Instantiate loader for kubeconfig file.
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	// Create a Kubernetes core/v1 client.
	client, err := corev1client.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &PodExecutor{
		client: client,
		config: config,
	}, nil
}

func (p *PodExecutor) Exec(namespace, podName, containerName string, command ...string) ([]byte, []byte, error) {
	request := p.client.RESTClient().
		Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   command,
			Stdout:    true,
			Stderr:    true,
			Stdin:     false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(p.config, "POST", request.URL())
	if err != nil {
		return nil, nil, err
	}

	stdOut := bytes.Buffer{}
	stdErr := bytes.Buffer{}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: bufio.NewWriter(&stdOut),
		Stderr: bufio.NewWriter(&stdErr),
		Stdin:  nil,
		Tty:    false,
	})

	return stdOut.Bytes(), stdErr.Bytes(), err
}

func (p *PodExecutor) SetGlobalSysVar(namespace, podName string, query string) error {
	cmd := []string{"xenoncli", "mysql", "sysvar", query}
	_, stderr, err := p.Exec(namespace, podName, "xenon", cmd...)
	if err != nil {
		return err
	}
	if len(stderr) != 0 {
		return fmt.Errorf("run command %s in xenon failed: %s", cmd, stderr)
	}
	return nil
}

func (p *PodExecutor) CloseXenonSemiCheck(namespace, podName string) error {
	cmd := []string{"xenoncli", "raft", "disablechecksemisync"}
	_, stderr, err := p.Exec(namespace, podName, "xenon", cmd...)
	if err != nil {
		return err
	}
	if len(stderr) != 0 {
		return fmt.Errorf("run command %s in xenon failed: %s", cmd, stderr)
	}
	return nil
}

func (p *PodExecutor) XenonTryLeader(namespace, podName string) error {
	cmd := []string{"xenoncli", "raft", "trytoleader"}
	_, stderr, err := p.Exec(namespace, podName, "xenon", cmd...)
	if err != nil {
		return err
	}
	if len(stderr) != 0 {
		return fmt.Errorf("run command %s in xenon failed: %s", cmd, stderr)
	}
	return nil
}
