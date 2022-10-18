package api

import (
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func pv(hostPath string, capacity string) *corev1.PersistentVolume {
	return &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "radondb-nfs-pv",
			Labels: map[string]string{
				"type": "local",
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(capacity + "Gi"),
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName: "radondb-nfs-hostpath",
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: hostPath,
				},
			},
		},
	}
}

func stc() *storagev1.StorageClass {
	return &storagev1.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StorageClass",
			APIVersion: "storage.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "radondb-nfs-hostpath",
		},
		Provisioner:       "kubernetes.io/no-provisioner",
		ReclaimPolicy:     toPVReclaimPolicyPtr("Retain"),
		VolumeBindingMode: toModePtr(storagev1.VolumeBindingWaitForFirstConsumer),
	}
}

func pvc() *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "radondb-nfs-pvc",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: toStringPtr("radondb-nfs-hostpath"),
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("30Gi"),
				},
			},
		},
	}
}

func toModePtr(m storagev1.VolumeBindingMode) *storagev1.VolumeBindingMode { return &m }

func toPVReclaimPolicyPtr(s string) *corev1.PersistentVolumeReclaimPolicy {
	t := corev1.PersistentVolumeReclaimPolicy(s)
	return &t
}

func toStringPtr(s string) *string { return &s }
