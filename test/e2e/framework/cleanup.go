/*
Copyright 2022 RadonDB.

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
	"os/exec"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/gruntwork-io/terratest/modules/k8s"
)

var (
	MySQLClusterCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: mysqlclusters.mysql.radondb.com
`
	MySQLUserCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: mysqlusers.mysql.radondb.com	
`
	MySQLBackUpCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: backups.mysql.radondb.com
`
)

func (f *Framework) CleanUpCRDs() {
	optional := f.kubectlOptions
	// CRD is cluster-scoped resources.
	optional.Namespace = ""
	IgnoreNotFound(k8s.KubectlDeleteFromStringE(f.t, optional, MySQLClusterCRD))
	IgnoreNotFound(k8s.KubectlDeleteFromStringE(f.t, optional, MySQLUserCRD))
	IgnoreNotFound(k8s.KubectlDeleteFromStringE(f.t, optional, MySQLBackUpCRD))
}

func CleanUpOperatorAtNS(ns string) {
	cmd1 := exec.Command("helm", "list", "--namespace", ns, "--short")
	cmd2 := exec.Command("xargs", "helm", "--namespace", ns, "delete")
	cmd2.Stdout = os.Stdout
	in, _ := cmd2.StdinPipe()
	if cmd1.Stdout == nil {
		return
	}
	cmd1.Stdout = in
	cmd2.Start()
	cmd1.Run()
	in.Close()

	Expect(cmd2.Wait()).Should(Succeed())
}

func (f *Framework) CleanUpNS() {
	IgnoreNotFound(k8s.DeleteNamespaceE(f.t, f.kubectlOptions, TestContext.E2ETestNamespace))
}

func (f *Framework) CleanUpChaos() {
	k8s.RunKubectl(f.t, f.kubectlOptions, "delete", "podchaos", "--all", "-n", f.kubectlOptions.Namespace)
}

func IgnoreNotFound(err error) {
	if err != nil && strings.Contains(err.Error(), "not found") {
		// Do nothing
	} else {
		Expect(err).ToNot(HaveOccurred())
	}
}
