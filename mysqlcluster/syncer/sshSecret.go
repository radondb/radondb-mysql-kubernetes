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
	"github.com/presslabs/controller-util/pkg/syncer"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewSecretSyncer returns secret syncer.
func NewSShKeySyncer(cli client.Client, c *mysqlcluster.MysqlCluster) syncer.Interface {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.SShKey),
			Namespace: c.Namespace,
		},
	}

	return syncer.NewObjectSyncer("Secret", c.Unwrap(), secret, cli, func() error {

		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}

		if len(secret.Data["id_ecdsa"]) == 0 {
			pub, priv, err := GenSSHKey()
			if err != nil {
				return err
			}
			secret.Data["id_ecdsa"] = priv
			secret.Data["authorized_keys"] = pub

		}

		return nil
	})
}
