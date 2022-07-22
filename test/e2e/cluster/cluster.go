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

	. "github.com/onsi/ginkgo/v2"
	ginkgoTypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
)

var _ = Describe("MySQL Cluster E2E Tests", Label("Cluster"), func() {
	var (
		f                *framework.Framework
		cluster          *apiv1alpha1.MysqlCluster
		clusterKey       *types.NamespacedName
		two, three, five int32 = 2, 3, 5
	)

	BeforeEach(func() {
		// Init framework.
		if f == nil {
			By("Init framework")
			f = &framework.Framework{
				BaseName: "mysqlcluster-e2e",
				Log:      framework.Log,
			}
			f.BeforeEach()
		}
		Expect(f).ShouldNot(BeNil(), "failed to init framework")
	})

	ReportAfterEach(func(report SpecReport) {
		if report.State == ginkgoTypes.SpecStatePassed {
			GinkgoWriter.Printf("%s : %s\n", report.LeafNodeText, report.RunTime)
		} else {
			GinkgoWriter.Printf("%s : %s\n", report.LeafNodeText, report.State)
		}
	})

	// Run the full scale in/out test with label filter: Scale.
	// Run only scale out(2 -> 3 -> 5): Scale out.
	// Run only scale in(5 -> 3 -> 2): Scale in.
	When("Test cluster scale", Label("Scale"), Ordered, func() {
		// Init a cluster or get an exist cluster.
		BeforeAll(func() {
			clusterKey = f.InitOrGetCluster(two)
			cluster = &apiv1alpha1.MysqlCluster{}
			Expect(f.Client.Get(context.TODO(), *clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)

			By("Testing the cluster readiness")
			f.WaitClusterReadiness(clusterKey)
		})

		Context("Init", Label("Init"), func() {
			// BeforeAll will init a cluster or get an exist cluster.
			// this container need do no thing and just for record init time.
			Specify("Replicas: 0 -> 2", func() {
			})
		})

		Context("Scale out", Label("Scale out"), Ordered, func() {
			// Guarantee the initial replicas is 2.
			BeforeAll(func() {
				f.UpdateClusterReplicas(cluster, two)
				f.WaitClusterReadiness(clusterKey)
			})

			Specify("Replicas: 2 -> 3", func() {
				cluster.Spec.Replicas = &three
				f.UpdateClusterReplicas(cluster, three)
				f.WaitClusterReadiness(clusterKey)
			})

			Specify("Replicas: 3 -> 5", func() {
				f.UpdateClusterReplicas(cluster, five)
				f.WaitClusterReadiness(clusterKey)
			})
		})

		Context("Scale in", Label("Scale In"), Ordered, func() {
			// Guarantee the initial replicas is 5.
			BeforeAll(func() {
				f.UpdateClusterReplicas(cluster, five)
				f.WaitClusterReadiness(clusterKey)
			})

			Specify("Replicas: 5 -> 3", func() {
				f.UpdateClusterReplicas(cluster, three)
				f.WaitClusterReadiness(clusterKey)
			})
			Specify("Replicas: 3 -> 2", func() {
				f.UpdateClusterReplicas(cluster, two)
				f.WaitClusterReadiness(clusterKey)
			})
		})
	})
})
