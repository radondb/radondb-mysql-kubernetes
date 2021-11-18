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
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

const (
	// TODO(gry): make sure the default port, do we need it?
	defaultHost = "https://127.0.0.1:6443"

	// DefaultNumNodes is the number of nodes. If not specified, then number of nodes is auto-detected
	DefaultNumNodes = -1
)

// TestContextType contains test settings and global state. Due to
// historic reasons, it is a mixture of items managed by the test
// framework itself, cloud providers and individual tests.
// The goal is to move anything not required by the framework
// into the code which uses the settings.
//
// The recommendation for those settings is:
// - They are stored in their own context structure or local
//   variables.
// - The standard `flag` package is used to register them.
//   The flag name should follow the pattern <part1>.<part2>....<partn>
//   where the prefix is unlikely to conflict with other tests or
//   standard packages and each part is in lower camel case. For
//   example, test/e2e/storage/csi/context.go could define
//   storage.csi.numIterations.
// - framework/config can be used to simplify the registration of
//   multiple options with a single function call:
//   var storageCSI {
//       NumIterations `default:"1" usage:"number of iterations"`
//   }
//   _ config.AddOptions(&storageCSI, "storage.csi")
// - The direct use Viper in tests is possible, but discouraged because
//   it only works in test suites which use Viper (which is not
//   required) and the supported options cannot be
//   discovered by a test suite user.
//
// Test suite authors can use framework/viper to make all command line
// parameters also configurable via a configuration file.
type TestContextType struct {
	KubeConfig  string
	KubeContext string
	Host        string

	OutputDir string

	// If set to true test will dump data about the namespace in which test was running.
	DumpLogsOnFailure bool
	// Disables dumping cluster log from master and nodes after all tests.
	DisableLogDump bool
	TimeoutSeconds int
}

// TestContext should be used by all tests to access common context data.
var TestContext TestContextType

// Register flags common to all e2e test suites.
func RegisterCommonFlags() {
	// Turn on verbose by default to get spec names
	config.DefaultReporterConfig.Verbose = true

	// Turn on EmitSpecProgress to get spec progress (especially on interrupt)
	config.GinkgoConfig.EmitSpecProgress = true

	// Randomize specs as well as suites
	config.GinkgoConfig.RandomizeAllSpecs = true

	flag.StringVar(&TestContext.KubeHost, "kubernetes-host", "", "The kubernetes host, or apiserver, to connect to")
	flag.StringVar(&TestContext.KubeConfig, "kubernetes-config", os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "Path to config containing embedded authinfo for kubernetes. Default value is from environment variable "+clientcmd.RecommendedConfigPathEnvVar)
	flag.StringVar(&TestContext.KubeContext, "kubernetes-context", "", "config context to use for kuberentes. If unset, will use value from 'current-context'")

	flag.StringVar(&TestContext.ReportDir, "report-dir", "", "Optional directory to store junit and pod logs output in. If not specified, no junit or logs files will be output")
	flag.StringVar(&TestContext.ChartPath, "operator-chart-path", "../../charts/mysql-operator", "The chart name or path for mysql operator")

	flag.StringVar(&TestContext.OperatorImageTag, "operator-image-tag", "latest", "Image tag for mysql operator.")

	flag.IntVar(&TestContext.TimeoutSeconds, "pod-wait-timeout", 1200, "Timeout to wait for a pod to be ready.")
	flag.BoolVar(&TestContext.DumpLogsOnFailure, "dump-logs-on-failure", true, "Dump pods logs when a test fails.")
}

func RegisterParseFlags() {
	RegisterCommonFlags()
	flag.Parse()
}
