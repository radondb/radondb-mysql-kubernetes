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
	"math"
	"strconv"
	"time"

	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apiv1alpha1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/internal"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster/container"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

func (s *StatefulSetSyncer) SfsReadOnly(ctx context.Context) error {
	if s.Spec.ReadOnlys == nil {
		currentStatefulset := appsv1.StatefulSet{}
		err := s.cli.Get(context.TODO(), types.NamespacedName{Name: getReadOnlyStatefulSetName(s), Namespace: s.Namespace}, &currentStatefulset)
		if err != nil && k8serrors.IsNotFound(err) {
			goto next1
		} else {
			s.cli.Delete(context.TODO(), &currentStatefulset)
			pvcs := corev1.PersistentVolumeClaimList{}
			if err := s.cli.List(ctx,
				&pvcs,
				&client.ListOptions{
					Namespace:     s.sfs.Namespace,
					LabelSelector: readOnlyLables(s).AsSelector(),
				},
			); err != nil {
				return err
			}

			for _, item := range pvcs.Items {
				if err := s.cli.Delete(ctx, &item); err != nil {
					return err
				}
			}
		}
	next1:
		// delete the readonly service
		service := corev1.Service{}
		err = s.cli.Get(ctx, types.NamespacedName{Name: s.GetNameForResource(utils.ReadOnlyHeadlessSVC), Namespace: s.Namespace}, &service)
		if err != nil && k8serrors.IsNotFound(err) {
			goto next2
		} else {
			s.cli.Delete(ctx, &service)
		}
	next2:
		serviceNodeport := corev1.Service{}
		err = s.cli.Get(ctx, types.NamespacedName{Name: s.GetNameForResource(utils.ReadOnlySvc), Namespace: s.Namespace}, &serviceNodeport)
		if err != nil && k8serrors.IsNotFound(err) {
			return nil
		} else {
			s.cli.Delete(ctx, &serviceNodeport)
		}
		return nil
	}

	if s.Status.State == apiv1alpha1.ClusterReadyState {
		// check and create the Statefulset
		readonlyStatefulSet, err := GetReadonlyStatefulSet(s)
		if err != nil {
			return errors.Errorf("get readonly deployment for cluster '%s': %v", s.Name, err)
		}
		if err := controllerutil.SetControllerReference(s.Unwrap(), readonlyStatefulSet, s.cli.Scheme()); err != nil {
			return err
		}
		//1. get Statefulset exist?
		currentStatefulset := appsv1.StatefulSet{}
		err = s.cli.Get(context.TODO(), types.NamespacedName{Name: readonlyStatefulSet.Name, Namespace: readonlyStatefulSet.Namespace}, &currentStatefulset)
		//2. if not exist, do nothing
		if err != nil && k8serrors.IsNotFound(err) {
			if err := s.cli.Create(context.TODO(), readonlyStatefulSet); err != nil && !k8serrors.IsAlreadyExists(err) {
				return errors.Wrapf(err, "create readonlhy statefulset for cluster '%s'", s.Name)
			}
		} else if err != nil {
			return errors.Wrapf(err, "get readonly deployment '%s'", readonlyStatefulSet.Name)
		}
		//3. update it
		currentStatefulset.Spec = readonlyStatefulSet.Spec
		if err := s.cli.Update(context.TODO(), &currentStatefulset); err != nil {
			// do expand pvc
			if k8serrors.IsInvalid(err) && ReadOnlyCanExtend(ctx, s, readonlyStatefulSet) {
				if err := ExtendReadOnlyPVCs(ctx, s, readonlyStatefulSet); err != nil {
					return errors.Wrapf(err, "extend readonly's pvc for cluster '%s'", s.Name)
				}
			} else {
				return errors.Wrapf(err, "update readonly statefulset '%s'", s.Name)
			}

		}
		// Update pvc.
		if err := deletePvcReadOnly(ctx, s); err != nil {
			return errors.Wrapf(err, "delete extra  readonly pvc '%s'", s.Name)
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

func GetReadonlyStatefulSet(cr *StatefulSetSyncer) (*appsv1.StatefulSet, error) {
	readonlySfs := getReadOnlyStatefulSetName(cr)

	initSidecar := container.EnsureContainer(utils.ContainerInitSidecarName, cr.MysqlCluster)
	initMysql := container.EnsureContainer(utils.ContainerInitMysqlName, cr.MysqlCluster)
	if cr.Spec.ReadOnlys.Resources != nil {
		initMysql.Resources = *cr.Spec.ReadOnlys.Resources
	}

	var containers []corev1.Container
	for _, v := range cr.sfs.Spec.Template.Spec.Containers {
		if v.Name != utils.ContainerXenonName && v.Name != utils.ContainerMysqlName {
			containers = append(containers, v)
		}
	}

	mysql := container.EnsureContainer(utils.ContainerMysqlName, cr.MysqlCluster)
	// ReadOnly mysql cannot use the mysqlchecker to do readness check
	mysql.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"sh",
					"-c",
					`if [ -f '/var/lib/mysql/sleep-forever' ] ;then exit 0 ; fi; test $(mysql -uroot -NB -e "SELECT 1") -eq 1`,
				},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      15,
		PeriodSeconds:       15,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	if cr.Spec.ReadOnlys.Resources != nil {
		mysql.Resources = *cr.Spec.ReadOnlys.Resources
		// (RO) calc innodb buffer and add it to env
		ib_pool, ib_inst, ib_logsize := cr.calcInnodbParam(cr.Spec.ReadOnlys.Resources)
		initSidecar.Env = append(initSidecar.Env,
			corev1.EnvVar{
				Name:  utils.ROIbPool,
				Value: ib_pool,
			},
			corev1.EnvVar{
				Name:  utils.ROIbInst,
				Value: ib_inst,
			},
			corev1.EnvVar{
				Name:  utils.ROIbLog,
				Value: ib_logsize,
			},
		)
	}
	initContainers := []corev1.Container{initSidecar, initMysql}
	containers = append(containers, mysql)
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      readonlySfs,
			Namespace: cr.Namespace,
			//OwnerReferences: cr.OwnerReferences, need set outside.
			Annotations: cr.Annotations,
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
					//Annotations: cr.sfs.Spec.Template.Annotations,
					Annotations: func() map[string]string {
						tmp := make(map[string]string)
						for k, v := range cr.sfs.Spec.Template.Annotations {
							tmp[k] = v
						}
						tmp["host-name"] = cr.Spec.ReadOnlys.Host
						return tmp
					}(),
				},
				Spec: corev1.PodSpec{
					InitContainers:     initContainers,
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

func getReadOnlyStatefulSetName(cr *StatefulSetSyncer) string {
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

		if status, err := internal.CheckSuperReadOnly(sqlRunner); err != nil {
			return err
		} else {
			if status != corev1.ConditionTrue {
				sqlRunner.QueryExec(internal.NewQuery("SET GLOBAL super_read_only=on"))
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
		var isReplicating corev1.ConditionStatus
		var err error
		if _, isReplicating, err = internal.CheckSlaveStatus(sqlRunner, s.Spec.ReplicaLag); err != nil {
			//Notice!!! this has error, just show error message, can not return.
			s.log.V(1).Info("slave status has gotten error", "error", err)
		}
		if isReplicating == corev1.ConditionFalse {
			// No.1 start slave
			if errStart := sqlRunner.QueryExec(internal.NewQuery("start slave;")); errStart != nil {
				s.log.V(1).Info("start slave gotten error", "error", errStart)
				// No2. change master and start
				changeSql := fmt.Sprintf(`stop slave;CHANGE MASTER TO MASTER_HOST='%s', MASTER_PORT=%d, MASTER_USER='%s', MASTER_PASSWORD='%s',
MASTER_AUTO_POSITION=1; start slave;`, buildMasterName(s), 3306, "root", cfg.Password)
				if err2 := sqlRunner.QueryExec(internal.NewQuery(changeSql)); err2 != nil {
					s.log.V(1).Info("change master and start slave gotten error", "error", err2)
				}
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
	if *cr.Spec.Replicas == 1 {
		return fmt.Sprintf("%s-0.%s.%s", cr.sfs.Spec.ServiceName, cr.sfs.Spec.ServiceName, cr.Namespace)
	}
	if len(cr.Spec.ReadOnlys.Host) == 0 {
		return fmt.Sprintf("%s-follower", cr.Name)
	} else {
		return fmt.Sprintf("%s.%s.%s", cr.Spec.ReadOnlys.Host, cr.sfs.Spec.ServiceName, cr.Namespace)
	}

}

func deletePvcReadOnly(ctx context.Context, s *StatefulSetSyncer) error {
	if s.Spec.ReadOnlys.Num == 0 {
		s.log.Info("skip update pvc because replicas is 0")
		return nil
	}
	pvcs := corev1.PersistentVolumeClaimList{}
	if err := s.cli.List(ctx,
		&pvcs,
		&client.ListOptions{
			Namespace:     s.sfs.Namespace,
			LabelSelector: readOnlyLables(s).AsSelector(),
		},
	); err != nil {
		return err
	}

	for _, item := range pvcs.Items {
		if item.DeletionTimestamp != nil {
			s.log.Info("pvc is being deleted", "pvc", item.Name, "key", s.Unwrap())
			continue
		}

		ordinal, err := utils.GetOrdinal(item.Name)
		if err != nil {
			s.log.Error(err, "pvc deletion error", "key", s.Unwrap())
			continue
		}

		if ordinal >= int(s.Spec.ReadOnlys.Num) {
			s.log.Info("cleaning up pvc", "pvc", item.Name, "key", s.Unwrap())
			if err := s.cli.Delete(ctx, &item); err != nil {
				return err
			}
		}
	}

	return nil
}

func ReadOnlyCanExtend(ctx context.Context, s *StatefulSetSyncer, roSfs *appsv1.StatefulSet) bool {
	// Get it again. Becaus it has been mutated in Sync.
	var cursfs = appsv1.StatefulSet{}
	if err := s.cli.Get(ctx, client.ObjectKeyFromObject(roSfs), &cursfs); err != nil {
		if k8serrors.IsNotFound(err) {
			return false
		}
	}
	oldRequest := cursfs.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.DeepCopy()

	newStorage := roSfs.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage()
	// If newStorage is not greater than oldStorage, do not expand.
	if newStorage.Cmp(*oldRequest.Storage()) != 1 {
		s.log.Info("read only pod ExpandPVC", "result", "can not expand", "reason", "new pvc is not larger than old pvc")
		return false
	}
	return true
}

func ExtendReadOnlyPVCs(ctx context.Context, s *StatefulSetSyncer, roSfs *appsv1.StatefulSet) error {
	// delelete statefulset
	if err := s.cli.Delete(ctx, roSfs); err != nil {
		return err
	}
	pvcs := corev1.PersistentVolumeClaimList{}
	if err := s.cli.List(ctx,
		&pvcs,
		&client.ListOptions{
			Namespace:     s.sfs.Namespace,
			LabelSelector: readOnlyLables(s).AsSelector(),
		},
	); err != nil {
		return err
	}

	for _, item := range pvcs.Items {
		// Notice: Only Resources.Requests can update, other field update will failure.
		// If storage Class's allowVolumeExpansion is false, update will failure.
		item.Spec.Resources.Requests = s.sfs.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests
		if err := s.cli.Update(ctx, &item); err != nil {
			return err
		}
		if err := wait.PollImmediate(time.Second*2, time.Duration(waitLimit)*time.Second, func() (bool, error) {
			// Check the pvc status.
			var currentPVC corev1.PersistentVolumeClaim
			if err2 := s.cli.Get(ctx, client.ObjectKeyFromObject(&item), &currentPVC); err2 != nil {
				return true, err2
			}
			var conditons = currentPVC.Status.Conditions
			// Notice: When expanding not start, or been completed, conditons is nil
			if conditons == nil {
				// If change storage request when replicas are creating, should check the currentPVC.Status.Capacity.
				// for example:
				// Pod0 has created successful,but Pod1 is creating. then change PVC from 20Gi to 30Gi .
				// Pod0's PVC need to expand, but Pod1's PVC has created as 30Gi, so need to skip it.
				if equality.Semantic.DeepEqual(currentPVC.Status.Capacity, item.Spec.Resources.Requests) {
					return true, nil
				}
				return false, nil
			}
			status := conditons[0].Type
			if status == "FileSystemResizePending" {
				return true, nil
			}
			return false, nil
		}); err != nil {
			return err
		}
	}
	if err := s.cli.Create(ctx, roSfs); err != nil {
		return err
	}
	return nil
}

func (s *StatefulSetSyncer) calcInnodbParam(resource *corev1.ResourceRequirements) (string, string, string) {
	const (
		_         = iota // ignore first value by assigning to blank identifier
		kb uint64 = 1 << (10 * iota)
		mb
		gb
	)
	var defaultSize, innodbBufferPoolSize uint64
	innodbBufferPoolSize = 128 * mb
	mem := uint64(resource.Requests.Memory().Value())
	cpu := resource.Limits.Cpu().MilliValue()
	if mem <= 1*gb {
		defaultSize = uint64(0.45 * float64(mem))
	} else {
		defaultSize = uint64(0.6 * float64(mem))
	}
	innodbBufferPoolSize = utils.Max(defaultSize, innodbBufferPoolSize)

	instances := math.Max(math.Min(math.Ceil(float64(cpu)/float64(1000)), math.Floor(float64(innodbBufferPoolSize)/float64(gb))), 1)
	// c.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_size"] = strconv.FormatUint(innodbBufferPoolSize, 10)
	// c.Spec.MysqlOpts.MysqlConf["innodb_buffer_pool_instances"] = strconv.Itoa(int(instances))

	// innodb_log_file_size = 25 % of innodb_buffer_pool_size
	// Minimum Value (≥ 5.7.11)	4194304
	// Minimum Value (≤ 5.7.10)	1048576
	// Maximum Value	512GB / innodb_log_files_in_group
	// but wet set it not over than 8G, you should set it in the config file when you want over 8G.
	// See https://dev.mysql.com/doc/refman/5.7/en/innodb-parameters.html#sysvar_innodb_log_file_size

	const innodbDefaultLogFileSize uint64 = 1073741824
	var innodbLogFileSize uint64 = innodbDefaultLogFileSize // 1GB, default value
	// if innodb_log_file_size is not set, calculate it
	logGroups, err := strconv.Atoi(s.Spec.MysqlOpts.MysqlConf["innodb_log_file_groups"])
	if err != nil {
		logGroups = 1
	}

	// https://dev.mysql.com/doc/refman/8.0/en/innodb-dedicated-server.html
	// Table 15.9 Automatically Configured Log File Size
	// Buffer Pool Size	Log File Size
	// Less than 8GB	512MiB
	// 8GB to 128GB	1024MiB
	// Greater than 128GB	2048MiB
	if innodbBufferPoolSize < (8 * gb) {
		innodbLogFileSize = (512 * mb) / (uint64(logGroups))
	} else if innodbBufferPoolSize <= (128 * gb) {
		innodbLogFileSize = 1 * gb
	} else {
		innodbLogFileSize = 2 * gb
	}
	return strconv.FormatUint(innodbBufferPoolSize, 10), strconv.Itoa(int(instances)), strconv.FormatUint(innodbLogFileSize, 10)
}
