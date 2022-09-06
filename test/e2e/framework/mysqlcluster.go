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
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	RadonDBMySQL57ClusterTemplate = `
apiVersion: mysql.radondb.com/v1alpha1
kind: MysqlCluster
metadata:
  name: %s
spec:
  replicas: %d
  xenonOpts:
    image: radondb/xenon:v2.2.1
`
	MySQL57ReleaseAssetURL = "https://github.com/radondb/radondb-mysql-kubernetes/releases/latest/download/mysql_v1alpha1_mysqlcluster.yaml"
)

func newCluster(name, ns string, replicas int32) *apiv1alpha1.MysqlCluster {
	return &apiv1alpha1.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: apiv1alpha1.MysqlClusterSpec{
			Replicas:     &replicas,
			MysqlVersion: TestContext.MysqlVersion,
			PodPolicy: apiv1alpha1.PodPolicy{
				SidecarImage: TestContext.SidecarImagePath,
			},
		},
	}
}

func (f *Framework) getExistCluster() *apiv1alpha1.MysqlCluster {
	existClusters := &apiv1alpha1.MysqlClusterList{}
	Expect(f.Client.List(context.TODO(), existClusters, &client.ListOptions{
		Namespace: TestContext.E2ETestNamespace,
	})).To(Succeed(), "failed to list clusters")

	if len(existClusters.Items) > 0 {
		return &existClusters.Items[0]
	}
	return &apiv1alpha1.MysqlCluster{}
}

// WaitClusterReadiness determine whether the cluster is ready.
func (f *Framework) WaitClusterReadiness(clusterKey *types.NamespacedName) {
	cluster := &apiv1alpha1.MysqlCluster{}
	Expect(f.Client.Get(context.TODO(), *clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", clusterKey.Name)

	timeout := f.Timeout
	if *cluster.Spec.Replicas > 0 {
		timeout = time.Duration(*cluster.Spec.Replicas) * f.Timeout
	}
	// Wait for pods to be ready.
	f.ClusterEventuallyReplicas(cluster, timeout)
	// Wait for xenon to be ready.
	f.ClusterEventuallyRaftStatus(cluster)
}

func (f *Framework) ClusterEventuallyReplicas(cluster *apiv1alpha1.MysqlCluster, timeout time.Duration) {
	Eventually(func() int {
		cl := &apiv1alpha1.MysqlCluster{}
		f.Client.Get(context.TODO(), types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, cl)
		f.Log.Logf(f.t, "ready nodes: %d/%d", cl.Status.ReadyNodes, *cl.Spec.Replicas)
		return cl.Status.ReadyNodes
	}, timeout, POLLING).Should(Equal(int(*cluster.Spec.Replicas)), "Not ready replicas of cluster '%s'", cluster.Name)
}

func (f *Framework) ClusterEventuallyRaftStatus(cluster *apiv1alpha1.MysqlCluster) {
	Eventually(func() bool {
		cl := &apiv1alpha1.MysqlCluster{}
		f.Client.Get(context.TODO(), types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, cl)
		f.Log.Logf(f.t, "checking xenon")
		return isXenonReadiness(cl)
	}, TIMEOUT, POLLING).Should(BeTrue(), "Not ready xenon of cluster '%s'", cluster.Name)
}

// isXenonReadiness determine whether the role of the cluster is normal.
// 1. Cluster must have Leader node.
// 2. Cluster can only have a Leader node.
func isXenonReadiness(cluster *apiv1alpha1.MysqlCluster) bool {
	leader := ""
	invalidCount := 0
	for _, node := range cluster.Status.Nodes {
		switch node.RaftStatus.Role {
		case string(utils.Leader):
			if leader != "" {
				return false
			}
			leader = node.Name
		case string(utils.Follower):
		default:
			invalidCount++
		}
	}

	return invalidCount == 0 && leader != ""
}

// GetClusterLabels returns labels.Set for the given cluster.
func GetClusterLabels(cluster *apiv1alpha1.MysqlCluster) labels.Set {
	labels := labels.Set{
		"mysql.radondb.com/cluster": cluster.Name,
		"app.kubernetes.io/name":    "mysql",
	}

	return labels
}

func (f *Framework) InstallMySQLClusterUsingTemplate() error {
	cluster := fmt.Sprintf(RadonDBMySQL57ClusterTemplate, TestContext.ClusterReleaseName, 2)
	if err := k8s.KubectlApplyFromStringE(f.t, f.kubectlOptions, cluster); err != nil {
		return err
	}
	return nil
}

func (f *Framework) InstallMySQLClusterUsingAsset() error {
	if err := k8s.KubectlApplyE(f.t, f.kubectlOptions, MySQL57ReleaseAssetURL); err != nil {
		return err
	}
	return nil
}
