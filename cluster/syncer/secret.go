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
	"fmt"

	"github.com/presslabs/controller-util/rand"
	"github.com/presslabs/controller-util/syncer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/radondb/radondb-mysql-kubernetes/cluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

const (
	// The length of the secret string.
	rStrLen = 12
)

// NewSecretSyncer returns secret syncer.
func NewSecretSyncer(cli client.Client, c *cluster.Cluster) syncer.Interface {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.Secret),
			Namespace: c.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Secret", c.Unwrap(), secret, cli, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}

		secret.Data["operator-user"] = []byte(utils.OperatorUser)
		if err := addRandomPassword(secret.Data, "operator-password"); err != nil {
			return err
		}

		secret.Data["metrics-user"] = []byte(utils.MetricsUser)
		if err := addRandomPassword(secret.Data, "metrics-password"); err != nil {
			return err
		}

		if c.Spec.MetricsOpts.Enabled {
			dataSource := fmt.Sprintf("%s:%s@(localhost:3306)/", utils.MetricsUser, utils.BytesToString(secret.Data["metrics-password"]))
			secret.Data["data-source"] = []byte(dataSource)
		}

		secret.Data["replication-user"] = []byte(utils.ReplicationUser)
		if err := addRandomPassword(secret.Data, "replication-password"); err != nil {
			return err
		}

		if c.Spec.MysqlOpts.RootHost != "127.0.0.1" && c.Spec.MysqlOpts.RootPassword == "" {
			if err := addRandomPassword(secret.Data, "root-password"); err != nil {
				return err
			}
		} else {
			secret.Data["root-password"] = []byte(c.Spec.MysqlOpts.RootPassword)
		}

		secret.Data["mysql-user"] = []byte(c.Spec.MysqlOpts.User)
		secret.Data["mysql-password"] = []byte(c.Spec.MysqlOpts.Password)
		secret.Data["mysql-database"] = []byte(c.Spec.MysqlOpts.Database)
		return nil
	})
}

// addRandomPassword checks if a key exists and if not registers a random string for that key
func addRandomPassword(data map[string][]byte, key string) error {
	if len(data[key]) == 0 {
		// NOTE: use only alpha-numeric string, this strings are used unescaped in MySQL queries.
		random, err := rand.AlphaNumericString(rStrLen)
		if err != nil {
			return err
		}
		data[key] = []byte(random)
	}
	return nil
}
