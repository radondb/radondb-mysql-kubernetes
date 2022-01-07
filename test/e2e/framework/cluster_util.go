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
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gomegatypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	k8score "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	pf "github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework/portforward"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	POLLING         = 2 * time.Second
	TIMEOUT         = time.Minute
	FAILOVERPOLLING = 200 * time.Millisecond
	FAILOVERTIMEOUT = 2 * time.Minute
)

func (f *Framework) ClusterEventuallyCondition(cluster *apiv1alpha1.MysqlCluster,
	condType apiv1alpha1.ClusterConditionType, status corev1.ConditionStatus, timeout time.Duration) {
	Eventually(func() []apiv1alpha1.ClusterCondition {
		key := types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}
		if err := f.Client.Get(context.TODO(), key, cluster); err != nil {
			return nil
		}
		return cluster.Status.Conditions
	}, timeout, POLLING).Should(ContainElement(MatchFields(IgnoreExtras, Fields{
		"Type":   Equal(condType),
		"Status": Equal(status),
	})), "Testing cluster '%s' for condition %s to be %s", cluster.Name, condType, status)

}

func (f *Framework) NodeEventuallyCondition(cluster *apiv1alpha1.MysqlCluster, nodeName string,
	condType apiv1alpha1.NodeConditionType, status corev1.ConditionStatus, timeout time.Duration) {
	Eventually(func() []apiv1alpha1.NodeCondition {
		key := types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}
		if err := f.Client.Get(context.TODO(), key, cluster); err != nil {
			return nil
		}

		for _, ns := range cluster.Status.Nodes {
			if ns.Name == nodeName {
				return ns.Conditions
			}
		}

		return nil
	}, timeout, POLLING).Should(ContainElement(MatchFields(IgnoreExtras, Fields{
		"Type":   Equal(condType),
		"Status": Equal(status),
	})), "Testing node '%s' of the cluster '%s'", cluster.Name, nodeName)
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
// 3. There are only two roles of Leader and Follower in the cluster.
func isXenonReadiness(cluster *apiv1alpha1.MysqlCluster) bool {
	leader := []string{}
	follower := []string{}
	for _, node := range cluster.Status.Nodes {
		if node.RaftStatus.Role == string(utils.Leader) {
			leader = append(leader, node.Name)
		} else if node.RaftStatus.Role == string(utils.Follower) {
			follower = append(follower, node.Name)
		} else {
			return false
		}
	}
	if len(leader) != 1 {
		return false
	}
	if len(follower) != len(cluster.Status.Nodes)-len(leader) {
		return false
	}
	return true
}

// HaveClusterCond is a helper func that returns a matcher to check for an existing condition in a ClusterCondition list.
func HaveClusterCond(condType apiv1alpha1.ClusterConditionType, status corev1.ConditionStatus) gomegatypes.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"Status": MatchFields(IgnoreExtras, Fields{
			"Conditions": ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(condType),
				"Status": Equal(status),
			})),
		})},
	))
}

func (f *Framework) RefreshClusterFn(cluster *apiv1alpha1.MysqlCluster) func() *apiv1alpha1.MysqlCluster {
	return func() *apiv1alpha1.MysqlCluster {
		key := types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		}
		c := &apiv1alpha1.MysqlCluster{}
		f.Client.Get(context.TODO(), key, c)
		return c
	}
}

// HaveClusterRepliacs matcher for replicas
func HaveClusterReplicas(replicas int) gomegatypes.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"Status": MatchFields(IgnoreExtras, Fields{
			"ReadyNodes": Equal(replicas),
		}),
	}))
}

// GetClusterLabels returns labels.Set for the given cluster
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

func NewCluster(name, ns string) *apiv1alpha1.MysqlCluster {
	two := int32(2)
	return &apiv1alpha1.MysqlCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: apiv1alpha1.MysqlClusterSpec{
			Replicas: &two,
			PodPolicy: apiv1alpha1.PodPolicy{
				SidecarImage:    TestContext.SidecarImage,
				ImagePullPolicy: "Always",
			},
		},
	}
}

func (f *Framework) ExecSQLOnNode(cluster apiv1alpha1.MysqlCluster, podName, query string) (*sql.Rows, error) {
	kubeCfg, err := LoadConfig()
	Expect(err).NotTo(HaveOccurred())

	user := cluster.Spec.MysqlOpts.User
	password := cluster.Spec.MysqlOpts.Password

	client := k8score.NewForConfigOrDie(kubeCfg).RESTClient()
	tunnel := pf.NewTunnel(client, kubeCfg, cluster.Namespace,
		podName,
		3306,
	)
	defer tunnel.Close()

	err = tunnel.ForwardPort()
	Expect(err).NotTo(HaveOccurred(), "Failed setting up port-forarding for pod: %s", podName)

	dsn := fmt.Sprintf("%s:%s@tcp(localhost:%d)/radondb?timeout=20s&multiStatements=true", user, password, tunnel.Local)
	db, err := sql.Open("mysql", dsn)
	Expect(err).To(Succeed(), "Failed connection to mysql DSN: %s", dsn)
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("err: %s, query: %s", err, query)
	}
	return rows, nil
}

func (f Framework) IsPodExist(roleLabel map[string]string, cluster *apiv1alpha1.MysqlCluster) bool {
	lo := &client.ListOptions{
		Namespace:     cluster.Namespace,
		LabelSelector: labels.SelectorFromSet(GetClusterLabels(cluster)),
	}
	roleRequirement, err := labels.NewRequirement("role", selection.Equals, []string{roleLabel["role"]})
	if err != nil {
		fmt.Sprintln("failed to create roleRequirement")
		return false
	}
	lo.LabelSelector.Add(*roleRequirement)
	podList, err := f.ClientSet.CoreV1().Pods(cluster.Namespace).List(context.TODO(), *lo.AsListOptions())
	if err != nil {
		fmt.Sprintln("failed to get pod")
		return false
	}
	if len(podList.Items) > 0 {
		return true
	}
	return false
}

func (f Framework) WaitServiceAvailable(clusterKey types.NamespacedName, roleLabel map[string]string) {
	Eventually(func() bool {
		cluster := apiv1alpha1.MysqlCluster{}
		f.Client.Get(context.TODO(), clusterKey, &cluster)
		return f.IsPodExist(roleLabel, &cluster)
	}, FAILOVERTIMEOUT, FAILOVERPOLLING).Should(BeTrue(), "service is unavailable")
}

// WaitClusterReadiness determine whether the cluster is ready.
func WaitClusterReadiness(f *Framework, cluster *apiv1alpha1.MysqlCluster) time.Duration {
	startTime := time.Now()
	timeout := f.Timeout
	if *cluster.Spec.Replicas > 0 {
		timeout = time.Duration(*cluster.Spec.Replicas) * f.Timeout
	}
	// Wait for pods to be ready.
	f.ClusterEventuallyReplicas(cluster, timeout)
	// Wait for xenon to be ready.
	f.ClusterEventuallyRaftStatus(cluster)
	// Wait for condition to be ready.
	f.ClusterEventuallyCondition(cluster, apiv1alpha1.ConditionReady, corev1.ConditionTrue, f.Timeout)
	return time.Since(startTime)
}
