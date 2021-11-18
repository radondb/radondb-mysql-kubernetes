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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	k8score "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	pf "github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework/portforward"
)

var (
	POLLING = 2 * time.Second
	// SQLPOLLING = 10 * time.Second
)

func (f *Framework) ClusterEventuallyCondition(cluster *api.MysqlCluster,
	condType api.ClusterConditionType, status corev1.ConditionStatus, timeout time.Duration) {
	Eventually(func() []api.ClusterCondition {
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

func (f *Framework) NodeEventuallyCondition(cluster *api.MysqlCluster, nodeName string,
	condType api.NodeConditionType, status corev1.ConditionStatus, timeout time.Duration) {
	Eventually(func() []api.NodeCondition {
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

// HaveClusterCond is a helper func that returns a matcher to check for an existing condition in a ClusterCondition list.
func HaveClusterCond(condType api.ClusterConditionType, status corev1.ConditionStatus) gomegatypes.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"Status": MatchFields(IgnoreExtras, Fields{
			"Conditions": ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(condType),
				"Status": Equal(status),
			})),
		})},
	))
}

func (f *Framework) RefreshClusterFn(cluster *api.MysqlCluster) func() *api.MysqlCluster {
	return func() *api.MysqlCluster {
		key := types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		}
		c := &api.MysqlCluster{}
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
func GetClusterLabels(cluster *api.MysqlCluster) labels.Set {
	labels := labels.Set{
		"mysql.radondb.com/cluster": cluster.Name,
		"app.kubernetes.io/name":    "mysql",
	}

	return labels
}

func (f *Framework) GetClusterPVCsFn(cluster *api.MysqlCluster) func() []corev1.PersistentVolumeClaim {
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

// func (f *Framework) TestDataReadiness(cluster *api.MysqlCluster, tables, tableSize int) bool {
// 	count := 0
// 	podName := fmt.Sprintf("%s-mysql-0", cluster.Name)
// 	rows, err := f.ExecSQLOnNode(*cluster, podName, fmt.Sprintf("select count(*) from sbtest%d;", tables))
// 	if err != nil {
// 		return false
// 	}
// 	defer rows.Close()

// 	if rows.Next() {
// 		rows.Scan(&count)
// 	}
// 	if count == tableSize {
// 		return true
// 	}
// 	return false
// }

func (f *Framework) ExecSQLOnNode(cluster api.MysqlCluster, podName, query string) (*sql.Rows, error) {
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
