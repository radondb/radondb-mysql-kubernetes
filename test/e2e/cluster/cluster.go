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

package cluster

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
)

const (
	POLLING = 2 * time.Second
)

var _ = Describe("MySQL Cluster E2E Tests", func() {
	f := framework.NewFramework("mc-1")
	two := int32(2)
	three := int32(3)
	five := int32(5)

	var (
		cluster    *apiv1alpha1.MysqlCluster
		clusterKey types.NamespacedName
		name       string
	)

	BeforeEach(func() {
		// Be careful, mysql allowed hostname lenght is <63.
		name = fmt.Sprintf("cl-%d", rand.Int31()/1000)

		By("creating a new cluster")
		cluster = framework.NewCluster(name, f.Namespace.Name)
		clusterKey = types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}
		Expect(f.Client.Create(context.TODO(), cluster)).To(Succeed(), "failed to create cluster '%s'", cluster.Name)

		By("testing the cluster readiness")
		framework.WaitClusterReadiness(f, cluster)
		Expect(f.Client.Get(context.TODO(), clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)
	})

	It("scale out/in a cluster, 2 -> 3 -> 5 -> 3 -> 2", func() {
		By("test cluster is ready after scale out 2 -> 3")
		cluster.Spec.Replicas = &three
		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
		fmt.Println("scale time: ", framework.WaitClusterReadiness(f, cluster))

		By("test cluster is ready after scale out 3 -> 5")
		cluster.Spec.Replicas = &five
		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
		fmt.Println("scale time: ", framework.WaitClusterReadiness(f, cluster))

		By("test cluster is ready after scale in 5 -> 3")
		cluster.Spec.Replicas = &three
		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
		fmt.Println("scale time: ", framework.WaitClusterReadiness(f, cluster))

		By("test cluster is ready after scale in 3 -> 2")
		cluster.Spec.Replicas = &two
		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
		fmt.Println("scale time: ", framework.WaitClusterReadiness(f, cluster))
	})

})
