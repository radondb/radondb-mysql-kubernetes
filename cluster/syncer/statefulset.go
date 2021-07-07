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

package syncer

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/radondb/radondb-mysql-kubernetes/cluster"
	"github.com/radondb/radondb-mysql-kubernetes/cluster/container"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// NewStatefulSetSyncer returns statefulset syncer.
func NewStatefulSetSyncer(cli client.Client, c *cluster.Cluster) syncer.Interface {
	obj := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.StatefulSet),
			Namespace: c.Namespace,
		},
	}

	return syncer.NewObjectSyncer("StatefulSet", c.Unwrap(), obj, cli, func() error {
		obj.Spec.ServiceName = c.GetNameForResource(utils.StatefulSet)
		obj.Spec.Replicas = c.Spec.Replicas
		obj.Spec.Selector = metav1.SetAsLabelSelector(c.GetSelectorLabels())

		obj.Spec.Template.ObjectMeta.Labels = c.GetLabels()
		for k, v := range c.Spec.PodSpec.Labels {
			obj.Spec.Template.ObjectMeta.Labels[k] = v
		}
		obj.Spec.Template.ObjectMeta.Labels["role"] = "candidate"
		obj.Spec.Template.ObjectMeta.Labels["healthy"] = "no"

		obj.Spec.Template.Annotations = c.Spec.PodSpec.Annotations
		if len(obj.Spec.Template.ObjectMeta.Annotations) == 0 {
			obj.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
		}
		if c.Spec.MetricsOpts.Enabled {
			obj.Spec.Template.ObjectMeta.Annotations["prometheus.io/scrape"] = "true"
			obj.Spec.Template.ObjectMeta.Annotations["prometheus.io/port"] = fmt.Sprintf("%d", utils.MetricsPort)
		}

		err := mergo.Merge(&obj.Spec.Template.Spec, ensurePodSpec(c), mergo.WithTransformers(transformers.PodSpec))
		if err != nil {
			return err
		}
		// mergo will add new keys for Tolerations and keep the others instead of removing them
		obj.Spec.Template.Spec.Tolerations = c.Spec.PodSpec.Tolerations

		if c.Spec.Persistence.Enabled {
			if obj.Spec.VolumeClaimTemplates, err = c.EnsureVolumeClaimTemplates(cli.Scheme()); err != nil {
				return err
			}
		}
		return nil
	})
}

// ensurePodSpec used to ensure the podspec.
func ensurePodSpec(c *cluster.Cluster) corev1.PodSpec {
	initSidecar := container.EnsureContainer(utils.ContainerInitSidecarName, c)
	initMysql := container.EnsureContainer(utils.ContainerInitMysqlName, c)
	initContainers := []corev1.Container{initSidecar, initMysql}

	mysql := container.EnsureContainer(utils.ContainerMysqlName, c)
	xenon := container.EnsureContainer(utils.ContainerXenonName, c)
	containers := []corev1.Container{mysql, xenon}
	if c.Spec.MetricsOpts.Enabled {
		containers = append(containers, container.EnsureContainer(utils.ContainerMetricsName, c))
	}
	if c.Spec.PodSpec.SlowLogTail {
		containers = append(containers, container.EnsureContainer(utils.ContainerSlowLogName, c))
	}
	if c.Spec.PodSpec.SlowLogTail {
		containers = append(containers, container.EnsureContainer(utils.ContainerAuditLogName, c))
	}

	return corev1.PodSpec{
		InitContainers:     initContainers,
		Containers:         containers,
		Volumes:            c.EnsureVolumes(),
		SchedulerName:      c.Spec.PodSpec.SchedulerName,
		ServiceAccountName: c.GetNameForResource(utils.ServiceAccount),
		Affinity:           c.Spec.PodSpec.Affinity,
		PriorityClassName:  c.Spec.PodSpec.PriorityClassName,
		Tolerations:        c.Spec.PodSpec.Tolerations,
	}
}
