package install

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
)

var _ = Describe("prepare", Ordered, func() {
	f := framework.NewFramework("e2e-test")

	It("create ns", func() {
		f.CreateNS()
	})
})
