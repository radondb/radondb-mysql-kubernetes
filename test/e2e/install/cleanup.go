package install

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
)

var _ = Describe("clean", Ordered, func() {
	f := framework.NewFramework("e2e-test")

	It("clean up crds", func() {
		f.CleanUpCRDs()
	})

	It("clean up the release", func() {
		framework.CleanUpOperatorAtNS(framework.DefaultE2ETestNS)
	})

	It("clean up ns", func() {
		f.CleanUpNS()
	})
})
