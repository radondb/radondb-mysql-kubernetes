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
	"context"
	"errors"
	"fmt"
	"github.com/blang/semver"
	"github.com/go-ini/ini"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	osrun "runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sort"
	"strings"

	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

type mysqlCMSyncer struct {
	*mysqlcluster.MysqlCluster

	cli client.Client

	cm *corev1.ConfigMap

	log logr.Logger
}

// NewMysqlCMSyncer returns mysql configmap syncer.
func NewMysqlCMSyncer(cli client.Client, c *mysqlcluster.MysqlCluster) *mysqlCMSyncer {
	return &mysqlCMSyncer{
		MysqlCluster: c,
		cli:          cli,
		cm: &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.GetNameForResource(utils.ConfigMap),
				Namespace: c.Namespace,
				Labels:    c.GetLabels(),
			},
		},
		log: logf.Log.WithName("MySQLCMSyncer"),
	}
}

// Object returns the object for which sync applies.
func (s *mysqlCMSyncer) Object() interface{} {
	return s.cm
}

// Owner returns the object owner or nil if object does not have one.
func (s *mysqlCMSyncer) ObjectOwner() runtime.Object {
	return s.MysqlCluster
}

// Sync persists data into the external store.
func (s *mysqlCMSyncer) Sync(ctx context.Context) (SyncResult, error) {
	var err error
	var kind string
	result := SyncResult{}

	result.Operation, err = s.createOrUpdate(ctx)

	// Get namespace and name.
	key := client.ObjectKeyFromObject(s.cm)
	// Get groupVersionKind.
	gvk, gvkErr := apiutil.GVKForObject(s.cm, s.cli.Scheme())
	if gvkErr != nil {
		kind = fmt.Sprintf("%T", s.cm)
	} else {
		kind = gvk.String()
	}

	switch {
	case errors.Is(err, ErrOwnerDeleted):
		s.log.Info(string(result.Operation), "key", key, "kind", kind, "error", err)
		err = nil
	case errors.Is(err, ErrIgnore):
		s.log.Info("syncer skipped", "key", key, "kind", kind, "error", err)
		err = nil
	case err != nil:
		result.SetEventData("Warning", basicEventReason(s.Name, err), fmt.Sprintf("%s %s failed syncing: %s", kind, key, err))
		s.log.Error(err, string(result.Operation), "key", key, "kind", kind)
	default:
		result.SetEventData("Normal", basicEventReason(s.Name, err), fmt.Sprintf("%s %s %s successfully", kind, key, result.Operation))
		s.log.Info(string(result.Operation), "key", key, "kind", kind)
	}
	return result, err
}

func (s *mysqlCMSyncer) createOrUpdate(ctx context.Context) (controllerutil.OperationResult, error) {
	var err error
	if err = s.cli.Get(ctx, client.ObjectKeyFromObject(s.cm), s.cm); err != nil {
		if !k8serrors.IsNotFound(err) {
			return resultNone, err
		}

		if s.Spec.MysqlOpts.MysqlConfTemplate != "" {
			return resultNone, fmt.Errorf("template is not exist: %s", s.Spec.MysqlOpts.MysqlConfTemplate)
		}

		if err = s.generateTemplate(ctx); err != nil {
			return resultNone, err
		}

		if err = s.cli.Create(ctx, s.cm); err != nil {
			return resultNone, err
		} else {
			return resultCreated, nil
		}
	}

	if err := s.appendConf(); err != nil {
		return resultNone, err
	}

	if err := s.setControllerReference(); err != nil {
		return resultNone, err
	}

	if err = s.cli.Update(ctx, s.cm); err != nil {
		return resultNone, err
	}

	return resultNone, nil
}

func (s *mysqlCMSyncer) generateTemplate(ctx context.Context) error {
	if s.Spec.MysqlOpts.MysqlConfTemplate == "" {
		data, err := buildMysqlConf(s.MysqlCluster)
		if err != nil {
			return fmt.Errorf("failed to create mysql configs: %s", err)
		}

		dataPlugin, err := buildMysqlPluginConf(s.MysqlCluster)
		if err != nil {
			return fmt.Errorf("failed to create mysql plugin configs: %s", err)
		}
		s.cm.Data = map[string]string{
			"my.cnf":            data,
			utils.PluginConfigs: dataPlugin,
		}
		return nil
	}
	return fmt.Errorf("MysqlConfTemplate is empty")
}

// Notice: The parameters will not be removed from the cm when its removed from the mysqlConf/pluginConf.
func (s *mysqlCMSyncer) appendConf() error {
	if err := s.createOrReplaceIniKey("my.cnf", s.Spec.MysqlOpts.MysqlConf); err != nil {
		return err
	}
	if s.Spec.MysqlVersion == "8.0" {
		var str string = pluginConfigs["plugin-load"]
		if osrun.GOARCH == "aarch64" || osrun.GOARCH == "arm64" {
			str = "\"semisync_master.so;semisync_slave.so\""
		}
		str = str[0:len(str)-1] + ";mysql_clone.so\""
		if s.Spec.MysqlOpts.PluginConf != nil {
			s.Spec.MysqlOpts.PluginConf["plugin-load"] = str
		} else {
			s.Spec.MysqlOpts.PluginConf = map[string]string{"plugin-load": str}
		}
	}
	if err := s.createOrReplaceIniKey("plugin.cnf", s.Spec.MysqlOpts.PluginConf); err != nil {
		return err
	}
	return nil
}

func (s *mysqlCMSyncer) setControllerReference() error {
	if s.MysqlCluster == nil {
		return fmt.Errorf("owner is nil")
	}
	// Set owner reference only if owner resource is not being deleted, otherwise the owner
	// reference will be reset in case of deleting with cascade=false.
	if s.Unwrap().GetDeletionTimestamp().IsZero() {
		if err := controllerutil.SetControllerReference(s.Unwrap(), s.cm, s.cli.Scheme()); err != nil {
			return err
		}
	} else if ctime := s.Unwrap().GetCreationTimestamp(); ctime.IsZero() {
		// The owner is deleted, don't recreate the resource if does not exist, because gc
		// will not delete it again because has no owner reference set.
		return fmt.Errorf("owner is deleted")
	}
	return nil
}

func (s *mysqlCMSyncer) createOrReplaceIniKey(key string, patch map[string]string) error {
	if len(patch) == 0 {
		return nil
	}
	if f, ok := s.cm.Data[key]; ok {
		if iniFile, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true, AllowBooleanKeys: true}, []byte(f)); err != nil {
			return fmt.Errorf("failed to load %s, err: %s", key, err.Error())
		} else {
			sec := iniFile.Section("mysqld")
			keys := []string{}
			for k := range patch {
				keys = append(keys, k)
			}
			sort.Sort(StringsConnectedByBar(keys))

			for _, k := range keys {
				// TODO: repeatable key, like plugin-load-add
				// replace
				if sec.HasKey(barskey(k)) {
					sec.Key(barskey(k)).SetValue(patch[k])
				} else if sec.HasKey(underscorekey(k)) {
					sec.Key(underscorekey(k)).SetValue(patch[k])
				} else { // Not in sec.
					// Add it to sec
					if _, err := sec.NewKey(k, patch[k]); err != nil {
						return fmt.Errorf("failed to add key to config section: %s", err)
					}
				}
			}
			data, err := writeConfigs(iniFile)
			if err != nil {
				return fmt.Errorf("failed to write configs: %s", err)
			}
			s.cm.Data[key] = data
		}
	}
	return nil
}

// buildMysqlConf build the mysql config.
func buildMysqlConf(c *mysqlcluster.MysqlCluster) (string, error) {
	var log = logf.Log.WithName("mysqlcluster.syncer.buildMysqlConf")
	cfg := ini.Empty(ini.LoadOptions{IgnoreInlineComment: true})
	sec := cfg.Section("mysqld")
	arr := strings.Split(c.Spec.MysqlOpts.Image, ":")
	VerCurrent := arr[len(arr)-1]
	mysqlSemVerCurr, _ := semver.Parse(VerCurrent)
	mysqlSemVer31, _ := semver.Parse("8.0.31")
	c.EnsureMysqlConf()
	mysqlVersion := ""
	if strings.Contains(c.Spec.MysqlOpts.Image, "8.0") {
		mysqlVersion = "8.0"
	} else if strings.Contains(c.Spec.MysqlOpts.Image, "5.7") {
		mysqlVersion = "5.7"
	}
	switch mysqlVersion {
	case "8.0":
		addKVConfigsToSection(sec, mysql80Configs)
	case "5.7":
		addKVConfigsToSection(sec, mysql57Configs)
	}
	if mysqlSemVerCurr.GE(mysqlSemVer31) {
		// at first , copy all the map, just for version over 8.0.31
		//var mysqlStaticConfigsCpy, mysqlSysConfigsCpy, mysqlCommonConfigsCpy map[string]string
		mysqlStaticConfigsCpy := make(map[string]string)
		mysqlSysConfigsCpy := make(map[string]string)
		mysqlCommonConfigsCpy := make(map[string]string)
		cpyMap(mysqlStaticConfigs, mysqlStaticConfigsCpy)
		cpyMap(mysqlSysConfigs, mysqlSysConfigsCpy)
		cpyMap(mysqlCommonConfigs, mysqlCommonConfigsCpy)
		replaceMapKey(mysqlStaticConfigsCpy, "slave_pending_jobs_size_max", "replica_pending_jobs_size_max")
		replaceMapKey(mysqlStaticConfigsCpy, "slave_preserve_commit_order", "replica_preserve_commit_order")
		// slave_parallel_workers -> replica_parallel_workers
		replaceMapKey(mysqlStaticConfigsCpy, "slave_parallel_workers", "replica_parallel_workers")
		mysqlStaticConfigsCpy["replica_parallel_workers"] = "1"
		// mysqlSysConfigs
		// 1. slave_parallel_type -> replica_parallel_type
		replaceMapKey(mysqlSysConfigsCpy, "slave_parallel_type", "replica_parallel_type")
		// mysqlCommonConfigs
		// sync_master_info -> sync_source_info
		replaceMapKey(mysqlCommonConfigsCpy, "sync_master_info", "sync_source_info")
		addKVConfigsToSection(sec, mysqlSysConfigsCpy, mysqlCommonConfigsCpy, mysqlStaticConfigsCpy)
	} else {
		addKVConfigsToSection(sec, mysqlSysConfigs, mysqlCommonConfigs, mysqlStaticConfigs)
	}

	if c.Spec.MysqlOpts.InitTokuDB {
		addKVConfigsToSection(sec, mysqlTokudbConfigs)
	}

	for _, key := range mysqlBooleanConfigs {
		if mysqlSemVerCurr.GE(mysqlSemVer31) {
			switch key {
			case "skip-host-cache":
				sec.NewKey("host_cache_size", "0")
				continue
			case "skip-slave-start":
				key = "skip_replica_start"
			case "log-slave-updates":
				key = "log_replica_updates"
			}
		}
		if _, err := sec.NewBooleanKey(key); err != nil {
			log.Error(err, "failed to add boolean key to config section", "key", key)
		}
	}
	if len(c.Spec.TlsSecretName) != 0 {
		addKVConfigsToSection(sec, mysqlSSLConfigs)
	}
	addKVConfigsToSection(sec, c.Spec.MysqlOpts.MysqlConf)

	data, err := writeConfigs(cfg)
	if err != nil {
		return "", err
	}

	return data, nil
}

// Build the Plugin Cnf file.
func buildMysqlPluginConf(c *mysqlcluster.MysqlCluster) (string, error) {
	cfg := ini.Empty(ini.LoadOptions{IgnoreInlineComment: true})
	sec := cfg.Section("mysqld")
	pluginConfigsCpy := make(map[string]string)
	cpyMap(pluginConfigs, pluginConfigsCpy)

	if osrun.GOARCH == "aarch64" || osrun.GOARCH == "arm64" {
		delete(pluginConfigsCpy, "audit_log_file")
		delete(pluginConfigsCpy, "audit_log_exclude_accounts")
		delete(pluginConfigsCpy, "audit_log_buffer_size")
		delete(pluginConfigsCpy, "audit_log_policy")
		delete(pluginConfigsCpy, "audit_log_rotate_on_size")
		delete(pluginConfigsCpy, "audit_log_rotations")
		delete(pluginConfigsCpy, "audit_log_format")
	}
	addKVConfigsToSection(sec, pluginConfigsCpy)
	addKVConfigsToSection(sec, c.Spec.MysqlOpts.PluginConf)
	data, err := writeConfigs(cfg)
	if err != nil {
		return "", err
	}

	return data, nil
}

// addKVConfigsToSection add a map[string]string to a ini.Section
func addKVConfigsToSection(s *ini.Section, extraMysqld ...map[string]string) {
	var log = logf.Log.WithName("mysqlcluster.syncer.addKVConfigsToSection")
	for _, extra := range extraMysqld {
		keys := []string{}
		for key := range extra {
			keys = append(keys, key)
		}

		// sort keys
		sort.Sort(StringsConnectedByBar(keys))

		for _, k := range keys {
			//exist, just replace
			if s.HasKey(barskey(k)) {
				s.Key(barskey(k)).SetValue(extra[k])
			} else if s.HasKey(underscorekey(k)) {
				s.Key(underscorekey(k)).SetValue(extra[k])
			} else {
				//else add it
				value := extra[k]
				if _, err := s.NewKey(k, value); err != nil {
					log.Error(err, "failed to add key to config section", "key", k, "value", extra[k], "section", s)
				}
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

func barskey(k string) string {
	return strings.ReplaceAll(k, "_", "-")
}

func underscorekey(k string) string {
	return strings.ReplaceAll(k, "-", "_")
}

type StringsConnectedByBar []string

func (x StringsConnectedByBar) Len() int           { return len(x) }
func (x StringsConnectedByBar) Less(i, j int) bool { return barskey(x[i]) < barskey(x[j]) }
func (x StringsConnectedByBar) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func replaceMapKey(dict map[string]string, key1 string, key2 string) {
	dict[key2] = dict[key1]
	delete(dict, key1)
}

func cpyMap(src map[string]string, dest map[string]string) {
	for k, v := range src {
		dest[k] = v
	}
}
