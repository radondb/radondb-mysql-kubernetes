/*
Copyright 2018 Pressinfra SRL

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

package cluster

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

const (
	POLLING = 2 * time.Second
)

var _ = Describe("MySQL Cluster E2E Tests", func() {
	f := framework.NewFramework("mc-1")
	two := int32(2)
	three := int32(3)
	five := int32(5)

	sysbenchOptions := framework.SysbenchOptions{
		Timeout:   10 * time.Minute,
		Threads:   8,
		Tables:    4,
		TableSize: 10000,
	}

	var (
		cluster    *api.MysqlCluster
		clusterKey types.NamespacedName
		name       string
	)

	BeforeEach(func() {
		// be careful, mysql allowed hostname lenght is <63
		name = fmt.Sprintf("cl-%d", rand.Int31()/1000)

		By("creating a new cluster")
		cluster = framework.NewCluster(name, f.Namespace.Name)
		clusterKey = types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}
		Expect(f.Client.Create(context.TODO(), cluster)).To(Succeed(), "failed to create cluster '%s'", cluster.Name)

		By("testing the cluster readiness")
		testClusterReadiness(f, cluster)

		Expect(f.Client.Get(context.TODO(), clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)

		f.PrepareData(cluster, &sysbenchOptions)
		// By("testing the data readiness")
		// f.WaitDataReady(cluster, sysbenchOptions.Tables, sysbenchOptions.TableSize)
	})

	It("scale out/in a cluster, 2 -> 3 -> 5 -> 3 -> 2", func() {
		f.Client.Get(context.TODO(), types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, cluster)

		// start oltp test before testing scale in/out
		f.RunOltpTest(cluster, &sysbenchOptions)

		By("test cluster is ready after scale out 2 -> 3")
		cluster.Spec.Replicas = &three
		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
		testClusterReadiness(f, cluster)
		testXenonReadiness(cluster)

		By("test cluster is ready after scale out 3 -> 5")
		cluster.Spec.Replicas = &five
		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
		testClusterReadiness(f, cluster)
		testXenonReadiness(cluster)

		By("test cluster is ready after scale in 5 -> 3")
		cluster.Spec.Replicas = &three
		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
		testClusterReadiness(f, cluster)
		testXenonReadiness(cluster)

		By("test cluster is ready after scale in 3 -> 2")
		cluster.Spec.Replicas = &two
		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
		testClusterReadiness(f, cluster)
		testXenonReadiness(cluster)
	})
})

// testClusterReadiness determine whether the cluster is ready.
func testClusterReadiness(f *framework.Framework, cluster *api.MysqlCluster) {
	timeout := f.Timeout
	if *cluster.Spec.Replicas > 0 {
		timeout = time.Duration(*cluster.Spec.Replicas) * f.Timeout
	}

	// wait for pods to be ready
	Eventually(func() int {
		cl := &api.MysqlCluster{}
		f.Client.Get(context.TODO(), types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, cl)
		return cl.Status.ReadyNodes
	}, timeout, POLLING).Should(Equal(int(*cluster.Spec.Replicas)), "Not ready replicas of cluster '%s'", cluster.Name)

	f.ClusterEventuallyCondition(cluster, api.ConditionReady, corev1.ConditionTrue, f.Timeout)
}

// testXenonReadiness determine whether the role of the cluster is normal.
func testXenonReadiness(cluster *api.MysqlCluster) {
	leader := []string{}
	follower := []string{}
	for _, node := range cluster.Status.Nodes {
		if node.RaftStatus.Role == string(utils.Leader) {
			leader = append(leader, node.Name)
		} else if node.RaftStatus.Role == string(utils.Follower) {
			follower = append(follower, node.Name)
		} else {
			Expect(node).Should(BeEmpty(), "some nodes not ready")
		}
	}
	Expect(len(leader)).Should(Equal(1), "cluster need have only one leader")
	Expect(len(follower)).Should(Equal(len(cluster.Status.Nodes)-len(leader)), "in addition to the leader node, the cluster has only follower nodes")
}
