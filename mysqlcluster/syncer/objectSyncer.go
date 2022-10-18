/*
Copyright 2022 RadonDB.
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
	"errors"
	"fmt"

	"github.com/go-test/deep"
	"github.com/iancoleman/strcase"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	eventNormal  = "Normal"
	eventWarning = "Warning"

	resultNone    = controllerutil.OperationResultNone
	resultCreated = controllerutil.OperationResultCreated
	resultUpdated = controllerutil.OperationResultUpdated
)

var (
	// ErrOwnerDeleted is returned when the object owner is marked for deletion.
	ErrOwnerDeleted = fmt.Errorf("owner is deleted")

	// ErrIgnore is returned for ignored errors.
	// Ignored errors are treated by the syncer as successful syncs.
	ErrIgnore = fmt.Errorf("ignored error")
)

// Interface represents a syncer. A syncer persists an object
// (known as subject), into a store (kubernetes apiserver or generic stores)
// and records kubernetes events.
type Syncer interface {
	// Object returns the object for which sync applies.
	Object() interface{}

	// Owner returns the object owner or nil if object does not have one.
	ObjectOwner() runtime.Object

	// Sync persists data into the external store.
	Sync(context.Context) (SyncResult, error)
}

// SyncResult is a result of an Sync call.
type SyncResult struct {
	Operation    controllerutil.OperationResult
	EventType    string
	EventReason  string
	EventMessage string
}

// SetEventData sets event data on an SyncResult.
func (r *SyncResult) SetEventData(eventType, reason, message string) {
	r.EventType = eventType
	r.EventReason = reason
	r.EventMessage = message
}

// ObjectSyncer is a syncer.Interface for syncing kubernetes.Objects only by
// passing a SyncFn.
type ObjectSyncer struct {
	Owner     client.Object
	Obj       client.Object
	SyncFn    controllerutil.MutateFn
	Name      string
	Client    client.Client
	ForceSync bool

	previousObject runtime.Object
}

// NewObjectSyncer creates a new kubernetes.Object syncer for a given object
// with an owner and persists data using controller-runtime's CreateOrUpdate.
// The name is used for logging and event emitting purposes and should be an
// valid go identifier in upper camel case. (eg. MysqlStatefulSet).
func NewObjectSyncer(name string, owner, obj client.Object, c client.Client, forceSync bool, syncFn controllerutil.MutateFn) Syncer {
	return &ObjectSyncer{
		Owner:     owner,
		Obj:       obj,
		SyncFn:    syncFn,
		Name:      name,
		Client:    c,
		ForceSync: forceSync,
	}
}

// Object returns the ObjectSyncer subject.
func (s *ObjectSyncer) Object() interface{} {
	return s.Obj
}

// ObjectOwner returns the ObjectSyncer owner.
func (s *ObjectSyncer) ObjectOwner() runtime.Object {
	return s.Owner
}

// Sync does the actual syncing and implements the syncer.Inteface Sync method.
func (s *ObjectSyncer) Sync(ctx context.Context) (SyncResult, error) {
	var err error

	result := SyncResult{}
	log := logf.FromContext(ctx, "syncer", s.Name)
	key := client.ObjectKeyFromObject(s.Obj)

	if s.ForceSync {
		result.Operation, err = controllerutil.CreateOrUpdate(ctx, s.Client, s.Obj, s.mutateFn())
	} else {
		result.Operation, err = CreateIfNotExist(ctx, s.Client, s.Obj, s.mutateFn())
	}

	// check deep diff
	diff := deep.Equal(redact(s.previousObject), redact(s.Obj))

	// don't pass to user error for owner deletion, just don't create the object
	// nolint: gocritic
	if errors.Is(err, ErrOwnerDeleted) {
		log.Info(string(result.Operation), "key", key, "kind", s.objectType(s.Obj), "error", err)
		err = nil
	} else if errors.Is(err, ErrIgnore) {
		log.V(1).Info("syncer skipped", "key", key, "kind", s.objectType(s.Obj), "error", err)
		err = nil
	} else if err != nil {
		result.SetEventData(eventWarning, basicEventReason(s.Name, err),
			fmt.Sprintf("%s %s failed syncing: %s", s.objectType(s.Obj), key, err))
		log.Error(err, string(result.Operation), "key", key, "kind", s.objectType(s.Obj), "diff", diff)
	} else {
		result.SetEventData(eventNormal, basicEventReason(s.Name, err),
			fmt.Sprintf("%s %s %s successfully", s.objectType(s.Obj), key, result.Operation))
		log.V(1).Info(string(result.Operation), "key", key, "kind", s.objectType(s.Obj), "diff", diff)
	}

	return result, err
}

func CreateIfNotExist(ctx context.Context, c client.Client, obj client.Object, f controllerutil.MutateFn) (controllerutil.OperationResult, error) {
	var err error
	if err = c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		if !k8serrors.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}

		if err = f(); err != nil {
			return controllerutil.OperationResultNone, err
		}

		if err = c.Create(ctx, obj); err != nil {
			return controllerutil.OperationResultNone, err
		} else {
			return controllerutil.OperationResultCreated, nil
		}
	}
	return controllerutil.OperationResultNone, nil
}

// objectType returns the type of a runtime.Object.
func (s *ObjectSyncer) objectType(obj runtime.Object) string {
	if obj != nil {
		gvk, err := apiutil.GVKForObject(obj, s.Client.Scheme())
		if err != nil {
			return fmt.Sprintf("%T", obj)
		}

		return gvk.String()
	}

	return "nil"
}

// Given an ObjectSyncer, returns a controllerutil.MutateFn which also sets the
// owner reference if the subject has one.
func (s *ObjectSyncer) mutateFn() controllerutil.MutateFn {
	return func() error {
		s.previousObject = s.Obj.DeepCopyObject()

		err := s.SyncFn()
		if err != nil {
			return err
		}

		if s.Owner == nil {
			return nil
		}

		// set owner reference only if owner resource is not being deleted, otherwise the owner
		// reference will be reset in case of deleting with cascade=false.
		if s.Owner.GetDeletionTimestamp().IsZero() {
			if err := controllerutil.SetControllerReference(s.Owner, s.Obj, s.Client.Scheme()); err != nil {
				return err
			}
		} else if ctime := s.Obj.GetCreationTimestamp(); ctime.IsZero() {
			// the owner is deleted, don't recreate the resource if does not exist, because gc
			// will not delete it again because has no owner reference set
			return ErrOwnerDeleted
		}

		return nil
	}
}

func basicEventReason(objKindName string, err error) string {
	if err != nil {
		return fmt.Sprintf("%sSyncFailed", strcase.ToCamel(objKindName))
	}

	return fmt.Sprintf("%sSyncSuccessfull", strcase.ToCamel(objKindName))
}

// Redacts sensitive data from runtime.Object making them suitable for logging.
func redact(obj runtime.Object) runtime.Object {
	switch exposed := obj.(type) {
	case *corev1.Secret:
		redacted := exposed.DeepCopy()
		redacted.Data = nil
		redacted.StringData = nil
		exposed.ObjectMeta.DeepCopyInto(&redacted.ObjectMeta)

		return redacted
	case *corev1.ConfigMap:
		redacted := exposed.DeepCopy()
		redacted.Data = nil

		return redacted
	}

	return obj
}

// Sync mutates the subject of the syncer interface using controller-runtime
// CreateOrUpdate method, when obj is not nil. It takes care of setting owner
// references and recording kubernetes events where appropriate.
func Sync(ctx context.Context, syncer Syncer, recorder record.EventRecorder) error {
	result, err := syncer.Sync(ctx)
	owner := syncer.ObjectOwner()

	if recorder != nil && owner != nil && result.EventType != "" && result.EventReason != "" && result.EventMessage != "" {
		if err != nil || result.Operation != controllerutil.OperationResultNone {
			recorder.Eventf(owner, result.EventType, result.EventReason, result.EventMessage)
		}
	}

	return err
}
