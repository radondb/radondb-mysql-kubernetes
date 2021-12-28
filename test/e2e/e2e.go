/*
Copyright 2015 The Kubernetes Authors.

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

package e2e

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/golang/glog"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework/ginkgowrapper"
)

const (
	operatorNamespace = "mysql-operator"
)

var _ = SynchronizedBeforeSuite(func() []byte {
	// BeforeSuite logic.
	return nil
}, func(data []byte) {
	// all other nodes
	framework.Logf("Running BeforeSuite actions on all node")
})

// Similar to SynchornizedBeforeSuite, we want to run some operations only once (such as collecting cluster logs).
// Here, the order of functions is reversed; first, the function which runs everywhere,
// and then the function that only runs on the first Ginkgo node.
var _ = SynchronizedAfterSuite(func() {
	// AfterSuite logic.
}, func() {
	// Run only Ginkgo on node 1
	framework.Logf("Running AfterSuite actions on node 1")
})

// RunE2ETests checks configuration parameters (specified through flags) and then runs
// E2E tests using the Ginkgo runner.
// If a "report directory" is specified, one or more JUnit test reports will be
// generated in this directory, and cluster logs will also be saved.
// This function is called on each Ginkgo node in parallel mode.
func RunE2ETests(t *testing.T) {
	runtimeutils.ReallyCrash = true

	RegisterFailHandler(ginkgowrapper.Fail)
	// Disable skipped tests unless they are explicitly requested.
	if len(config.GinkgoConfig.FocusStrings) == 0 && len(config.GinkgoConfig.SkipStrings) == 0 {
		config.GinkgoConfig.SkipStrings = []string{`\[Flaky\]`, `\[Feature:.+\]`}
	}

	rps := func() (rps []Reporter) {
		// Run tests through the Ginkgo runner with output to console + JUnit for Jenkins
		if framework.TestContext.ReportDir != "" {
			// TODO: we should probably only be trying to create this directory once
			// rather than once-per-Ginkgo-node.
			if err := os.MkdirAll(framework.TestContext.ReportDir, 0755); err != nil {
				glog.Errorf("Failed creating report directory: %v", err)
				return
			}
			// add junit report
			rps = append(rps, reporters.NewJUnitReporter(path.Join(framework.TestContext.ReportDir, fmt.Sprintf("junit_%v%02d.xml", "mysql_o_", config.GinkgoConfig.ParallelNode))))

			// add logs dumper
			if framework.TestContext.DumpLogsOnFailure {
				rps = append(rps, NewLogsPodReporter(operatorNamespace, path.Join(framework.TestContext.ReportDir,
					fmt.Sprintf("pods_logs_%d_%d.txt", config.GinkgoConfig.RandomSeed, config.GinkgoConfig.ParallelNode))))
			}
		} else {
			// if reportDir is not specified then print logs to stdout
			if framework.TestContext.DumpLogsOnFailure {
				rps = append(rps, NewLogsPodReporter(operatorNamespace, ""))
			}
		}
		return
	}()

	glog.Infof("Starting e2e run on Ginkgo node %d", config.GinkgoConfig.ParallelNode)

	RunSpecsWithDefaultAndCustomReporters(t, "MySQL Operator E2E Suite", rps)
}
