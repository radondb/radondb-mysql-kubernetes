package user

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("login", Ordered, func() {
	f := framework.NewFramework("e2e-test")

	BeforeAll(func() {
		f.BeforeEach()
		// Make sure operator is available
		By("check webhook")
		f.WaitUntilServiceAvailable(framework.WebhookServiceName)
		By("check manager")
		Expect(f.CheckServiceEndpoint(framework.HealthCheckServiceName, 8081, "healthz")).Should(Succeed())

		By("check mysql cluster")
		f.WaitClusterReadiness(&types.NamespacedName{Name: "sample", Namespace: f.Namespace.Name})
	})

	It("root@localhost", func() {
		f.CheckLogIn("root", "")
	})

	It("radondb_usr@localhost", func() {
		f.CheckLogIn("radondb_usr", "RadonDB@123")
	})
})
