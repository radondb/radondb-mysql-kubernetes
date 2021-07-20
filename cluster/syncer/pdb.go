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
	"github.com/presslabs/controller-util/syncer"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/radondb/radondb-mysql-kubernetes/cluster"
	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

// NewPDBSyncer returns podDisruptionBudget syncer.
func NewPDBSyncer(cli client.Client, c *cluster.Cluster) syncer.Interface {
	pdb := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.GetNameForResource(utils.PodDisruptionBudget),
			Namespace: c.Namespace,
		},
	}

	return syncer.NewObjectSyncer("PDB", c.Unwrap(), pdb, cli, func() error {
		if pdb.Spec.MinAvailable != nil {
			// this mean that pdb is created and should return because spec is imutable
			return nil
		}
		ma := intstr.FromString(c.Spec.MinAvailable)
		pdb.Spec.MinAvailable = &ma
		pdb.Spec.Selector = metav1.SetAsLabelSelector(c.GetSelectorLabels())
		return nil
	})
}
