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
	"time"

	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/internal"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster/container"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

func (s *StatefulSetSyncer) SfsReadOnly(ctx context.Context) error {
	if s.Spec.ReadOnlys == nil {
		return nil
	}
	if s.Status.State == apiv1alpha1.ClusterReadyState {
		// check and create the Statefulset
		readonlyStatefulSet, err := GetReadonlyStatefulSet(s)
		if err != nil {
			return errors.Errorf("get readonly deployment for cluster '%s': %v", s.Name, err)
		}
		if err := controllerutil.SetControllerReference(s.Unwrap(), &readonlyStatefulSet, s.cli.Scheme()); err != nil {
			return err
		}
		//1. get Statefulset exist?
		currentStatefulset := appsv1.StatefulSet{}
		err = s.cli.Get(context.TODO(), types.NamespacedName{Name: readonlyStatefulSet.Name, Namespace: readonlyStatefulSet.Namespace}, &currentStatefulset)
		//2. if not exist, do nothing
		if err != nil && k8serrors.IsNotFound(err) {
			if err := s.cli.Create(context.TODO(), &readonlyStatefulSet); err != nil && !k8serrors.IsAlreadyExists(err) {
				return errors.Wrapf(err, "create readonlhy statefulset for cluster '%s'", s.Name)
			}
		} else if err != nil {
			return errors.Wrapf(err, "get readonly deployment '%s'", readonlyStatefulSet.Name)
		}
		//3. update it
		currentStatefulset.Spec = readonlyStatefulSet.Spec
		if err := s.cli.Update(context.TODO(), &currentStatefulset); err != nil {
			return errors.Wrapf(err, "update readonly statefulset '%s'", s.Name)
		}
		// 4. if all finished, do the change master
		if err = wait.PollImmediate(time.Second*5, time.Minute*2, func() (bool, error) {
			err = s.cli.Get(context.TODO(), types.NamespacedName{Name: readonlyStatefulSet.Name, Namespace: readonlyStatefulSet.Namespace}, &currentStatefulset)
			if err != nil {
				return false, err
			}
			if currentStatefulset.Status.ReadyReplicas == currentStatefulset.Status.Replicas {
				return true, nil
			} else {
				return false, nil
			}
		}); err != nil {
			return errors.Wrapf(err, "wait readonly statefulset ready '%s'", s.Name)
		}

		for i := 0; i < int(currentStatefulset.Status.Replicas); i++ {
			hostName := buildHostName(&currentStatefulset, i)
			if err := putMySQLReadOnly(s, hostName); err != nil {
				return errors.Wrapf(err, "set mysql to  readonly  '%s'", s.Name)
			}
		}

	}
	return nil
}

func GetReadonlyStatefulSet(cr *StatefulSetSyncer) (appsv1.StatefulSet, error) {
	readonlySfs := getReadOnlyStatefulSet(cr)

	var containers []corev1.Container
	for _, v := range cr.sfs.Spec.Template.Spec.Containers {
		if v.Name != utils.ContainerXenonName && v.Name != utils.ContainerMysqlName {
			containers = append(containers, v)
		}
	}
	mysql := container.EnsureContainer(utils.ContainerMysqlName, cr.MysqlCluster)
	if cr.Spec.ReadOnlys.Resources != nil {
		mysql.Resources = cr.Spec.MysqlOpts.Resources
	}
	containers = append(containers, mysql)
	return appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      readonlySfs,
			Namespace: cr.Namespace,
			//OwnerReferences: cr.OwnerReferences, need set outside.

		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &cr.Spec.ReadOnlys.Num,
			Selector: &metav1.LabelSelector{
				MatchLabels: readOnlyLables(cr),
			},
			ServiceName: cr.GetNameForResource(utils.ReadOnlyHeadlessSVC),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      readonlySfs,
					Namespace: cr.Namespace,
					Labels:    readOnlyLables(cr),
				},
				Spec: corev1.PodSpec{
					InitContainers:     cr.sfs.Spec.Template.Spec.InitContainers,
					Containers:         containers,
					Volumes:            cr.EnsureVolumes(),
					SchedulerName:      cr.Spec.PodPolicy.SchedulerName,
					ServiceAccountName: cr.GetNameForResource(utils.ServiceAccount),
					Affinity:           cr.Spec.ReadOnlys.Affinity,
					PriorityClassName:  cr.Spec.PodPolicy.PriorityClassName,
					Tolerations: func() []corev1.Toleration {
						if cr.Spec.ReadOnlys.Tolerations != nil {
							return cr.Spec.ReadOnlys.Tolerations
						} else {
							return cr.Spec.PodPolicy.Tolerations
						}
					}(),
				},
			},
			VolumeClaimTemplates: cr.sfs.Spec.VolumeClaimTemplates,
		},
	}, nil
}

func getReadOnlyStatefulSet(cr *StatefulSetSyncer) string {
	return cr.Name + "-ro"
}

func readOnlyLables(cr *StatefulSetSyncer) labels.Set {
	labels := cr.GetLabels()
	labels["app.kubernetes.io/name"] = "mysql-readonly"
	labels["readonly"] = "true"
	return labels
}

func putMySQLReadOnly(s *StatefulSetSyncer, host string) error {
	var sqlRunner internal.SQLRunner
	closeCh := make(chan func())

	var closeConn func()
	errCh := make(chan error)

	cfg, errOut := internal.NewConfigFromClusterKey(
		s.cli, s.MysqlCluster.GetClusterKey(), utils.RootUser, host)
	go func(sqlRunner *internal.SQLRunner, errCh chan error, closeCh chan func()) {
		var err error
		*sqlRunner, closeConn, err = s.SQLRunnerFactory(cfg, errOut)
		if err != nil {
			s.log.V(1).Info("failed to get sql runner", "error", err)
			errCh <- err
			return
		}
		if closeConn != nil {
			closeCh <- closeConn
			return
		}
		errCh <- nil
	}(&sqlRunner, errCh, closeCh)

	select {
	case errOut = <-errCh:
		return errOut
	case closeConn := <-closeCh:
		defer closeConn()
	case <-time.After(time.Second * 5):
	}
	if sqlRunner != nil {
		// 1. set it to readonly
		if status, err := internal.CheckReadOnly(sqlRunner); err != nil {
			return err
		} else {
			if status != corev1.ConditionTrue {
				sqlRunner.QueryExec(internal.NewQuery("SET GLOBAL read_only=on"))
			}
		}

		// 2. set rpl_semi_sync_slave_enabled off
		if status, err := internal.CheckSemSync(sqlRunner); err != nil {
			return err
		} else {
			if status == corev1.ConditionTrue {
				sqlRunner.QueryExec(internal.NewQuery("SET GLOBAL rpl_semi_sync_slave_enabled=off"))
			}
		}
		// 3. change master
		if _, isReplicating, err := internal.CheckSlaveStatus(sqlRunner); err != nil {
			return err
		} else {
			if isReplicating == corev1.ConditionFalse {
				// chang master
				//"CHANGE MASTER TO\n  " + strings.Join(args, ",\n  ")
				changeSql := fmt.Sprintf(`CHANGE MASTER TO MASTER_HOST='%s', MASTER_PORT=%d, MASTER_USER='%s', MASTER_PASSWORD='%s',
 MASTER_AUTO_POSITION=1; start slave;`, buildMasterName(s), 3306, "root", cfg.Password)
				sqlRunner.QueryExec(internal.NewQuery(changeSql))
			}
		}
	}
	return errOut
}

func buildHostName(cr *appsv1.StatefulSet, index int) string {
	return fmt.Sprintf("%s-%d.%s.%s", cr.Name, index, cr.Name, cr.Namespace)
}

func buildMasterName(cr *StatefulSetSyncer) string {
	// if the ReadOnlyType Host is nil
	if len(cr.Spec.ReadOnlys.Host) == 0 {
		return fmt.Sprintf("%s-follower", cr.Name)
	} else {
		return fmt.Sprintf("%s.%s.%s", cr.Spec.ReadOnlys.Host, cr.sfs.Spec.ServiceName, cr.Namespace)
	}

}
