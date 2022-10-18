package install

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("install cluster", Label("installCluster"), Ordered, func() {
	f := framework.NewFramework("e2e-test")

	BeforeAll(func() {
		f.BeforeEach()
		// Make sure operator is available
		By("check webhook")
		f.WaitUntilServiceAvailable(framework.WebhookServiceName)
		By("check manager")
		Expect(f.CreateOperatorHealthCheckService()).Should(Succeed())
		Expect(f.CheckServiceEndpoint(framework.HealthCheckServiceName, 8081, "healthz")).Should(Succeed())
	})

	AfterAll(func() {})

	It("using asset", Label("asset"), func() {
		Expect(f.InstallMySQLClusterUsingAsset()).Should(Succeed())
		By("wait cluster ready")
		f.WaitClusterReadiness(&types.NamespacedName{Name: framework.TestContext.ClusterReleaseName, Namespace: f.Namespace.Name})
	})
})
