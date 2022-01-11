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

package framework

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
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

func (f *Framework) InitAClusterForTesting(replicas int32) string {
	// Be careful, mysql allowed hostname lenght is <63.
	cluster := newCluster(fmt.Sprintf("cl-%d", rand.Int31()/1000), f.Namespace.Name, replicas)
	if err := f.Client.Create(context.TODO(), cluster); err != nil {
		return ""
	}
	return cluster.Name
}

func (f *Framework) getExistCluster() *apiv1alpha1.MysqlCluster {
	existClusters := &apiv1alpha1.MysqlClusterList{}
	Expect(f.Client.List(context.TODO(), existClusters, &client.ListOptions{
		Namespace: RadondbMysqlE2eNamespace,
	})).To(Succeed(), "failed to list clusters")

	if len(existClusters.Items) > 0 {
		return &existClusters.Items[0]
	}
	return &apiv1alpha1.MysqlCluster{}
}

func (f *Framework) InitOrGetCluster(replicas int32) *types.NamespacedName {
	clusterKey := &types.NamespacedName{Namespace: RadondbMysqlE2eNamespace}
	if cluster := f.getExistCluster(); cluster.Name != "" {
		clusterKey.Name = cluster.Name
		cluster.Spec.Replicas = &replicas
		Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed(), "failed to update clusters")
	} else {
		clusterKey.Name = f.InitAClusterForTesting(replicas)
	}
	return clusterKey
}

func (f *Framework) UpdateClusterReplicas(cluster *apiv1alpha1.MysqlCluster, replicas int32) {
	Expect(f.Client.Get(context.TODO(), client.ObjectKeyFromObject(cluster), cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)
	cluster.Spec.Replicas = &replicas
	Expect(f.Client.Update(context.TODO(), cluster)).To(Succeed())
}

// WaitClusterReadiness determine whether the cluster is ready.
func (f *Framework) WaitClusterReadiness(clusterKey *types.NamespacedName) {
	cluster := &apiv1alpha1.MysqlCluster{}
	if err := f.Client.Get(context.TODO(), *clusterKey, cluster); err != nil {
		Failf(fmt.Sprintf("Failed to get cluster %s", clusterKey.String()))
	}
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
		return cl.Status.ReadyNodes
	}, timeout, POLLING).Should(Equal(int(*cluster.Spec.Replicas)), "Not ready replicas of cluster '%s'", cluster.Name)
}

func (f *Framework) ClusterEventuallyRaftStatus(cluster *apiv1alpha1.MysqlCluster) {
	Eventually(func() bool {
		cl := &apiv1alpha1.MysqlCluster{}
		f.Client.Get(context.TODO(), types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, cl)
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

func (f *Framework) GetClusterPVCsFn(cluster *apiv1alpha1.MysqlCluster) func() []corev1.PersistentVolumeClaim {
	return func() []corev1.PersistentVolumeClaim {
		pvcList := &corev1.PersistentVolumeClaimList{}
		lo := &client.ListOptions{
			Namespace:     cluster.Namespace,
			LabelSelector: labels.SelectorFromSet(GetClusterLabels(cluster)),
		}
		f.Client.List(context.TODO(), pvcList, lo)
		return pvcList.Items
	}
}

func (f *Framework) GetClusterPods(cluster *apiv1alpha1.MysqlCluster) func() []corev1.Pod {
	return func() []corev1.Pod {
		podList := &corev1.PodList{}
		lo := &client.ListOptions{
			Namespace:     cluster.Namespace,
			LabelSelector: labels.SelectorFromSet(GetClusterLabels(cluster)),
		}
		f.Client.List(context.TODO(), podList, lo)
		return podList.Items
	}
}
