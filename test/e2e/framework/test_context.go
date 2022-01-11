/*
Copyright 2016 The Kubernetes Authors.

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

const (
	// The namespace where the resource created by E2E.
	RadondbMysqlE2eNamespace = "radondb-mysql-e2e"
	// The name of the Operator to create.
	OperatorReleaseName = "e2e-test"
	// Export POD logs and test overview.
	DumpLogs = true
	// Optional directory to store junit and pod logs output in.
	// If not specified, it will beset to the current date.
	ReportDirPrefix = ""

	// Specify the directory that Helm Install will be executed.
	ChartPath = "../../charts/mysql-operator"
	// Image path of mysql operator.
	OperatorImagePath = "radondb/mysql-operator"
	// Image tag of mysql operator.
	OperatorImageTag = "latest"
	// Image path for mysql sidecar.
	SidecarImage = "radondb/mysql-sidecar:latest"

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
	KubeHost    string
	KubeConfig  string
	KubeContext string

	ReportDirPrefix string

	ChartPath   string
	ChartValues string

	OperatorImagePath string
	OperatorImageTag  string
	SidecarImage      string

	TimeoutSeconds int
	DumpLogs       bool
}

var TestContext TestContextType

// Register flags common to all e2e test suites.
func RegisterCommonFlags() {
	flag.StringVar(&TestContext.KubeHost, "kubernetes-host", "", "The kubernetes host, or apiserver, to connect to")
	flag.StringVar(&TestContext.KubeConfig, "kubernetes-config", os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "Path to config containing embedded authinfo for kubernetes. Default value is from environment variable "+clientcmd.RecommendedConfigPathEnvVar)
	flag.StringVar(&TestContext.KubeContext, "kubernetes-context", "", "config context to use for kuberentes. If unset, will use value from 'current-context'")

	flag.StringVar(&TestContext.ReportDirPrefix, "report-dir", ReportDirPrefix, "Optional directory to store logs output in.")

	flag.StringVar(&TestContext.ChartPath, "operator-chart-path", ChartPath, "The chart name or path for mysql operator")
	flag.StringVar(&TestContext.OperatorImagePath, "operator-image-path", OperatorImagePath, "Image tag of mysql operator.")
	flag.StringVar(&TestContext.OperatorImageTag, "operator-image-tag", OperatorImageTag, "Image tag of mysql operator.")
	flag.StringVar(&TestContext.SidecarImage, "sidecar-image", SidecarImage, "Image path of mysql sidecar.")

	flag.IntVar(&TestContext.TimeoutSeconds, "pod-wait-timeout", 1200, "Timeout to wait for a pod to be ready.")
	flag.BoolVar(&TestContext.DumpLogs, "dump-logs-on-failure", DumpLogs, "Dump logs.")
}

func RegisterParseFlags() {
	RegisterCommonFlags()
	flag.Parse()
}
