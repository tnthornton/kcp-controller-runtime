/*
Copyright 2018 The Kubernetes Authors.

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

package handler

import (
	"context"
	"reflect"

	"github.com/kcp-dev/logicalcluster/v3"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var enqueueLog = logf.RuntimeLog.WithName("eventhandler").WithName("EnqueueRequestForObject")

type empty struct{}

var _ EventHandler = &EnqueueRequestForObject{}

// EnqueueRequestForObject enqueues a Request containing the Name and Namespace of the object that is the source of the Event.
// (e.g. the created / deleted / updated objects Name and Namespace). handler.EnqueueRequestForObject is used by almost all
// Controllers that have associated Resources (e.g. CRDs) to reconcile the associated Resource.
type EnqueueRequestForObject = TypedEnqueueRequestForObject[client.Object]

// TypedEnqueueRequestForObject enqueues a Request containing the Name and Namespace of the object that is the source of the Event.
// (e.g. the created / deleted / updated objects Name and Namespace).  handler.TypedEnqueueRequestForObject is used by almost all
// Controllers that have associated Resources (e.g. CRDs) to reconcile the associated Resource.
//
// TypedEnqueueRequestForObject is experimental and subject to future change.
type TypedEnqueueRequestForObject[T client.Object] struct{}

// Create implements EventHandler.
func (e *TypedEnqueueRequestForObject[T]) Create(ctx context.Context, evt event.TypedCreateEvent[T], q workqueue.RateLimitingInterface) {
	if isNil(evt.Object) {
		enqueueLog.Error(nil, "CreateEvent received with no metadata", "event", evt)
		return
	}
	q.Add(request(evt.Object))
}

// Update implements EventHandler.
func (e *TypedEnqueueRequestForObject[T]) Update(ctx context.Context, evt event.TypedUpdateEvent[T], q workqueue.RateLimitingInterface) {
	switch {
	case !isNil(evt.ObjectNew):
		q.Add(request(evt.ObjectNew))
	case !isNil(evt.ObjectOld):
		q.Add(request(evt.ObjectOld))
	default:
		enqueueLog.Error(nil, "UpdateEvent received with no metadata", "event", evt)
	}
}

// Delete implements EventHandler.
func (e *TypedEnqueueRequestForObject[T]) Delete(ctx context.Context, evt event.TypedDeleteEvent[T], q workqueue.RateLimitingInterface) {
	if isNil(evt.Object) {
		enqueueLog.Error(nil, "DeleteEvent received with no metadata", "event", evt)
		return
	}
	q.Add(request(evt.Object))
}

// Generic implements EventHandler.
func (e *TypedEnqueueRequestForObject[T]) Generic(ctx context.Context, evt event.TypedGenericEvent[T], q workqueue.RateLimitingInterface) {
	if isNil(evt.Object) {
		enqueueLog.Error(nil, "GenericEvent received with no metadata", "event", evt)
		return
	}
	q.Add(request(evt.Object))
}

func request(obj client.Object) reconcile.Request {
	return reconcile.Request{
		// TODO(kcp) Need to implement a non-kcp-specific way to support this
		ClusterName: logicalcluster.From(obj).String(),
		NamespacedName: types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		},
	}
}

func isNil(arg any) bool {
	if v := reflect.ValueOf(arg); !v.IsValid() || ((v.Kind() == reflect.Ptr ||
		v.Kind() == reflect.Interface ||
		v.Kind() == reflect.Slice ||
		v.Kind() == reflect.Map ||
		v.Kind() == reflect.Chan ||
		v.Kind() == reflect.Func) && v.IsNil()) {
		return true
	}
	return false
}
