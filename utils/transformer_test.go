package utils_test

import (
	"github.com/imdario/mergo"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var _ = ginkgo.Describe("transformer", func() {
	var two int32 = 2
	var three int32 = 3

	templatePodSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "mysql",
				Image: "percona:latest",
				Command: []string{
					"cmd1",
				},
				Env: []corev1.EnvVar{
					{
						Name: "MYSQL_ROOT_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "sample-password",
								},
								Key: "MYSQL_ROOT_PASSWORD",
							},
						},
					},
				},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
				},
			},
			{
				Name:  "xenon",
				Image: "xenon:latest",
				Env: []corev1.EnvVar{
					{
						Name:  "oldenv",
						Value: "oldenv",
					},
				},
			},
		},
	}
	templateStsSpec := appsv1.StatefulSetSpec{
		Replicas: &two,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.OnDeleteStatefulSetStrategyType,
		},
		Template: corev1.PodTemplateSpec{
			Spec: templatePodSpec,
		},
	}

	ginkgo.It("init should successfully", func() {
		actualStsSpec := appsv1.StatefulSetSpec{}
		mergo.Merge(&actualStsSpec, templateStsSpec, mergo.WithTransformers(utils.StsSpec))

		gomega.Expect(actualStsSpec).Should(gomega.Equal(templateStsSpec))
	})

	ginkgo.It("merge n times should successfully", func() {
		actualStsSpec := appsv1.StatefulSetSpec{}
		mergo.Merge(&actualStsSpec, templateStsSpec, mergo.WithTransformers(utils.StsSpec))
		gomega.Expect(actualStsSpec.Template.Spec.Containers[0].Command[0]).Should(gomega.Equal("cmd1"))

		actualStsSpec.Template.Spec.Containers[0].Command = append(actualStsSpec.Template.Spec.Containers[0].Command, "cmd2")
		mergo.Merge(&actualStsSpec, templateStsSpec, mergo.WithTransformers(utils.StsSpec))
		mergo.Merge(&actualStsSpec, templateStsSpec, mergo.WithTransformers(utils.StsSpec))

		gomega.Expect(len(actualStsSpec.Template.Spec.Containers[0].Command)).Should(gomega.Equal(2))
		gomega.Expect(actualStsSpec.Template.Spec.Containers[0].Command[0]).Should(gomega.Equal("cmd1"))
		gomega.Expect(actualStsSpec.Template.Spec.Containers[0].Command[1]).Should(gomega.Equal("cmd2"))
	})

	ginkgo.It("add containers should successfully", func() {
		actualStsSpec := *templateStsSpec.DeepCopy()
		actualStsSpec.Template.Spec.Containers = append(actualStsSpec.Template.Spec.Containers, corev1.Container{Name: "test"})
		mergo.Merge(&actualStsSpec, templateStsSpec, mergo.WithTransformers(utils.StsSpec))

		gomega.Expect(len(actualStsSpec.Template.Spec.Containers)).Should(gomega.Equal(3))
		gomega.Expect(actualStsSpec.Template.Spec.Containers[2].Name).Should(gomega.Equal("test"))
	})

	ginkgo.It("modify envs should successfully", func() {
		actualStsSpec := *templateStsSpec.DeepCopy()
		actualStsSpec.Template.Spec.Containers[0].Env[0].ValueFrom = &corev1.EnvVarSource{}
		actualStsSpec.Template.Spec.Containers[1].Env = append(actualStsSpec.Template.Spec.Containers[1].Env, corev1.EnvVar{Name: "newenv", Value: "newenv"})
		mergo.Merge(&actualStsSpec, templateStsSpec, mergo.WithTransformers(utils.StsSpec))

		gomega.Expect(actualStsSpec.Template.Spec.Containers[0].Env[0].ValueFrom).ShouldNot(gomega.BeNil())
		gomega.Expect(len(actualStsSpec.Template.Spec.Containers[1].Env)).Should(gomega.Equal(2))
	})

	ginkgo.It("modify container image should not successfully", func() {
		actualStsSpec := *templateStsSpec.DeepCopy()
		actualStsSpec.Template.Spec.Containers[0].Image = "mysql:latest"
		mergo.Merge(&actualStsSpec, templateStsSpec, mergo.WithTransformers(utils.StsSpec))

		gomega.Expect(len(actualStsSpec.Template.Spec.Containers)).Should(gomega.Equal(2))
		gomega.Expect(actualStsSpec.Template.Spec.Containers[0].Image).ShouldNot(gomega.Equal("mysql:latest"))
	})

	ginkgo.It("merge replicas,updateStrategy should not successfully", func() {
		actualStsSpec := *templateStsSpec.DeepCopy()
		actualStsSpec.Replicas = &three
		actualStsSpec.UpdateStrategy = appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		}
		mergo.Merge(&actualStsSpec, templateStsSpec, mergo.WithTransformers(utils.StsSpec))

		gomega.Expect(actualStsSpec.Replicas).Should(gomega.Equal(&two))
		gomega.Expect(actualStsSpec.UpdateStrategy.Type).Should(gomega.Equal(appsv1.OnDeleteStatefulSetStrategyType))
	})

	ginkgo.It("modify resources should not successfully", func() {
		actualStsSpec := *templateStsSpec.DeepCopy()
		testResources := corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("3"),
			},
		}
		actualStsSpec.Template.Spec.Containers[0].Resources = testResources
		mergo.Merge(&actualStsSpec, templateStsSpec, mergo.WithTransformers(utils.StsSpec))

		gomega.Expect(actualStsSpec.Template.Spec.Containers[0].Resources.Limits).ShouldNot(gomega.Equal(testResources))
	})
})
