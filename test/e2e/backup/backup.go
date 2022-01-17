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
package backup

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("MySQL Backup E2E Tests", func() {
	f := framework.NewFramework("mcbackup-1")
	_ = f

	var (
		cluster      *apiv1alpha1.MysqlCluster
		clusterKey   types.NamespacedName
		name         string
		backupSecret *corev1.Secret
		//timeout      time.Duration
		POLLING   = 2 * time.Second
		backupDir string
		leader    string
		follower  string
	)

	BeforeEach(func() {
		// be careful, mysql allowed hostname lenght is <63
		name = fmt.Sprintf("bk-%d", rand.Int31()/1000)

		//timeout = 350 * time.Second

		By("create a new backup secret")
		backupSecret = f.NewBackupSecret()
		Expect(f.Client.Create(context.TODO(), backupSecret)).To(Succeed(), "create backup secret failed")
		By("creating a new cluster")
		cluster = framework.NewCluster(name, f.Namespace.Name)
		clusterKey = types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}
		cluster.Spec.BackupSecretName = backupSecret.Name
		Expect(f.Client.Create(context.TODO(), cluster)).To(Succeed(),
			"failed to create cluster '%s'", cluster.Name)
		By("waiting the cluster readiness")
		framework.WaitClusterReadiness(f, cluster)
		//get leader
		for _, node := range cluster.Status.Nodes {
			if node.RaftStatus.Role == string(utils.Leader) {
				leader = strings.Split(node.Name, ".")[0]
			} else if node.RaftStatus.Role == string(utils.Follower) {
				follower = strings.Split(node.Name, ".")[0]
			}
		}

		Expect(f.Client.Get(context.TODO(), clusterKey, cluster)).To(Succeed(), "failed to get cluster %s", cluster.Name)

		Eventually(f.RefreshClusterFn(cluster), f.Timeout, POLLING).Should(
			framework.HaveClusterReplicas(2))
		Eventually(f.RefreshClusterFn(cluster), f.Timeout, POLLING).Should(
			framework.HaveClusterCond(apiv1alpha1.ConditionReady, corev1.ConditionTrue))

		// refresh cluster
		Expect(f.Client.Get(context.TODO(), clusterKey, cluster)).To(Succeed(),
			"failed to get cluster %s", cluster.Name)

	})

	It("backup to object store", func() {
		//exectute sql command in mysql pod
		By("executing insert data")
		_, err := f.ExecSQLOnNode(*cluster, leader, "create table testtable (id int)")
		Expect(err).To(BeNil())
		_, err = f.ExecSQLOnNode(*cluster, leader, "insert into testtable values (1),(2),(3)")
		Expect(err).To(BeNil())
		rows, err := f.ExecSQLOnNode(*cluster, leader, "select * from testtable")
		Expect(err).To(BeNil(), "failed to execute sql")
		defer rows.Close()
		var id int
		ids := make([]int, 0)
		for rows.Next() {
			if err := rows.Scan(&id); err != nil {
				Fail(err.Error())
			}
			ids = append(ids, id)
		}
		Expect(ids).To(Equal([]int{1, 2, 3}))
		Eventually(func() []int {
			rows, _ :=
				f.ExecSQLOnNode(*cluster, follower, "select * from testtable ")
			if rows == nil {
				return nil
			}
			defer rows.Close()
			var id int
			ids := make([]int, 0)
			for rows.Next() {
				if err := rows.Scan(&id); err != nil {
					Fail(err.Error())
				}
				ids = append(ids, id)
			}
			return ids
		}, f.Timeout, POLLING).Should(Equal([]int{1, 2, 3}))
		By("executing a backup ")
		// do the backup
		backup := framework.NewBackup(cluster, leader)
		Expect(f.Client.Create(context.TODO(), backup)).To(Succeed(),
			"failed to create backup '%s'", backup.Name)

		Eventually(f.RefreshBackupFn(backup), f.Timeout, POLLING).Should(
			framework.HaveBackupComplete())

		if str, err := framework.GetPodLogs(f.ClientSet, f.Namespace.Name, leader, "backup"); err == nil {
			r, _ := regexp.Compile("backup_[0-9]+")
			backupDir = r.FindString(str)
		}
		nameRestore := fmt.Sprintf("rs-%d", rand.Int31()/1000)
		By("creating a new cluster from backup")
		clusterRestore := framework.NewCluster(nameRestore, f.Namespace.Name)
		clusterKeyRestore := types.NamespacedName{Name: clusterRestore.Name, Namespace: clusterRestore.Namespace}
		clusterRestore.Spec.BackupSecretName = backupSecret.Name
		clusterRestore.Spec.RestoreFrom = backupDir
		Expect(f.Client.Create(context.TODO(), clusterRestore)).To(Succeed(),
			"failed to create clusterRestore '%s'", clusterRestore.Name)
		By("waiting the clusterRestore readiness")
		framework.WaitClusterReadiness(f, clusterRestore)
		Eventually(f.RefreshClusterFn(clusterRestore), f.Timeout, POLLING).Should(
			framework.HaveClusterReplicas(2))
		Eventually(f.RefreshClusterFn(clusterRestore), f.Timeout, POLLING).Should(
			framework.HaveClusterCond(apiv1alpha1.ConditionReady, corev1.ConditionTrue))
		Eventually(func() []int {
			rows, _ := f.ExecSQLOnNode(*clusterRestore, fmt.Sprintf("%s-mysql-0", nameRestore), "select * from testtable ")
			if rows == nil {
				return nil
			}
			defer rows.Close()
			var id int
			ids := make([]int, 0)
			for rows.Next() {
				if err := rows.Scan(&id); err != nil {
					Fail(err.Error())
				}
				ids = append(ids, id)
			}
			return ids
		}, f.Timeout, POLLING).Should(Equal([]int{1, 2, 3}))

		// refresh clusterRestore
		Expect(f.Client.Get(context.TODO(), clusterKeyRestore, clusterRestore)).To(Succeed(),
			"failed to get clusterRestore %s", clusterRestore.Name)
		Expect(f.Client.Delete(context.TODO(), clusterRestore)).To(Succeed(),
			"failed to delete clusterRestore '%s'", clusterRestore.Name)
	})

})
