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

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	// MySQLPidFile is the path of mysql pid file.
	MySQLPidFile  = "/var/run/mysqld/mysqld.pid"
	SleepFlagFile = "/var/lib/mysql/sleep-forever"
)

type KubeAPI struct {
	Client *kubernetes.Clientset
	Config *rest.Config
}

type RunRemoteCommandConfig struct {
	Container, Namespace, PodName string
}
type raftStatus struct {
	Leader string   `json:"leader"`
	State  string   `json:"state"`
	Nodes  []string `json:"nodes"`
}

type MySQLNode struct {
	PodName   string
	Namespace string
	Role      string
}

func GetClientSet() (*kubernetes.Clientset, error) {
	// Creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %v", err)
	}
	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}
	return clientset, nil
}

func PatchRoleLabelTo(n MySQLNode) error {
	// Creates the clientset
	clientset, err := GetClientSet()
	if err != nil {
		return fmt.Errorf("failed to create clientset: %v", err)
	}
	patch := fmt.Sprintf(`{"metadata":{"labels":{"role":"%s"}}}`, n.Role)
	_, err = clientset.CoreV1().Pods(n.Namespace).Patch(context.TODO(), n.PodName, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch pod role label: %v", err)
	}
	return nil
}

func XenonPingMyself() error {
	args := []string{"xenon", "ping"}
	cmd := exec.Command("xenoncli", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to exec xenoncli xenon ping: %v", err)
	}
	return nil
}

func GetRaftStatus() *raftStatus {
	args := []string{"raft", "status"}
	cmd := exec.Command("xenoncli", args...)
	res, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to exec xenoncli raft status: %v", err)
	}
	raftStatus := raftStatus{}
	if err := json.Unmarshal(res, &raftStatus); err != nil {
		log.Fatalf("failed to unmarshal raft status: %v", err)
	}
	return &raftStatus
}

func GetRole() string {
	return GetRaftStatus().State
}

func SleepFlag() bool {
	_, err := os.Stat(SleepFlagFile)
	return !os.IsNotExist(err)
}

func GetMySQLPid() (int, error) {
	d, err := ioutil.ReadFile(MySQLPidFile)
	if err != nil {
		return -1, fmt.Errorf("failed to read mysql pid file: %v", err)
	}
	pid, err := strconv.Atoi(string(bytes.TrimSpace(d)))
	if err != nil {
		return -1, fmt.Errorf("error parsing pid from %s: %s", MySQLPidFile, err)
	}
	return pid, nil
}

func IsMySQLRunning() bool {
	pid, err := GetMySQLPid()
	if err != nil {
		return false
	}
	// check if the process is running
	_, err = os.FindProcess(pid)
	return err == nil
}

func (k *KubeAPI) Exec(namespace, pod, container string, stdin io.Reader, command []string) (string, string, error) {
	var stdout, stderr bytes.Buffer

	var Scheme = runtime.NewScheme()
	if err := corev1.AddToScheme(Scheme); err != nil {
		log.Fatalf("failed to add to scheme: %v", err)
		return "", "", err
	}
	var ParameterCodec = runtime.NewParameterCodec(Scheme)

	request := k.Client.CoreV1().RESTClient().Post().
		Resource("pods").SubResource("exec").
		Namespace(namespace).Name(pod).
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     stdin != nil,
			Stdout:    true,
			Stderr:    true,
		}, ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.Config, "POST", request.URL())

	if err == nil {
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  stdin,
			Stdout: &stdout,
			Stderr: &stderr,
		})
	}

	return stdout.String(), stderr.String(), err
}

func RunRemoteCommand(kubeapi *KubeAPI, cfg RunRemoteCommandConfig, cmd []string) (string, string, error) {
	bashCmd := []string{"bash"}
	reader := strings.NewReader(strings.Join(cmd, " "))
	return kubeapi.Exec(cfg.Namespace, cfg.PodName, cfg.Container, reader, bashCmd)
}

func NewForConfig(config *rest.Config) (*KubeAPI, error) {
	var api KubeAPI
	var err error

	api.Config = config
	api.Client, err = kubernetes.NewForConfig(api.Config)

	return &api, err
}

func NewConfig() (*rest.Config, error) {
	// The default loading rules try to read from the files specified in the
	// environment or from the home directory.
	loader := clientcmd.NewDefaultClientConfigLoadingRules()

	// The deferred loader tries an in-cluster config if the default loading
	// rules produce no results.
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader, &clientcmd.ConfigOverrides{}).ClientConfig()
}
