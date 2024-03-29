/*
Copyright 2022 RadonDB.

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

package framework

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	// "github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework/ginkgowrapper"
)

// CreateTestingNS should be used by every test, note that we append a common prefix to the provided test name.
// Please see NewFramework instead of using this directly.
func CreateTestingNS(baseName string, c clientset.Interface, labels map[string]string) (*corev1.Namespace, error) {
	if labels == nil {
		labels = map[string]string{}
	}
	namespaceObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			// use a short name because long names produce long hostnames but
			// maximum allowed length by mysql is 60.
			// https://dev.mysql.com/doc/refman/8.0/en/change-master-to.html
			GenerateName: fmt.Sprintf("e2e-%v-", baseName),
			Namespace:    "",
			Labels:       labels,
		},
		Status: corev1.NamespaceStatus{},
	}
	// Be robust about making the namespace creation call.
	var got *corev1.Namespace
	if err := wait.PollImmediate(Poll, 30*time.Second, func() (bool, error) {
		var err error
		got, err = c.CoreV1().Namespaces().Create(context.TODO(), namespaceObj, metav1.CreateOptions{})
		if err != nil {
			Logf("Unexpected error while creating namespace: %v", err)
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, err
	}

	return got, nil
}

func RestclientConfig(kubeContext string) (*clientcmdapi.Config, error) {
	if TestContext.KubeConfig == "" {
		return nil, fmt.Errorf("KubeConfig must be specified to load client config")
	}
	c, err := clientcmd.LoadFromFile(TestContext.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error loading KubeConfig: %v", err.Error())
	}
	if kubeContext != "" {
		c.CurrentContext = kubeContext
	}
	return c, nil
}

func LoadConfig() (*restclient.Config, error) {
	c, err := RestclientConfig(TestContext.KubeContext)
	if err != nil {
		if TestContext.KubeConfig == "" {
			return restclient.InClusterConfig()
		} else {
			return nil, err
		}
	}

	return clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: TestContext.KubeHost}}).ClientConfig()
}

func waitTimeoutForPodReadyInNamespace(c clientset.Interface, podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, podRunningAndReady(c, podName, namespace))
}

func podRunningAndReady(c clientset.Interface, podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return podRunningAndReadyByPhase(*pod)
	}
}

func podRunningAndReadyByPhase(pod corev1.Pod) (bool, error) {
	switch pod.Status.Phase {
	case corev1.PodFailed, corev1.PodSucceeded:
		return false, errors.New("pod completed")
	case corev1.PodRunning:
		for _, cond := range pod.Status.Conditions {
			if cond.Type != corev1.PodReady {
				continue
			}
			return cond.Status == corev1.ConditionTrue, nil
		}
		return false, errors.New("pod ready condition not found")
	}
	return false, nil
}

// DeleteNS deletes the provided namespace, waits for it to be completely deleted, and then checks
// whether there are any pods remaining in a non-terminating state.
func DeleteNS(c clientset.Interface, namespace string, timeout time.Duration) error {
	var zero = int64(0)
	if err := c.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{GracePeriodSeconds: &zero}); err != nil {
		return err
	}

	// Wait for namespace to delete or timeout.
	if err := wait.PollImmediate(2*time.Second, timeout, func() (bool, error) {
		if _, err := c.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{}); err != nil {
			if strings.Contains(err.Error(), "not found") {
				return true, nil
			}
			Logf("Error while waiting for namespace to be terminated: %v", err)
			return false, nil
		}
		return false, nil
	}); err != nil {
		return err
	}
	return nil
}

func Logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	Log.Info(msg)
}

func GetPodLogs(c clientset.Interface, namespace, podName, containerName string) (string, error) {
	return getPodLogsInternal(c, namespace, podName, containerName, false)
}

func getPreviousPodLogs(c clientset.Interface, namespace, podName, containerName string) (string, error) {
	return getPodLogsInternal(c, namespace, podName, containerName, true)
}

// getPodLogsInternal is a utility function for gomega Eventually.
func getPodLogsInternal(c clientset.Interface, namespace, podName, containerName string, previous bool) (string, error) {
	logs, err := c.CoreV1().RESTClient().Get().
		Resource("pods").
		Namespace(namespace).
		Name(podName).SubResource("log").
		Param("container", containerName).
		Param("previous", strconv.FormatBool(previous)).
		Do(context.TODO()).
		Raw()
	if err != nil {
		return "", err
	}
	if err == nil && strings.Contains(string(logs), "Internal Error") {
		return "", fmt.Errorf("fetched log contains \"Internal Error\": %q", string(logs))
	}
	return string(logs), err
}

func kubectlLogPod(c clientset.Interface, pod corev1.Pod, containerNameSubstr string, logFunc func(ftm string, args ...interface{})) {
	for _, container := range pod.Spec.Containers {
		if strings.Contains(container.Name, containerNameSubstr) {
			// Contains() matches all strings if substr is empty
			logs, err := GetPodLogs(c, pod.Namespace, pod.Name, container.Name)
			if err != nil {
				logFunc("Failed to get logs of pod %v, container %v, err: %v", pod.Name, container.Name, err)
			}
			plogs, err := getPreviousPodLogs(c, pod.Namespace, pod.Name, container.Name)
			plogs = "PREVIOUS\n" + plogs
			if err != nil {
				plogs = fmt.Sprintf("Failed to get previous logs for pod %v, container %v, err: %v", pod.Name, container.Name, err)
			}
			logFunc("Logs of %v/%v:%v on node %v", pod.Namespace, pod.Name, container.Name, pod.Spec.NodeName)
			logFunc("%s : %s \nSTARTLOG\n%s\nENDLOG for container %v:%v:%v", plogs, containerNameSubstr, logs, pod.Namespace, pod.Name, container.Name)
		}
	}
}

func LogContainersInPodsWithLabels(c clientset.Interface, ns string, match map[string]string, containerSubstr string, logFunc func(ftm string, args ...interface{})) {
	podList, err := c.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.SelectorFromSet(match).String()})
	if err != nil {
		Logf("Error getting pods in namespace %q: %v", ns, err)
		return
	}
	for _, pod := range podList.Items {
		kubectlLogPod(c, pod, containerSubstr, logFunc)
	}
}

func (f *Framework) CreateNS() {
	k8s.CreateNamespace(f.t, f.kubectlOptions, TestContext.E2ETestNamespace)
}

func (f *Framework) CheckServiceEndpoint(name string, port int, path string) error {
	k8s.WaitUntilServiceAvailable(f.t, f.kubectlOptions, name, 10, 2*time.Second)
	svc := k8s.GetService(f.t, f.kubectlOptions, name)
	endpoint := k8s.GetServiceEndpoint(f.t, f.kubectlOptions, svc, port)
	// Setup a TLS configuration to submit with the helper, a blank struct is acceptable
	tlsConfig := tls.Config{}
	url := fmt.Sprintf("http://%s", endpoint)
	if path != "" {
		url = fmt.Sprintf("%s/%s", url, path)
	}

	http_helper.HttpGetWithRetryWithCustomValidation(
		f.t,
		url,
		&tlsConfig,
		30,
		10*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)
	return nil
}

func (f *Framework) WaitUntilServiceAvailable(serviceName string) {
	k8s.WaitUntilServiceAvailable(f.t, f.kubectlOptions, serviceName, 10, 2*time.Second)
}
