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

package api

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func rc(nfsServerImage string) *corev1.ReplicationController {
	return &corev1.ReplicationController{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "radondb-nfs-server",
		},
		Spec: corev1.ReplicationControllerSpec{
			Replicas: toInt32Ptr(1),
			Selector: map[string]string{
				"role": "nfs-server",
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"role": "nfs-server",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container(nfsServerImage)},
					Volumes: []corev1.Volume{
						{
							Name: "nfs-export-fast",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "radondb-nfs-pvc",
								},
							},
						},
					},
				},
			},
		},
	}
}

func svc() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "radondb-nfs-server",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "nfs",
					Port: 2049,
				},
				{
					Name: "mountd",
					Port: 20048,
				},
				{
					Name: "rpcbind",
					Port: 111,
				},
			},
			Selector: map[string]string{
				"role": "nfs-server",
			},
		},
	}
}

func container(nfsServerImage string) corev1.Container {
	return corev1.Container{
		Name:  "nfs-server",
		Image: nfsServerImage,
		Ports: []corev1.ContainerPort{
			{
				Name:          "nfs",
				ContainerPort: 2049,
			},
			{
				Name:          "mountd",
				ContainerPort: 20048,
			},
			{
				Name:          "rpcbind",
				ContainerPort: 111,
			},
		},
		SecurityContext: &corev1.SecurityContext{Privileged: toBoolPtr(true)},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "nfs-export-fast",
				MountPath: "/exports",
			},
		},
	}
}

func toInt32Ptr(i int32) *int32 { return &i }

func toBoolPtr(b bool) *bool { return &b }
