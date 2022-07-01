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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestStatefulSetSyncer_sfsUpdated(t *testing.T) {
	mockSfs := &appsv1.StatefulSet{
		Spec: appsv1.StatefulSetSpec{
			Replicas: &[]int32{1}[0],
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "test1",
						},
					},
				},
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}
	updateReplicas := func() *appsv1.StatefulSet {
		sfs := mockSfs.DeepCopy()
		sfs.Spec.Replicas = &[]int32{2}[0]
		return sfs
	}
	updateTemplate := func() *appsv1.StatefulSet {
		sfs := mockSfs.DeepCopy()
		sfs.Spec.Template.Spec.Containers[0].Name = "test2"
		return sfs
	}
	updateVolume := func() *appsv1.StatefulSet {
		sfs := mockSfs.DeepCopy()
		sfs.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[v1.ResourceStorage] = resource.MustParse("2Gi")
		return sfs
	}
	volumeNotExist := func() *appsv1.StatefulSet {
		sfs := mockSfs.DeepCopy()
		sfs.Spec.VolumeClaimTemplates = nil
		return sfs
	}

	type fields struct {
		sfs *appsv1.StatefulSet
	}
	type args struct {
		existing *appsv1.StatefulSet
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "update replicas",
			fields: fields{
				sfs: updateReplicas(),
			},
			args: args{
				existing: mockSfs,
			},
			want: true,
		},
		{
			name: "update template",
			fields: fields{
				sfs: updateTemplate(),
			},
			args: args{
				existing: mockSfs,
			},
			want: true,
		},
		{
			name: "update volume size",
			fields: fields{
				sfs: updateVolume(),
			},
			args: args{
				existing: mockSfs,
			},
			want: true,
		},
		{
			name: "no change",
			fields: fields{
				sfs: mockSfs,
			},
			args: args{
				existing: mockSfs,
			},
			want: false,
		},
		{
			name: "volume not exist",
			fields: fields{
				sfs: mockSfs,
			},
			args: args{
				existing: volumeNotExist(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StatefulSetSyncer{
				sfs: tt.fields.sfs,
			}
			if got := s.sfsUpdated(tt.args.existing); got != tt.want {
				t.Errorf("StatefulSetSyncer.sfsUpdated() = %v, want %v", got, tt.want)
			}
		})
	}
}
