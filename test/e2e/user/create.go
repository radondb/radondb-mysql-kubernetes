package user

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("create user", Ordered, func() {
	f := framework.NewFramework("e2e-test")

	BeforeAll(func() {
		f.BeforeEach()
		// Make sure operator is available
		By("check webhook")
		f.WaitUntilServiceAvailable(framework.WebhookServiceName)
		By("check manager")
		Expect(f.CreateOperatorHealthCheckService()).Should(Succeed())
		Expect(f.CheckServiceEndpoint(framework.HealthCheckServiceName, 8081, "healthz")).Should(Succeed())

		By("check mysql cluster")
		f.WaitClusterReadiness(&types.NamespacedName{Name: framework.TestContext.ClusterReleaseName, Namespace: f.Namespace.Name})
	})

	AfterAll(func() {
		By("clean up users")
		f.CleanUpUser()
	})

	It("create super_user@%", func() {
		f.CreateUserSecret()
		f.CreateSuperUser()
		By("check grants")
		f.CheckGantsForUser("super_user@'%'", true)
	})
})
