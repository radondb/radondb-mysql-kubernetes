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
	"context"
	"fmt"
	"strconv"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/pkg/errors"
	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

/** Algorithm
1. check configmao exist?
2. if exist goto 3 else goto 7
3. create pvc
4. create job restore the pvc
5. set pvc owner.
6. delete configmap
7. exit, then create statefulset
*/

// 1. check configmap exist?
func (s *StatefulSetSyncer) checkConfigMap(ctx context.Context) (*corev1.ConfigMap, *apiv1alpha1.JuiceOpt, error) {
	// get config map
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.GetNameForResource(utils.RestoreCMN),
			Namespace: s.Namespace,
			Labels:    s.GetLabels(),
		},
	}
	err := s.cli.Get(ctx,
		types.NamespacedName{Namespace: s.Namespace,
			Name: s.GetNameForResource(utils.RestoreCMN)}, cm)

	if err != nil {
		return nil, nil, err
	}
	if f, ok := cm.Data["juice.opt"]; ok {
		juiceopt := &apiv1alpha1.JuiceOpt{}
		if err := yaml.Unmarshal([]byte(f), juiceopt); err != nil {
			return cm, nil, err
		} else {
			return cm, juiceopt, nil
		}
	} else {
		return nil, nil, fmt.Errorf("do not has %s", cm.Name)
	}

}

// 3.  create pvc
func (s *StatefulSetSyncer) createPvcs(ctx context.Context) error {
	logger := s.log
	replicas := *s.Spec.Replicas
	var i int32
	//var pvcarr []*corev1.PersistentVolumeClaim
	for i = 0; i < replicas; i++ {
		pvc := s.CreateOnePVC(fmt.Sprintf("data-%s-mysql-%d", s.Name, i))
		s.setPvcOwner(ctx, pvc)
		// Check exist
		err := s.cli.Get(ctx, types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, pvc)
		if err != nil && k8sErrors.IsNotFound(err) {
			logger.Info("Creating a new volume for restore", "Namespace", pvc.Namespace, "Name", pvc.Name)
			err = s.cli.Create(ctx, pvc)
			if err != nil {
				return errors.Wrap(err, "create restore pvc")
			}
		} else if err != nil {
			return errors.Wrap(err, "get restore pvc")
		}
		//pvcarr = append(pvcarr, pvc)
	}

	return nil
}

// get the screte data, and generate command string
func (s *StatefulSetSyncer) genCmdStr(ctx context.Context, from, fromDate string, juiceopt *apiv1alpha1.JuiceOpt) (string, error) {

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      juiceopt.BackupSecretName,
			Namespace: s.Namespace,
		},
	}
	err := s.cli.Get(ctx,
		types.NamespacedName{Namespace: s.Namespace,
			Name: juiceopt.BackupSecretName}, secret)

	if err != nil {
		return "", err
	}
	url, bucket := secret.Data["s3-endpoint"], secret.Data["s3-bucket"]

	accesskey, secretkey := secret.Data["s3-access-key"], secret.Data["s3-secret-key"]
	juicebucket := utils.InstallBucket(string(url), string(bucket))
	fmt.Println(url, bucket, accesskey, secretkey, juicebucket)
	cmdstr := fmt.Sprintf(`
	rm -rf /var/lib/mysql/*
	juicefs format --storage s3 \
    --bucket  %s \
    --access-key %s \
    --secret-key %s \
    %s \
    %s`, juicebucket, accesskey, secretkey, juiceopt.JuiceMeta, juiceopt.JuiceName)
	cmdstr += fmt.Sprintf(`
	juicefs mount -d %s /%s/
	`, juiceopt.JuiceMeta, juiceopt.JuiceName)
	cmdstr += fmt.Sprintf(`
	export CLUSTER_NAME=%s
    source /backup.sh 
    restore \"%s\" %s
	juicefs umount /%s/
	touch /var/lib/mysql/restore-file
	chown -R mysql.mysql /var/lib/mysql
`, from, fromDate, from, juiceopt.JuiceName)
	//fmt.Println(cmdstr)

	return cmdstr, nil
}

// create resetore job
func (s *StatefulSetSyncer) createRestoreJob(ctx context.Context) error {
	var cmds string
	if cm, juiceopt, err := s.checkConfigMap(ctx); err != nil {
		return err
	} else {
		if from, ok := cm.Data["from"]; ok {
			DateTime := time.Now().Format("2006-01-02 09:07:41")
			if D, nice := cm.Data["date"]; nice {
				// check where correct time
				DateTime = D
			}
			if cmds, err = s.genCmdStr(ctx, from, DateTime, juiceopt); err != nil {
				return err
			} else {
				//create pvcs
				if err = s.createPvcs(ctx); err != nil {
					return err
				}
			}

		}
	}

	envs := []corev1.EnvVar{
		{
			Name:  "CONTAINER_TYPE",
			Value: utils.ContainerBackupJobName,
		},
		{
			Name:  "NAMESPACE",
			Value: s.Namespace,
		},
		{
			Name:  "SERVICE_NAME",
			Value: fmt.Sprintf("%s-mysql", s.Name),
		},

		{
			Name:  "REPLICAS",
			Value: "1",
		},
	}
	jobarr := []*batchv1.Job{}
	for i := 0; i < int(*s.Spec.Replicas); i++ {
		job := &batchv1.Job{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "batch/v1",
				Kind:       "Job",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "restore-" + strconv.Itoa(i),
				Namespace: s.Namespace,
			},
		}
		job.Labels = map[string]string{

			"Type": "restore",

			// Cluster used as selector.
			"Cluster": s.Name,
		}

		Containers := make([]corev1.Container, 1)
		Containers[0].Name = "restore"
		Containers[0].Image = s.Spec.PodPolicy.SidecarImage
		Volumes := []corev1.Volume{
			{
				Name: fmt.Sprintf("data-%s-mysql-%d", s.Name, i),
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: fmt.Sprintf("data-%s-mysql-%d", s.Name, i),
					},
				},
			},
			{
				Name: utils.SysFuseVolume,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/dev/fuse",
					},
				},
			},
		}
		Containers[0].VolumeMounts = []corev1.VolumeMount{

			{
				Name:      fmt.Sprintf("data-%s-mysql-%d", s.Name, i),
				MountPath: utils.DataVolumeMountPath,
			},
			{
				Name:      utils.SysFuseVolume,
				MountPath: utils.SysFuseVolumnMountPath,
			},
		}
		Containers[0].Env = envs
		Containers[0].Command = []string{"bash", "-c", "-x", cmds}
		Containers[0].SecurityContext = &corev1.SecurityContext{
			Privileged: func(i bool) *bool { return &i }(true),
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"CAP_SYS_ADMIN",
					"DAC_READ_SEARCH",
				},
			}}

		job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
		job.Spec.Template.Spec.ServiceAccountName = s.Name
		job.Spec.Template.Spec.Containers = Containers
		job.Spec.Template.Spec.Volumes = Volumes

		ownerRefs := s.sfs.GetOwnerReferences()
		job.SetOwnerReferences(ownerRefs)
		err := s.cli.Create(context.TODO(), job)
		if err != nil && !k8sErrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "create restore job")
		} else if err == nil {
			jobarr = append(jobarr, job)
			s.log.Info("Created a new restore job", "Namespace", job.Namespace, "Name", job.Name)
		}
	}
	// Wait all job complete
	count := 0
retry:
	for _, job := range jobarr {
		if err := s.cli.Get(context.TODO(),
			types.NamespacedName{Name: job.Name,
				Namespace: job.Namespace}, job); err != nil {
			return errors.Wrap(err, "get restore job")
		}
		switch {
		case job.Status.Active == 1:
			// it is running
			s.log.Info("job is running", "Namespace", job.Namespace, "Name", job.Name)
			time.Sleep(2 * time.Second)
			goto retry
		case job.Status.Succeeded == 1:
			count++
		case job.Status.Failed >= 1:
			return fmt.Errorf("restore job %s fail", job.Name)
		}
	}
	return nil
}

// 5. set pvc owner.
func (s *StatefulSetSyncer) setPvcOwner(ctx context.Context, pvc *corev1.PersistentVolumeClaim) error {
	ownerRefs := s.sfs.GetOwnerReferences()
	pvc.SetOwnerReferences(ownerRefs)
	return nil
}

func (s *StatefulSetSyncer) CreateOnePVC(name string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: s.Namespace,
		},
		Spec: s.sfs.Spec.VolumeClaimTemplates[0].Spec,
	}
}
