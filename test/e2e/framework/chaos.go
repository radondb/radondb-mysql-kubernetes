package framework

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chaosapis "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type PodChaosOptions struct {
	Name       string
	Duration   string
	Mode       chaosapis.SelectorMode
	Containers []string
	Labels     map[string]string
}

func newPodChaos(name, ns string) *chaosapis.PodChaos {
	return &chaosapis.PodChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
}

func (f *Framework) KillPod(option *PodChaosOptions) {
	if option.Name == "" {
		option.Name = fmt.Sprintf("kill-pod-%d", rand.Intn(1000))
	}
	if option.Mode == "" {
		option.Mode = chaosapis.OneMode
	}
	Expect(option.Labels).ShouldNot(BeEmpty())

	podChaos := newPodChaos(option.Name, f.kubectlOptions.Namespace)
	podChaos.Spec.Action = chaosapis.PodKillAction
	podChaos.Spec.Mode = option.Mode
	podChaos.Spec.Selector.LabelSelectors = option.Labels

	Expect(f.Client.Create(context.TODO(), podChaos, &client.CreateOptions{})).Should(Succeed())
	f.waitChaosInjected(podChaos.Name, 30)
}

func (f *Framework) KillContainers(option *PodChaosOptions) {
	if option.Name == "" {
		option.Name = fmt.Sprintf("kill-containers-%d", rand.Intn(1000))
	}
	if option.Mode == "" {
		option.Mode = chaosapis.OneMode
	}
	Expect(option.Labels).ShouldNot(BeEmpty())

	podChaos := newPodChaos(option.Name, f.kubectlOptions.Namespace)
	podChaos.Spec.Action = chaosapis.ContainerKillAction
	podChaos.Spec.Mode = option.Mode
	podChaos.Spec.Selector.LabelSelectors = option.Labels
	podChaos.Spec.ContainerNames = option.Containers

	Expect(f.Client.Create(context.TODO(), podChaos, &client.CreateOptions{})).Should(Succeed())
	f.waitChaosInjected(podChaos.Name, 30)
}

func (f *Framework) PodFailure(option *PodChaosOptions) {
	if option.Name == "" {
		option.Name = fmt.Sprintf("pod-failure-%d", rand.Intn(1000))
	}
	if option.Mode == "" {
		option.Mode = chaosapis.OneMode
	}
	if option.Duration == "" {
		option.Duration = "9999s"
	}
	Expect(option.Labels).ShouldNot(BeEmpty())

	podChaos := newPodChaos(option.Name, f.kubectlOptions.Namespace)
	podChaos.Spec.Action = chaosapis.PodFailureAction
	podChaos.Spec.Mode = option.Mode
	podChaos.Spec.Selector.LabelSelectors = option.Labels
	podChaos.Spec.Duration = &option.Duration

	Expect(f.Client.Create(context.TODO(), podChaos, &client.CreateOptions{})).Should(Succeed())
	f.waitChaosInjected(podChaos.Name, 30)
}

func (f *Framework) waitChaosInjected(chaosName string, timeout time.Duration) {
	err := wait.PollImmediate(2*time.Second, timeout, func() (bool, error) {
		f.Log.Logf(f.t, "%s injecting", chaosName)

		chaos := chaosapis.PodChaos{}
		f.Client.Get(context.TODO(), client.ObjectKey{Name: chaosName, Namespace: f.kubectlOptions.Namespace}, &chaos)
		for _, cond := range chaos.Status.Conditions {
			if cond.Type == chaosapis.ConditionAllInjected && cond.Status == corev1.ConditionTrue {
				f.Log.Logf(f.t, "%s injected", chaosName)
				return true, nil
			}
		}
		return false, nil
	})
	Expect(err).Should(BeNil(), "failed to inject chaos")
}
