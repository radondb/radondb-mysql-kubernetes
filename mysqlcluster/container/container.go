/*
Copyright 2021 RadonDB.

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

package container

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// container interface.
type container interface {
	getName() string
	getImage() string
	getCommand() []string
	getEnvVars() []corev1.EnvVar
	getLifecycle() *corev1.Lifecycle
	getResources() corev1.ResourceRequirements
	getPorts() []corev1.ContainerPort
	getLivenessProbe() *corev1.Probe
	getReadinessProbe() *corev1.Probe
	getVolumeMounts() []corev1.VolumeMount
}

func getStartupProbe(name string) *corev1.Probe {
	if name == utils.ContainerMysqlName {
		return &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"sh",
						"-c",
						`if test $(mysql -uroot -h127.0.0.1 -NB -e "SELECT 1") -eq 1; then sed -i "/^RESET SLAVE ALL;/d" /etc/mysql/conf.d/init.sql; fi`,
					},
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      5,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			FailureThreshold:    5,
		}
	} else {
		return nil
	}
}

// EnsureContainer ensure a container by the giving name.
func EnsureContainer(name string, c *mysqlcluster.MysqlCluster) corev1.Container {
	var ctr container
	var security *corev1.SecurityContext = nil
	switch name {
	case utils.ContainerInitSidecarName:
		ctr = &initSidecar{c, name}
	case utils.ContainerInitMysqlName:
		ctr = &initMysql{c, name}
	case utils.ContainerMysqlName:
		ctr = &mysql{c, name}
	case utils.ContainerXenonName:
		ctr = &xenon{c, name}
	case utils.ContainerMetricsName:
		ctr = &metrics{c, name}
	case utils.ContainerSlowLogName:
		ctr = &slowLog{c, name}
	case utils.ContainerAuditLogName:
		ctr = &auditLog{c, name}
	case utils.ContainerBackupName:
		ctr = &backupSidecar{c, name}
		needAdmin := true
		security = &corev1.SecurityContext{
			Privileged: &needAdmin,
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"CAP_SYS_ADMIN",
					"DAC_READ_SEARCH",
				},
			}}
	}

	return corev1.Container{
		Name:            ctr.getName(),
		Image:           mysqlcluster.GetImage(ctr.getImage()),
		ImagePullPolicy: c.Spec.PodPolicy.ImagePullPolicy,
		Command:         ctr.getCommand(),
		Env:             ctr.getEnvVars(),
		Lifecycle:       ctr.getLifecycle(),
		Resources:       ctr.getResources(),
		Ports:           ctr.getPorts(),
		LivenessProbe:   ctr.getLivenessProbe(),
		ReadinessProbe:  ctr.getReadinessProbe(),
		StartupProbe:    getStartupProbe(name),
		VolumeMounts:    ctr.getVolumeMounts(),
		SecurityContext: security,
	}
}
