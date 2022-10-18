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
	"encoding/json"
	"fmt"

	"github.com/presslabs/controller-util/pkg/syncer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// NewXenonCMSyncer returns xenon configmap syncer.
func NewXenonCMSyncer(cli client.Client, c *mysqlcluster.MysqlCluster) syncer.Interface {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.XenonMetaData),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
	}

	return syncer.NewObjectSyncer("ConfigMap", c.Unwrap(), cm, cli, func() error {
		if int(*c.Spec.Replicas) == 0 {
			return nil
		}

		metaData, err := buildXenonMetaData(c)
		if err != nil {
			return fmt.Errorf("failed to build xenon metadata: %s", err)
		}

		cm.Data = map[string]string{
			"peers.json": metaData,
		}

		return nil
	})
}

type XenonMetaData struct {
	Idlepeers []string `json:"idlepeers"`
	Peers     []string `json:"peers"`
}

// buildXenonMetaData build the default metadata of xenon.
func buildXenonMetaData(c *mysqlcluster.MysqlCluster) (string, error) {
	replicas := c.Spec.Replicas
	xenonMetaData := XenonMetaData{}
	for i := 0; i < int(*replicas); i++ {
		xenonMetaData.Peers = append(xenonMetaData.Peers,
			fmt.Sprintf("%s-%d.%s.%s:%d",
				c.GetNameForResource(utils.StatefulSet),
				i,
				c.GetNameForResource(utils.HeadlessSVC),
				c.Namespace,
				utils.XenonPort,
			))
	}
	metaJson, err := json.Marshal(xenonMetaData)
	if err != nil {
		return "", err
	}
	return string(metaJson), nil
}
