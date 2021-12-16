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

package framework

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	api "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

type SysbenchOptions struct {
	Timeout   time.Duration
	Threads   int
	Tables    int
	TableSize int
}

func (o *SysbenchOptions) cmd(leaderHost string) string {
	return fmt.Sprintf(`sysbench \
		--db-driver=mysql \
		--mysql-host=%s \
		--mysql-port=3306 \
		--mysql-user=radondb_usr \
		--mysql-password=RadonDB@123 \
		--mysql-db=radondb \
		--report-interval=10 \
		--time=%d \
		--threads=%d \
		--tables=%d \
		--table_size=%d \
		/usr/share/sysbench/oltp_read_write.lua \
		`,
		leaderHost,
		int(o.Timeout.Seconds()),
		o.Threads,
		o.Tables,
		o.TableSize,
	)
}

func (f Framework) WaitPodReady(pod *corev1.Pod) {
	Eventually(func() bool {
		f.Client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pod)
		return pod.Status.ContainerStatuses[0].Ready
	}, f.Timeout, POLLING).Should(BeTrue(), "Not ready replicas of pod '%s'", pod.Name)
}

// func (f Framework) WaitDataReady(cluster *api.MysqlCluster, tables, tableSize int) {
// 	Eventually(f.TestDataReadiness(cluster, tables, tableSize), 2 * time.Minute, SQLPOLLING).Should(BeTrue(), "data not ready")
// }

func (f Framework) WaitPrepareFinished(pod *corev1.Pod) {
	Eventually(func() *corev1.ContainerStateTerminated {
		f.Client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pod)
		return pod.Status.ContainerStatuses[0].State.Terminated
	}, f.Timeout, POLLING).ShouldNot(BeNil(), "Not ready replicas of pod '%s'", pod.Name)
}

func (f Framework) PrepareData(mysqlcluster *api.MysqlCluster, options *SysbenchOptions) {
	name := "sysbench-prepare"
	cmd := options.cmd(fmt.Sprintf("%s-leader", mysqlcluster.Name))
	cmd = fmt.Sprintf("%s prepare", cmd)
	args := []string{cmd}

	By("create sysbench pod for preparing data")
	pod, err := f.createSysbenchPod(mysqlcluster.Namespace, name, args)
	Expect(err).Should(BeNil())

	By("wait pod ready")
	f.WaitPodReady(pod)
	Expect(f.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: mysqlcluster.Namespace}, pod)).To(Succeed(), "failed to prepare data")

	By("testing the data readiness")
	f.WaitPrepareFinished(pod)
}

func (f Framework) RunOltpTest(mysqlcluster *api.MysqlCluster, options *SysbenchOptions) {
	name := "sysbench-run"
	cmd := options.cmd(fmt.Sprintf("%s-leader", mysqlcluster.Name))
	cmd = fmt.Sprintf("%s run", cmd)
	args := []string{cmd}

	By("create sysbench pod for testing oltp")
	pod, err := f.createSysbenchPod(mysqlcluster.Namespace, name, args)
	Expect(err).Should(BeNil())

	By("wait pod ready")
	f.WaitPodReady(pod)
	Expect(f.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: mysqlcluster.Namespace}, pod)).To(Succeed(), "failed to run oltp test")
}

func (f Framework) createSysbenchPod(ns, name string, args []string) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	container := corev1.Container{
		Name:    "sysbench",
		Image:   "severalnines/sysbench",
		Command: []string{"/bin/bash", "-c", "--"},
		Args:    args,
	}

	pod.Spec.Containers = append(pod.Spec.Containers, container)
	pod.Spec.RestartPolicy = corev1.RestartPolicyNever

	var got *corev1.Pod
	if err := wait.PollImmediate(Poll, 2*time.Minute, func() (bool, error) {
		var err error
		got, err = f.ClientSet.CoreV1().Pods(ns).Create(context.TODO(), pod, metav1.CreateOptions{})
		if err != nil {
			Logf("Unexpected error while creating sysbench pod: %v", err)
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, err
	}
	return got, nil
}
