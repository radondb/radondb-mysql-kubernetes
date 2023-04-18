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

package api

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alhpa1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
)

func backup(ns string, clusterName string, hostName string, nfsAdd string, backupImage string) *v1alhpa1.Backup {
	return &v1alhpa1.Backup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Backup",
			APIVersion: "mysql.radondb.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backup-" + clusterName,
			Namespace: ns,
		},
		Spec: v1alhpa1.BackupSpec{
			Image:            backupImage,
			HostName:         hostName,
			NFSServerAddress: nfsAdd,
			ClusterName:      clusterName,
		},
	}
}

func getClusterLeader(c *v1alhpa1.MysqlCluster) (string, error) {
	var hostPath string
	for i := 0; i < len(c.Status.Nodes); i++ {
		if c.Status.Nodes[i].RaftStatus.Role == "LEADER" {
			hostPath = c.Status.Nodes[i].Name
			return hostPath, nil
		}
	}
	return "", fmt.Errorf("failed to gat cluster leader")
}
