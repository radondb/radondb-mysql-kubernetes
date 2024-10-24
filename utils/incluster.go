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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func UpdateforCRD(crdName string, cli client.Client, log *logr.Logger) error {
	// TODO: update CRD

	// MYNS=extension-dmp
	// CRD1=mysqlclusters.mysql.radondb.com
	// CRD2=backups.mysql.radondb.com
	// SEC=radondb-mysql-webhook-certs
	// CERT=$(kubectl -n $MYNS get secrets $SEC -ojsonpath='{.data.tls\.crt}')
	// kubectl patch CustomResourceDefinition $CRD1 --type=merge -p '{"spec":{"conversion":{"webhook":{"clientConfig":{"caBundle":"'$CERT'","service":{"namespace":"'$MYNS'"}}}}}}'
	// kubectl patch CustomResourceDefinition $CRD2 --type=merge -p '{"spec":{"conversion":{"webhook":{"clientConfig":{"caBundle":"'$CERT'","service":{"namespace":"'$MYNS'"}}}}}}'
	// echo $CERT
	// fetch a secret in dmp-extension namespace which is named radondb-mysql-webhook-certs
	// 1. first get os environment value MY_NAMESPACE, if not set, use default namespace ,"dmp-extension"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ns := os.Getenv("MY_NAMESPACE")
	if len(ns) == 0 {
		ns = "extension-dmp"
	}
	//2. get os environment value CERT_NAME, if not set, use default namespace "radondb-mysql-webhook-certs"
	certName := os.Getenv("CERT_NAME")
	if len(certName) == 0 {
		certName = "radondb-mysql-webhook-certs"
	}
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: ns,
		},
	}
	err := cli.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, secret)
	// if err is not found, return error
	if errors.IsNotFound(err) {
		return fmt.Errorf("secret %s not found", certName)
	}
	cert := secret.Data["tls.crt"]

	//fetch the  CustomResourceDefinition ,which name is mysqlclusters.mysql.radondb.com
	// 创建 CRD 实例
	//apiextensionsv1.AddToScheme(scheme)
	//CustomResourceDefinition
	crd := &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
	}
	// 使用客户端获取 CRD 资源
	errCRD := cli.Get(ctx, types.NamespacedName{Name: crdName}, crd)
	if errCRD != nil {
		return errCRD
	}
	hasBetaVersion := false
	for _, v := range crd.Spec.Versions {
		if v.Name == "v1beta1" {
			hasBetaVersion = true
		}
	}
	if !hasBetaVersion {
		return fmt.Errorf("has not v1beta1 version")
	}

	oldCrd := crd.DeepCopy()
	// if CustomResourceConversion's CABundle of Webhook is not equal to cert, update it
	// sometime the path and port are missing, I don't know why
	if oldCrd.Spec.Conversion == nil || oldCrd.Spec.Conversion.Webhook == nil || oldCrd.Spec.Conversion.Webhook.ClientConfig == nil ||
		oldCrd.Spec.Conversion.Webhook.ClientConfig.CABundle == nil ||
		oldCrd.Spec.Conversion.Webhook.ClientConfig.Service.Path == nil ||
		oldCrd.Spec.Conversion.Webhook.ClientConfig.Service.Port == nil ||
		!bytes.Equal(oldCrd.Spec.Conversion.Webhook.ClientConfig.CABundle, cert) {
		crd.Spec.Conversion = &apiextensionsv1.CustomResourceConversion{
			Strategy: apiextensionsv1.WebhookConverter,
			Webhook: &apiextensionsv1.WebhookConversion{
				ClientConfig: &apiextensionsv1.WebhookClientConfig{
					CABundle: []byte(cert),
					Service: &apiextensionsv1.ServiceReference{
						Namespace: ns,
						Name: func() string {
							if oldCrd.Spec.Conversion != nil && oldCrd.Spec.Conversion.Webhook != nil && oldCrd.Spec.Conversion.Webhook.ClientConfig != nil {
								return oldCrd.Spec.Conversion.Webhook.ClientConfig.Service.Name
							} else {
								return "radondb-mysql-webhook"
							}
						}(),
						Path: func() *string {
							var p string = "/convert"
							return &p
						}(),
						Port: func() *int32 {
							var serverPort int32 = 443
							return &serverPort
						}(),
					},
				},
				ConversionReviewVersions: []string{"v1"},
			},
		}
		log.Info("covert crd", "value", crd.Spec.Conversion)
	} else {
		return nil
	}
	errCRD = cli.Patch(ctx, crd, client.MergeFrom(oldCrd))
	if errCRD != nil {
		return errCRD
	}

	return nil
}

func RunUpdeteCRD(cli client.Client, log *logr.Logger) {
	go func() {
		// Just run in the first 500 seconds,almost eight minutes, because the crd webhook's CABundle is not correct just in the DMP reinstall period
		// if this process failed, just need to restart the operetor pod
		for i := 0; i < 100; i++ {
			time.Sleep(time.Second * 5)
			err := UpdateforCRD("mysqlclusters.mysql.radondb.com", cli, log)
			if err != nil {
				log.Info("update CRD failed", "error", err)
			}
			err = UpdateforCRD("backups.mysql.radondb.com", cli, log)
			if err != nil {
				log.Info("update CRD failed", "error", err)
			}
		}
		log.Info("check the crd about 8 minutes, now exit.")
	}()
}
