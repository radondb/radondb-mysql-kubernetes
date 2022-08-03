/*
Copyright 2022 RadonDB.

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

package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/radondb/radondb-mysql-kubernetes/internal"
)

var (
	ns      string
	podName string

	newMySQLChecker = internal.NewMysqlChecker
	getMyRole       = internal.GetMyRole
	getMyHealthy    = internal.GetMyHealthy
)

const (
	mysqlUser = "root"
	mysqlHost = "127.0.0.1"
	mysqlPort = 3306
	mysqlPwd  = ""
)

func init() {
	ns = os.Getenv("NAMESPACE")
	podName = os.Getenv("POD_NAME")
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s liveness|readiness|command|startup", os.Args[0])
	}
	switch os.Args[1] {
	case "liveness":
		if err := liveness(); err != nil {
			log.Fatalf("liveness failed: %s", err.Error())
		}
		log.Info("node is liveness")
	case "readiness":
		if err := readiness(); err != nil {
			log.Fatalf("readiness failed: %s", err.Error())
		}
		log.Info("node is readiness")
	case "startup":
		if err := startup(); err != nil {
			log.Fatalf("readiness failed: %s", err.Error())
		}
	case "command":
		if err := command(); err != nil {
			log.Fatalf("start failed: %s", err.Error())
		}
	default:
		log.Fatalf("Usage: %s liveness|readiness|command|startup", os.Args[0])
	}
}

func liveness() error {
	if internal.SleepForever() {
		log.Infof("sleep-forever is set, skip readiness check")
		return nil
	}
	sqlrunner, closeFn, err := internal.NewSQLRunner(localMySQLConfig())
	if err != nil {
		return err
	}
	defer closeFn()

	mc := newMySQLChecker(sqlrunner, localClientOptions())
	return mc.CheckMySQLLiveness()
}

func readiness() error {
	if internal.SleepForever() {
		log.Infof("sleep-forever is set, skip readiness check")
		return nil
	}

	sqlrunner, closeFn, err := internal.NewSQLRunner(localMySQLConfig())
	if err != nil {
		log.Errorf("failed to create sqlrunner: %s", err.Error())
	}
	defer closeFn()

	mc := newMySQLChecker(sqlrunner, localClientOptions())
	labels := mc.ReadLabels()
	role, oldHealthy := getMyRole(labels), getMyHealthy(labels) == "yes"
	log.Infof("role: %s, healthy: %t", role, oldHealthy)

	if newHealthy := mc.CheckMySQLHealthy(role); newHealthy != oldHealthy {
		healthyLebel := internal.ConvertHealthy(newHealthy)
		log.Infof("patch healthy label: %s", healthyLebel)

		if err := mc.PatchLabel("healthy", healthyLebel); err != nil {
			log.Errorf("failed to patch healthy label: %s", err.Error())
		}
	}
	return nil
}

func command() error {
	return nil
}

func startup() error {
	return nil
}

func localMySQLConfig() *internal.Config {
	return &internal.Config{
		User:     mysqlUser,
		Host:     mysqlHost,
		Port:     mysqlPort,
		Password: mysqlPwd,
	}
}

func localClientOptions() *internal.ClientOptions {
	return &internal.ClientOptions{NameSpace: ns, PodName: podName}
}
