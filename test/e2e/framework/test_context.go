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

// RegisterCommonFlags registers flags common to all e2e test suites.
// The flag set can be flag.CommandLine (if desired) or a custom
// flag set that then gets passed to viperconfig.ViperizeFlags.
//
// The other Register*Flags methods below can be used to add more
// test-specific flags. However, those settings then get added
// regardless whether the test is actually in the test suite.
//
// For tests that have been converted to registering their
// options themselves, copy flags from test/e2e/framework/config
// as shown in HandleFlags.
func RegisterCommonFlags(flags *flag.FlagSet) {
	flags.StringVar(&TestContext.KubeConfig, clientcmd.RecommendedConfigPathFlag, os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "Path to kubeconfig containing embedded authinfo.")
	flags.StringVar(&TestContext.KubeContext, clientcmd.FlagContext, "", "kubeconfig context to use/override. If unset, will use value from 'current-context'")
	flags.StringVar(&TestContext.Host, "host", "", fmt.Sprintf("The host, or apiserver, to connect to. Will default to %s if this argument and --kubeconfig are not set.", defaultHost))
	flags.StringVar(&TestContext.OutputDir, "e2e-output-dir", "/tmp", "Output directory for interesting/useful test data, like performance data, benchmarks, and other metrics.")
	flags.BoolVar(&TestContext.DumpLogsOnFailure, "dump-logs-on-failure", true, "If set to true test will dump data about the namespace in which test was running.")
	flags.BoolVar(&TestContext.DisableLogDump, "disable-log-dump", false, "If set to true, logs from master and nodes won't be gathered after test run.")
	flag.IntVar(&TestContext.TimeoutSeconds, "pod-wait-timeout", 100, "Timeout to wait for a pod to be ready.")
}
