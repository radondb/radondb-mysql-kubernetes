package chaos

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/radondb/radondb-mysql-kubernetes/test/e2e/framework"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var _ = Describe("podchaos", Ordered, func() {
	f := framework.NewFramework("e2e-test")
	clusterKey := &types.NamespacedName{Name: framework.TestContext.ClusterReleaseName, Namespace: f.Namespace.Name}

	BeforeAll(func() {
		f.BeforeEach()
		// Make sure operator is available
		By("checking webhook")
		f.WaitUntilServiceAvailable(framework.WebhookServiceName)
		By("checking manager")
		Expect(f.CreateOperatorHealthCheckService()).Should(Succeed())
		Expect(f.CheckServiceEndpoint(framework.HealthCheckServiceName, 8081, "healthz")).Should(Succeed())

		By("checking mysql cluster")
		f.WaitClusterReadiness(clusterKey)
	})

	AfterEach(func() {
		f.CleanUpChaos()
	})

	It("kill leader", func() {
		killLeader(f)
		time.Sleep(10 * time.Second)
		f.WaitClusterReadiness(clusterKey)
	})

	It("kill leader mysql", func() {
		killLeaderMySQL(f)
		time.Sleep(10 * time.Second)
		f.WaitClusterReadiness(clusterKey)
	})

	It("kill leader xenon", func() {
		killLeaderXenon(f)
		time.Sleep(10 * time.Second)
		f.WaitClusterReadiness(clusterKey)
	})

	It("leader failure", func() {
		LeaderFailure(f, "30s")
		time.Sleep(10 * time.Second)
		f.WaitClusterReadiness(clusterKey)
	})
})

func killLeader(f *framework.Framework) {
	By("killing leader")
	f.KillPod(&framework.PodChaosOptions{
		Labels: map[string]string{"role": string(utils.Leader)},
	})
}

func killLeaderMySQL(f *framework.Framework) {
	By("killing leader mysql")
	f.KillContainers(&framework.PodChaosOptions{
		Labels:     map[string]string{"role": string(utils.Leader)},
		Containers: []string{"mysql"},
	})
}

func killLeaderXenon(f *framework.Framework) {
	By("killing leader xenon")
	f.KillContainers(&framework.PodChaosOptions{
		Labels:     map[string]string{"role": string(utils.Leader)},
		Containers: []string{"xenon"},
	})
}

func LeaderFailure(f *framework.Framework, duration string) {
	By(fmt.Sprintf("leader failure: %v", duration))
	f.PodFailure(&framework.PodChaosOptions{
		Labels:   map[string]string{"role": string(utils.Leader)},
		Duration: duration,
	})
}
