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

	"github.com/presslabs/controller-util/syncer"
	appsv1 "k8s.io/api/apps/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"github.com/go-logr/logr"

	"github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
	"github.com/radondb/radondb-mysql-kubernetes/mysqlcluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	Initialized = v1alpha1.ClusterInitialized
	AllReady    = v1alpha1.ClusterAllReady
	Avaliable   = v1alpha1.ClusterAvaliable
	InUpgrade   = v1alpha1.ClusterInUpgrade

	CondTrue    = metav1.ConditionTrue
	CondFalse   = metav1.ConditionFalse
	CondUnknown = metav1.ConditionUnknown
)

// StatusSyncer used to update the status.
type StatusSyncer struct {
	*mysqlcluster.MysqlCluster

	cli client.Client

	statefulset *appsv1.StatefulSet
	// Logger
	log logr.Logger
}

// NewStatusSyncer returns a pointer to StatusSyncer.
func NewStatusSyncer(c *mysqlcluster.MysqlCluster, cli client.Client) *StatusSyncer {
	return &StatusSyncer{
		MysqlCluster: c,
		cli:          cli,
		statefulset: &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.GetNameForResource(utils.StatefulSet),
				Namespace: c.Namespace,
			},
		},
		// SQLRunnerFactory: sqlRunnerFactory,
		// XenonExecutor:    xenonExecutor,
		log: logf.Log.WithName("syncer.StatusSyncer"),
	}
}

// Object returns the object for which sync applies.
func (s *StatusSyncer) Object() interface{} { return nil }

// GetObject returns the object for which sync applies
// Deprecated: use github.com/presslabs/controller-util/syncer.Object() instead.
func (s *StatusSyncer) GetObject() interface{} { return nil }

// Owner returns the object owner or nil if object does not have one.
func (s *StatusSyncer) ObjectOwner() runtime.Object { return s.MysqlCluster }

// GetOwner returns the object owner or nil if object does not have one.
// Deprecated: use github.com/presslabs/controller-util/syncer.ObjectOwner() instead.
func (s *StatusSyncer) GetOwner() runtime.Object { return s.MysqlCluster }

// Sync persists data into the external store.
func (s *StatusSyncer) Sync(ctx context.Context) (syncer.SyncResult, error) {
	// Get statefulset.
	if err := s.cli.Get(ctx, client.ObjectKeyFromObject(s.statefulset), s.statefulset); err != nil {
		return syncer.SyncResult{}, err
	}
	// Set the status.
	s.Status.SetStatefulSetStatus(&s.statefulset.Status)
	
	// TODO: conditions
	s.Status.SetConditions([]v1alpha1.MySQLClusterCondition{
		{
			Type:   Initialized,
			Status: CondTrue,
		},
		{
			Type:   AllReady,
			Status: CondTrue,
		},
		{
			Type:   Avaliable,
			Status: CondTrue,
		},
		{
			Type:   InUpgrade,
			Status: CondTrue,
		},
	})

	// TODO: state
	s.Status.SetState(v1alpha1.ClusterReadyState)

	// TODO: rebuild

	return syncer.SyncResult{}, nil
}
