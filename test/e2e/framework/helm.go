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
	"fmt"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
)

const (
	RadonDBHelmRepoName           = "radondb"
	RadonDBHelmRepoURL            = "https://radondb.github.io/radondb-mysql-kubernetes/"
	RadonDBMySQLOperatorChartName = "radondb/mysql-operator"
	HealthCheckServiceName        = "mysql-operator-e2e-test"
	WebhookServiceName            = "radondb-mysql-webhook"
)

var OperatorHealthService = `
apiVersion: v1
kind: Service
metadata:
  name: %s
spec:
  selector:
    app: mysql-operator
  ports:
    - name: http
      port: 8081
      protocol: TCP
      targetPort: 8081
  type: ClusterIP
`

type OperatorOptions struct {
	releaseName string
}

func NewOperatorOptions() *OperatorOptions {
	o := OperatorOptions{}
	if o.releaseName == "" {
		o.releaseName = "e2e-demo"
	}
	return &o
}

func (f *Framework) AddRadonDBHelmRepo() {
	helm.AddRepo(f.t, &helm.Options{}, RadonDBHelmRepoName, RadonDBHelmRepoURL)
}

func (f *Framework) RemoveRadonDBHelmRepo() {
	helm.RemoveRepo(f.t, &helm.Options{}, RadonDBHelmRepoName)
}

func (f *Framework) InstallOperator(o *OperatorOptions) {
	helmOpts := helm.Options{}
	if f.kubectlOptions != nil {
		helmOpts.KubectlOptions = f.kubectlOptions
	}
	helm.Install(f.t, &helmOpts, RadonDBMySQLOperatorChartName, o.releaseName)
}

func (f *Framework) UninstallOperator(o *OperatorOptions) {
	helm.Delete(f.t, &helm.Options{}, o.releaseName, true)
}

func (f *Framework) CheckVersion() string {
	if TestContext.ExpectedVersion == "" {
		return ""
	}
	res, _ := helm.RunHelmCommandAndGetOutputE(f.t, &helm.Options{}, "search", "repo", RadonDBHelmRepoName, "--version", TestContext.ExpectedVersion)
	return res
}

func (f *Framework) CreateOperatorHealthCheckService() error {
	if err := k8s.KubectlApplyFromStringE(f.t, f.kubectlOptions, fmt.Sprintf(OperatorHealthService, HealthCheckServiceName)); err != nil {
		return err
	}
	return nil
}
