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
	"bytes"
	"fmt"
	"sort"

	"github.com/go-ini/ini"
	"github.com/presslabs/controller-util/syncer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/radondb/radondb-mysql-kubernetes/cluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// NewConfigMapSyncer returns configmap syncer.
func NewConfigMapSyncer(cli client.Client, c *cluster.Cluster) syncer.Interface {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.ConfigMap),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
	}

	return syncer.NewObjectSyncer("ConfigMap", c.Unwrap(), cm, cli, func() error {
		data, err := buildMysqlConf(c)
		if err != nil {
			return fmt.Errorf("failed to create mysql configs: %s", err)
		}

		cm.Data = map[string]string{
			"my.cnf":          data,
			"leader-start.sh": buildLeaderStart(c),
			"leader-stop.sh":  buildLeaderStop(c),
		}

		return nil
	})
}

// buildMysqlConf build the mysql config.
func buildMysqlConf(c *cluster.Cluster) (string, error) {
	cfg := ini.Empty(ini.LoadOptions{IgnoreInlineComment: true})
	sec := cfg.Section("mysqld")

	c.EnsureMysqlConf()

	addKVConfigsToSection(sec, mysqlSysConfigs, mysqlCommonConfigs, mysqlStaticConfigs, c.Spec.MysqlOpts.MysqlConf)

	if c.Spec.MysqlOpts.InitTokuDB {
		addKVConfigsToSection(sec, mysqlTokudbConfigs)
	}

	for _, key := range mysqlBooleanConfigs {
		if _, err := sec.NewBooleanKey(key); err != nil {
			log.Error(err, "failed to add boolean key to config section", "key", key)
		}
	}

	data, err := writeConfigs(cfg)
	if err != nil {
		return "", err
	}

	return data, nil
}

// addKVConfigsToSection add a map[string]string to a ini.Section
func addKVConfigsToSection(s *ini.Section, extraMysqld ...map[string]string) {
	for _, extra := range extraMysqld {
		keys := []string{}
		for key := range extra {
			keys = append(keys, key)
		}

		// sort keys
		sort.Strings(keys)

		for _, k := range keys {
			value := extra[k]
			if _, err := s.NewKey(k, value); err != nil {
				log.Error(err, "failed to add key to config section", "key", k, "value", extra[k], "section", s)
			}
		}
	}
}

// writeConfigs write to string ini.File
// nolint: interfacer
func writeConfigs(cfg *ini.File) (string, error) {
	var buf bytes.Buffer
	if _, err := cfg.WriteTo(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// buildLeaderStart build the leader-start.sh.
func buildLeaderStart(c *cluster.Cluster) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
curl -X PATCH -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" -H "Content-Type: application/json-patch+json" \
--cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/namespaces/%s/pods/$HOSTNAME \
-d '[{"op": "replace", "path": "/metadata/labels/role", "value": "leader"}]'
`, c.Namespace)
}

// buildLeaderStop build the leader-stop.sh.
func buildLeaderStop(c *cluster.Cluster) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
curl -X PATCH -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" -H "Content-Type: application/json-patch+json" \
--cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/namespaces/%s/pods/$HOSTNAME \
-d '[{"op": "replace", "path": "/metadata/labels/role", "value": "follower"}]'
`, c.Namespace)
}
