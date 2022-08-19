package install

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
)

var _ = Describe("install operator", Ordered, func() {
	f := framework.NewFramework("e2e-test")
	opt := framework.NewOperatorOptions()

	BeforeAll(func() {
		By("add radondb helm repo")
		f.AddRadonDBHelmRepo()

		By("check lastest version")
		Expect(f.CheckVersion()).Should(Not(Equal("No results found")))
	})

	AfterAll(func() {
		f.RemoveRadonDBHelmRepo()
	})

	It("helm install", func() {
		f.InstallOperator(opt)
	})

	It("check the webhook service", func() {
		f.WaitUntilServiceAvailable(framework.WebhookServiceName)
	})

	It("check operator health", func() {
		By("create health check service")
		Expect(f.CreateOperatorHealthCheckService()).Should(Succeed())
		f.WaitUntilServiceAvailable(framework.HealthCheckServiceName)

		By("check endpoint")
		Expect(f.CheckServiceEndpoint(framework.HealthCheckServiceName, 8081, "healthz")).Should(Succeed())
	})
})
