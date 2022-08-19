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

package framework

import (
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apis "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
)

type Framework struct {
	t              testing.TestingT
	kubectlOptions *k8s.KubectlOptions

	BaseName  string
	Namespace *core.Namespace

	Client    client.Client
	ClientSet clientset.Interface

	Timeout time.Duration

	Log logr.Logger
}

func NewFramework(baseName string) *Framework {
	defer GinkgoRecover()

	f := &Framework{
		t: GinkgoT(),
		kubectlOptions: &k8s.KubectlOptions{
			Namespace:  DefaultE2ETestNS,
			ConfigPath: GetKubeconfig(),
		},
		BaseName: baseName,
		Namespace: &core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: DefaultE2ETestNS,
			},
		},
		Log: Log,
	}

	return f
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEach() {
	// The fact that we need this feels like a bug in ginkgo.
	// https://github.com/onsi/ginkgo/issues/222
	f.Timeout = time.Duration(TestContext.TimeoutSeconds) * time.Second

	By("creating a kubernetes client")
	cfg, err := LoadConfig()
	Expect(err).NotTo(HaveOccurred())

	apis.AddToScheme(scheme.Scheme)

	f.Client, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())

	f.ClientSet, err = clientset.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	f.Namespace = &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: DefaultE2ETestNS,
		},
	}
}

func (f *Framework) CreateNamespace(labels map[string]string) (*core.Namespace, error) {
	return CreateTestingNS(f.BaseName, f.ClientSet, labels)
}

// WaitForPodReady waits for the pod to flip to ready in the namespace.
func (f *Framework) WaitForPodReady(podName string) error {
	return waitTimeoutForPodReadyInNamespace(f.ClientSet, podName,
		f.Namespace.Name, PodStartTimeout)
}

func GetKubeconfig() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		u, err := user.Current()
		if err != nil {
			panic(err)
		}
		kubeconfig = filepath.Join(u.HomeDir, ".kube", "config")
		if _, err := os.Stat(kubeconfig); err != nil && !os.IsNotExist(err) {
			kubeconfig = ""
		}
	}
	return kubeconfig
}
