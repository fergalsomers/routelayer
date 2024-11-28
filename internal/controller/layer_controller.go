/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	routelayerv1 "github.com/fergalsomers/routelayer/api/v1"
	"github.com/go-logr/logr"
)

// LayerReconciler reconciles a Layer object
type LayerReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	IstioEnabled bool // Whether integration with istio should be enabled or not. Defaults to false.
}

const (
	// name of our custom finalizer
	RouteLayerFinalizer = `routelayer.io/finalizer`
	WaitingState        = "Waiting"
	ReadyState          = "Ready"
	ErrorState          = "Error"
)

// +kubebuilder:rbac:groups=routelayer.github.com,resources=layers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=routelayer.github.com,resources=layers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=routelayer.github.com,resources=layers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Layer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *LayerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Determine if this is a create/update or delete request
	// Get the Layer object
	log := logger.WithValues("layer", req.NamespacedName)

	layer := &routelayerv1.Layer{}
	if err := r.Get(ctx, req.NamespacedName, layer); err != nil {
		log.Error(err, "unable to fetch Layer")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if layer.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !controllerutil.ContainsFinalizer(layer, RouteLayerFinalizer) {
			controllerutil.AddFinalizer(layer, RouteLayerFinalizer)
			if err := r.Update(ctx, layer); err != nil {
				return ctrl.Result{}, err
			}
		}
		// Object is either being created or updated
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(layer, RouteLayerFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			// Object is either being created or updated
			cntrl, err := r.deleteLayer(ctx, layer, log)

			if err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried.
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(layer, RouteLayerFinalizer)
			if err := r.Update(ctx, layer); err != nil {
				return ctrl.Result{}, err
			}
			return cntrl, nil
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Reconciliation Logic goes here
	cntrl, err := r.createUpdateLayer(ctx, layer, log)
	if err != nil {
		// if fail to delete the external dependency here, return with error
		// so that it can be retried.
		return ctrl.Result{}, err
	}

	if err := r.Status().Update(ctx, layer); err != nil { // We need to update the status
		return ctrl.Result{}, err
	}
	return cntrl, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LayerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&routelayerv1.Layer{}).
		Named("layer").
		Complete(r)
}

const (
	defaultWait = time.Second * 60
)

func (r *LayerReconciler) createUpdateLayer(ctx context.Context, layer *routelayerv1.Layer, log logr.Logger) (ctrl.Result, error) {
	log.Info("create/update layer")

	var parent *routelayerv1.Layer

	// Does the parent exist - realistically we should have a watcher on this
	// Since we need to process all the parents first
	if layer.Spec.Parent != "" {
		parent = &routelayerv1.Layer{}
		if err := r.Get(ctx, types.NamespacedName{Name: layer.Spec.Parent}, parent); err != nil {
			msg := fmt.Sprintf("Parent Layer %s not found", layer.Spec.Parent)
			layer.Status.Message = msg
			layer.Status.State = WaitingState

			return ctrl.Result{
				RequeueAfter: defaultWait,
			}, nil
		}
	}

	if r.IstioEnabled {
		// do some istio stufff
	}

	layer.Status.Message = "Layer created"
	layer.Status.State = ReadyState

	log.Info("layer", "resourceVersion", layer.ObjectMeta.ResourceVersion, "state", layer.Status.State)

	return ctrl.Result{}, nil
}

func (r *LayerReconciler) deleteLayer(ctx context.Context, layer *routelayerv1.Layer, log logr.Logger) (ctrl.Result, error) {
	log.Info("deleting layer")
	return ctrl.Result{}, nil
}
