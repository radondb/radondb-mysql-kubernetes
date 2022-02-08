/*
Copyright 2015 The Kubernetes Authors.

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
	"fmt"
	"io"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/client-go/kubernetes"
)

func LogPodsWithLabels(c clientset.Interface, ns string, match map[string]string, since time.Duration, out io.Writer) {
	podList, err := c.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.SelectorFromSet(match).String()})
	if err != nil {
		fmt.Fprintf(out, "error listing pods: %s", err)
		return
	}

	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			fmt.Fprintf(out, "\n\n===============\nSTART LOGS for %s (%s):\n", pod.Name, container.Name)
			runLogs(c, ns, pod.Name, container.Name, false, since, out)
			fmt.Fprintf(out, "\n\n===============\nSTOP LOGS for %s (%s):\n", pod.Name, container.Name)
		}
	}
}

func runLogs(client clientset.Interface, namespace, name, container string, previous bool, sinceStart time.Duration, out io.Writer) error {
	req := client.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Name(name).
		Resource("pods").
		SubResource("log").
		Param("container", container).
		Param("previous", strconv.FormatBool(previous)).
		Param("since", strconv.FormatInt(int64(sinceStart.Round(time.Second).Seconds()), 10))

	readCloser, err := req.Stream(context.TODO())
	if err != nil {
		return err
	}

	defer readCloser.Close()
	_, err = io.Copy(out, readCloser)
	return err
}
