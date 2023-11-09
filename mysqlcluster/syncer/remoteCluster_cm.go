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

package syncer

import (
	"context"
	"fmt"
	"strings"

	"github.com/presslabs/controller-util/pkg/syncer"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewXenonCMSyncer returns xenon configmap syncer.
func NewRemoteClusterCMSyncer(cli client.Client, c *mysqlcluster.MysqlCluster) syncer.Interface {

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.RemoteCluster),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
	}

	return syncer.NewObjectSyncer("ConfigMap", c.Unwrap(), cm, cli, func() error {
		if int(*c.Spec.Replicas) == 0 {
			return nil
		}

		if c.Spec.RemoteCluster == nil {
			return nil
		}

		metaData, err := buildRemoteClusterMetaData(cli, c)
		if err != nil {
			return fmt.Errorf("failed to build remote cluster metadata: %s", err)
		}

		cm.Data = map[string]string{
			"RemoteCluster.sh": metaData,
		}

		return nil
	})
}

func buildRemoteClusterMetaData(cli client.Client, c *mysqlcluster.MysqlCluster) (string, error) {
	// get secret for remote cluster
	secretName := fmt.Sprintf("%s-secret", c.Spec.RemoteCluster.Name)
	// use secretName to get secret
	secret := &corev1.Secret{}
	if err := cli.Get(context.TODO(), client.ObjectKey{Namespace: c.Spec.RemoteCluster.NameSpace, Name: secretName}, secret); err != nil {
		return "", fmt.Errorf("failed to get secret: %s", err)
	}

	// generate the shell script from secret information
	metaData := "#!/bin/bash\n"
	// export Backup password to environment variable
	metaData += fmt.Sprintf("export BACKUP_USER=%s\n", string(secret.Data["backup-user"]))
	metaData += fmt.Sprintf("export BACKUP_PASSWORD=%s\n", string(secret.Data["backup-password"]))
	// use curl to download remote cluster backup script
	serviceURL := fmt.Sprintf("http://%s-leader.%s:%v",
		c.Spec.RemoteCluster.Name, c.Spec.RemoteCluster.NameSpace, utils.XBackupPort)
	metaData += fmt.Sprintf("curl --user $BACKUP_USER:$BACKUP_PASSWORD %s/download|xbstream -x -C %s\n",
		serviceURL, utils.DataVolumeMountPath)
	metaData += strings.Join([]string{"xtrabackup", "--defaults-file=" + utils.MysqlConfVolumeMountPath + "/my.cnf", "--use-memory=3072M", "--prepare", "--apply-log-only", "--target-dir=" + utils.DataVolumeMountPath}, " ")
	metaData += "\n"
	metaData += strings.Join([]string{"xtrabackup", "--defaultsπ-file=" + utils.MysqlConfVolumeMountPath + "/my.cnf", "--use-memory=3072M", "--prepare", "--target-dir=" + utils.DataVolumeMountPath}, " ")
	metaData += "\nchown -R mysql.mysql " + utils.DataVolumeMountPath + "\n"
	return metaData, nil
}
