/*
Copyright 2026.

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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1alpha1 "github.com/moto999999/capi-demo-operator/api/v1alpha1"
)

const clusterRequestFinalizer = "infrastructure.capi.demo/finalizer"

// ClusterRequestReconciler reconciles a ClusterRequest object
type ClusterRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.capi.demo,resources=clusterrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.capi.demo,resources=clusterrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.capi.demo,resources=clusterrequests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Initialize a logger with the context of this request. This allows us to log messages that are specific to this reconciliation loop.
	log := logf.FromContext(ctx)

	// Step 1: Fetch the ClusterRequest
	// req only gives us the name — we always have to GET the object ourselves
	cr := &infrastructurev1alpha1.ClusterRequest{}
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		if errors.IsNotFound(err) {
			// Object deleted before we could reconcile — nothing to do
			// This is normal, not an error
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get ClusterRequest: %w", err)
	}

	log.Info("Reconciling ClusterRequest",
		"name", cr.Name,
		"phase", cr.Status.Phase,
		"deletionTimestamp", cr.DeletionTimestamp,
	)

	// Step 2: Handle deletion
	// DeletionTimestamp is set by Kubernetes when someone runs kubectl delete
	// but the object won't actually go away until all finalizers are removed
	if !cr.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, cr)
	}

	// Step 3: Add finalizer if not present
	// controllerutil.ContainsFinalizer is a helper from controller-runtime
	if !controllerutil.ContainsFinalizer(cr, clusterRequestFinalizer) {
		controllerutil.AddFinalizer(cr, clusterRequestFinalizer)
		if err := r.Update(ctx, cr); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer: %w", err)
		}
		// Return here — the Update will trigger another reconcile
		// where we'll continue with provisioning
		return ctrl.Result{}, nil
	}

	// Step 4: Normal reconciliation
	return r.reconcileNormal(ctx, cr)
}

func (r *ClusterRequestReconciler) reconcileNormal(ctx context.Context, cr *infrastructurev1alpha1.ClusterRequest) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Set phase to Provisioning if not already set
	// Notice: we check current phase to stay idempotent
	if cr.Status.Phase == "" || cr.Status.Phase == "Pending" {
		cr.Status.Phase = "Provisioning"
		// Status is a subresource — must use Status().Update(), not Update()
		if err := r.Status().Update(ctx, cr); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update phase to Provisioning: %w", err)
		}
		log.Info("Set phase to Provisioning")
	}

	// TODO next: create ControlPlane and MachinePool child resources

	return ctrl.Result{}, nil
}

func (r *ClusterRequestReconciler) reconcileDelete(ctx context.Context, cr *infrastructurev1alpha1.ClusterRequest) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Handling deletion of ClusterRequest", "name", cr.Name)

	// Set phase to Deleting so users can see what's happening
	if cr.Status.Phase != "Deleting" {
		cr.Status.Phase = "Deleting"
		if err := r.Status().Update(ctx, cr); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update phase to Deleting: %w", err)
		}
	}

	// TODO next: delete child resources here before removing finalizer

	// Remove finalizer — this is what actually allows Kubernetes to delete the object
	controllerutil.RemoveFinalizer(cr, clusterRequestFinalizer)
	if err := r.Update(ctx, cr); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
	}

	log.Info("Finalizer removed, ClusterRequest will be deleted")
	return ctrl.Result{}, nil
}

func (r *ClusterRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1alpha1.ClusterRequest{}).
		Named("clusterrequest").
		Complete(r)
}
