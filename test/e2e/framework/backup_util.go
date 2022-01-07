/*
Copyright 2018 Pressinfra SRL

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
	"context"
	"fmt"
	"os"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gomegatypes "github.com/onsi/gomega/types"
	apiv1alpha "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func GetS3EndPointName() string {
	S3 := os.Getenv("S3ENDPOINT")
	if len(S3) == 0 {
		Logf("S3ENDPOINT not set! Backups tests will not work")
	}
	return S3
}

func GetS3AccessKey() string {
	S3AccessKey := os.Getenv("S3ACCESSKEY")
	if len(S3AccessKey) == 0 {
		Logf("S3ACCESSKEY not set! Backups tests will not work")
	}
	return S3AccessKey
}

func GetS3SecretKey() string {
	S3SecretKey := os.Getenv("S3SECRETKEY")
	if len(S3SecretKey) == 0 {
		Logf("S3SECRETKEY not set! Backups tests will not work")
	}
	return S3SecretKey
}

func (f *Framework) NewBackupSecret() *corev1.Secret {
	// s3-endpoint:
	// s3-access-key:
	// s3-secret-key:
	// s3-bucket:
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-backup-secret", f.BaseName),
			Namespace: f.Namespace.Name,
		},
		StringData: map[string]string{
			"s3-endpoint":   GetS3EndPointName(),
			"s3-access-key": GetS3AccessKey(),
			"s3-secret-key": GetS3SecretKey(),
			"s3-bucket":     "radondb-backups",
		},
		Type: corev1.SecretTypeOpaque,
	}
}

func NewBackup(cluster *apiv1alpha.MysqlCluster, hostname string) *apiv1alpha.Backup {
	return &apiv1alpha.Backup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		Spec: apiv1alpha.BackupSpec{
			ClusterName: cluster.Name,
			HostName:    hostname,
			Image:       TestContext.SidecarImage,
		},
	}
}

func (f *Framework) RefreshBackupFn(backup *apiv1alpha.Backup) func() *apiv1alpha.Backup {
	return func() *apiv1alpha.Backup {
		key := types.NamespacedName{
			Name:      backup.Name,
			Namespace: backup.Namespace,
		}
		b := &apiv1alpha.Backup{}
		f.Client.Get(context.TODO(), key, b)
		return b
	}
}

func HaveBackupComplete() gomegatypes.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"Status": MatchFields(IgnoreExtras, Fields{
			"Completed": Equal(true),
		})},
	))
}
