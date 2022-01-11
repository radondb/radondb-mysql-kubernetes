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
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework/ginkgowrapper"

	// Test case source.
	// Comment out the package that you don't want to run.
	_ "github.com/radondb/radondb-mysql-kubernetes/test/e2e/cluster"
	_ "github.com/radondb/radondb-mysql-kubernetes/test/e2e/simplecase"
)

func init() {
	testing.Init()
	framework.RegisterParseFlags()
}

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}

var _ = SynchronizedBeforeSuite(func() []byte {
	kubeCfg, err := framework.LoadConfig()
	Expect(err).To(Succeed())

	c, err := client.New(kubeCfg, client.Options{})
	if err != nil {
		Fail(fmt.Sprintf("can't instantiate k8s client: %s", err))
	}

	By("Create Namespace")
	operatorNsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: framework.RadondbMysqlE2eNamespace,
		},
	}

	if err := c.Create(context.TODO(), operatorNsObj); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			Fail(fmt.Sprintf("can't create mysql-operator namespace: %s", err))
		}
	}

	if framework.TestContext.DumpLogs {
		By("Create log dir")
		os.MkdirAll(fmt.Sprintf("%s_%d", framework.TestContext.LogDirPrefix, GinkgoRandomSeed()), 0777)
	}

	By("Install operator")
	framework.HelmInstallChart(framework.OperatorReleaseName, framework.RadondbMysqlE2eNamespace)
	return nil
}, func(_ []byte) {
	framework.Logf("Running BeforeSuite actions on all node")
})

// Similar to SynchornizedBeforeSuite, we want to run some operations only once (such as collecting cluster logs).
// Here, the order of functions is reversed; first, the function which runs everywhere,
// and then the function that only runs on the first Ginkgo node.
var _ = SynchronizedAfterSuite(func() {
	// Run on all Ginkgo nodes.
	framework.Logf("Running AfterSuite actions on all node")
	framework.RunCleanupActions()

	// Get the kubernetes client.
	kubeCfg, err := framework.LoadConfig()
	Expect(err).To(Succeed())

	client, err := clientset.NewForConfig(kubeCfg)
	Expect(err).NotTo(HaveOccurred())

	By("Remove operator")
	framework.HelmPurgeRelease(framework.OperatorReleaseName, framework.RadondbMysqlE2eNamespace)

	By("Delete test namespace")
	if err := framework.DeleteNS(client, framework.RadondbMysqlE2eNamespace, framework.DefaultNamespaceDeletionTimeout); err != nil {
		framework.Failf(fmt.Sprintf("Can't delete namespace: %s", err))
	}
}, func() {
	framework.Logf("Running AfterSuite actions on node 1")
})

var _ = ReportAfterSuite("Collect log", func(report Report) {
	if framework.TestContext.DumpLogs {
		f, err := os.OpenFile(path.Join(fmt.Sprintf("%s_%d", framework.TestContext.LogDirPrefix, GinkgoRandomSeed()), "overview.txt"), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, specReport := range report.SpecReports {

			if specReport.FullText() != "" {
				// Collect the summary of all cases.
				fmt.Fprintf(f, "%s | %s | %v\n", specReport.FullText(), specReport.State, specReport.RunTime)
			}
			// TODO: Collect the POD log of failure cases.
		}
		f.Close()
	}
})

// RunE2ETests checks configuration parameters (specified through flags) and then runs
// E2E tests using the Ginkgo runner.
// If a "report directory" is specified, one or more JUnit test reports will be
// generated in this directory, and cluster logs will also be saved.
// This function is called on each Ginkgo node in parallel mode.
func RunE2ETests(t *testing.T) {
	RegisterFailHandler(ginkgowrapper.Fail)

	// Fetch the current config.
	suiteConfig, reporterConfig := GinkgoConfiguration()
	// Whether printing FullTrace.
	reporterConfig.FullTrace = true
	// Whether printing more detail.
	reporterConfig.Verbose = true
	if framework.TestContext.DumpLogs {
		if framework.TestContext.LogDirPrefix == "" {
			now := time.Now()
			framework.TestContext.LogDirPrefix = fmt.Sprintf("logs_%d%d_%d%d", now.Month(), now.Day(), now.Hour(), now.Minute())
		}
		// Path of JUnitReport.
		reporterConfig.JUnitReport = path.Join(fmt.Sprintf("%s_%d", framework.TestContext.LogDirPrefix, GinkgoRandomSeed()), "junit.xml")
	}

	RunSpecs(t, "MySQL Operator E2E Suite", Label("MySQL Operator"), suiteConfig, reporterConfig)
}
