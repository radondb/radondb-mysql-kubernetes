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

package framework

import (
	"flag"
	"os"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var Log = logf.Log.WithName("framework.util")

// Default values of the test config.
const (
	// Export POD logs and test overview.
	DumpLogs = false
	// Optional directory to store junit and pod logs output in.
	// If not specified, it will beset to the current date.
	LogDirPrefix = ""

	// // Image path of mysql operator.
	// OperatorImagePath = "radondb/mysql-operator:v2.3.0"
	// Image path for mysql sidecar.
	SidecarImagePath = "radondb/mysql57-sidecar:v3.0.0"

	// The namespace where the resource created by E2E.
	E2ETestNamespace = "radondb-mysql-e2e"
	// The name of the Operator to create.
	OperatorReleaseName = "e2e-test"
	// The name of the MySQL cluster to create.
	ClusterReleaseName = "sample"

	// How often to Poll pods, nodes and claims.
	Poll = 2 * time.Second
	// Time interval of checking cluster status.
	POLLING = 2 * time.Second
	// Timeout time of checking cluster status.
	TIMEOUT = time.Minute
	// Timeout of deleting namespace.
	DefaultNamespaceDeletionTimeout = 10 * time.Minute
	// Timeout of waiting for POD startup.
	PodStartTimeout = 1 * time.Hour
)

type TestContextType struct {
	KubeHost         string
	KubeConfig       string
	KubeContext      string
	ExpectedVersion  string
	E2ETestNamespace string
	// OperatorImagePath string
	ClusterReleaseName string
	SidecarImagePath   string
	MysqlVersion       string
	TimeoutSeconds     int
	LogDirPrefix       string
	DumpLogs           bool
}

var TestContext TestContextType

// Register flags common to all e2e test suites.
func RegisterCommonFlags() {
	flag.StringVar(&TestContext.KubeHost, "kubernetes-host", "", "The kubernetes host, or apiserver, to connect to")
	flag.StringVar(&TestContext.KubeConfig, "kubernetes-config", os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "Path to config containing embedded authinfo for kubernetes. Default value is from environment variable "+clientcmd.RecommendedConfigPathEnvVar)
	flag.StringVar(&TestContext.KubeContext, "kubernetes-context", "", "config context to use for kuberentes. If unset, will use value from 'current-context'")
	flag.StringVar(&TestContext.ExpectedVersion, "expected-version", "", "Expected Chart version, For ci.")
	flag.StringVar(&TestContext.E2ETestNamespace, "namespace", E2ETestNamespace, "Test namespace.")
	// flag.StringVar(&TestContext.OperatorImagePath, "operator-image-path", OperatorImagePath, "Image path of mysql operator.")
	flag.StringVar(&TestContext.ClusterReleaseName, "cluster-release", ClusterReleaseName, "Release name of the mysql cluster.")
	flag.StringVar(&TestContext.SidecarImagePath, "sidecar-image-path", SidecarImagePath, "Image path of mysql sidecar.")
	flag.StringVar(&TestContext.MysqlVersion, "mysql-version", "5.7", "The version of mysql to be installed.")
	flag.StringVar(&TestContext.LogDirPrefix, "log-dir-prefix", LogDirPrefix, "Prefix of the log directory.")
	flag.IntVar(&TestContext.TimeoutSeconds, "pod-wait-timeout", 1200, "Timeout to wait for a pod to be ready.")
	flag.BoolVar(&TestContext.DumpLogs, "dump-logs", false, "Dump logs when test case failed.")
}

func RegisterParseFlags() {
	RegisterCommonFlags()
	flag.Parse()
}
