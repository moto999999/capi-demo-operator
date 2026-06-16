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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1alpha1 "github.com/moto999999/capi-demo-operator/api/v1alpha1"
)

const machinePoolFinalizer = "infrastructure.capi.demo/machinepool-finalizer"

// MachinePoolReconciler reconciles a MachinePool object
type MachinePoolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.capi.demo,resources=machinepools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.capi.demo,resources=machinepools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.capi.demo,resources=machinepools/finalizers,verbs=update

func (r *MachinePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Step 1: Fetch the MachinePool instance
	mp := &infrastructurev1alpha1.MachinePool{}
	if err := r.Get(ctx, req.NamespacedName, mp); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get MachinePool: %w", err)
	}

	log.Info("Reconciling MachinePool",
		"name", mp.Name,
		"cluster", mp.Spec.ClusterName,
		"phase", mp.Status.Phase,
	)

	// Step 2: Handle deletion
	if !mp.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, mp)
	}

	// Step 3: Add finalizer if not present
	if !controllerutil.ContainsFinalizer(mp, machinePoolFinalizer) {
		controllerutil.AddFinalizer(mp, machinePoolFinalizer)
		if err := r.Update(ctx, mp); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// Step 4: Normal reconciliation
	return r.reconcileNormal(ctx, mp)
}

func (r *MachinePoolReconciler) reconcileNormal(ctx context.Context, mp *infrastructurev1alpha1.MachinePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Already ready, idempotent return (nothing to do)
	if mp.Status.Phase == "Ready" {
		return ctrl.Result{}, nil
	}

	// Set Provisioning and requeue — simulates async cloud API call
	// In a real provider this is where I'd call company's own cloud API
	if mp.Status.Phase == "" || mp.Status.Phase == "Pending" {
		mp.Status.Phase = "Provisioning"
		if err := r.Status().Update(ctx, mp); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set Provisioning phase: %w", err)
		}
		log.Info("Simulating cloud API call — provisioning server group")
		// Requeue after 5s to simulate cloud provisioning time
		// We'd normally requeue based on a watch on the cloud API or a status update, not a fixed time
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Provisioning → simulate cloud API response coming back
	if mp.Status.Phase == "Provisioning" {
		log.Info("Cloud API responded — server group ready")

		// This is what a real cloud API response would populate:
		// providerID = the cloud's native resource identifier
		// addresses  = IPs assigned by the cloud to the instances
		// networkID  = which cloud network they landed in
		mp.Status.ProviderID = fmt.Sprintf("company_name://server-group-%s", mp.Name)
		mp.Status.ReadyReplicas = mp.Spec.Replicas
		mp.Status.NetworkID = "net-demo-abc123"
		mp.Status.Addresses = generateFakeIPs(mp.Spec.Replicas)
		mp.Status.Phase = "Ready"

		if err := r.Status().Update(ctx, mp); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set Ready phase: %w", err)
		}
		log.Info("MachinePool ready",
			"providerID", mp.Status.ProviderID,
			"readyReplicas", mp.Status.ReadyReplicas,
		)
	}

	return ctrl.Result{}, nil
}

func (r *MachinePoolReconciler) reconcileDelete(ctx context.Context, mp *infrastructurev1alpha1.MachinePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Deleting MachinePool — would call cloud API to destroy server group", "name", mp.Name)

	// In a real provider: call cloud API to delete the server group
	// Only remove finalizer once cloud confirms deletion

	controllerutil.RemoveFinalizer(mp, machinePoolFinalizer)
	if err := r.Update(ctx, mp); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
	}
	return ctrl.Result{}, nil
}

// generateFakeIPs simulates the IP addresses a cloud API would assign
// In a real provider these come from the cloud API response
func generateFakeIPs(count int32) []string {
	ips := make([]string, count)
	for i := int32(0); i < count; i++ {
		ips[i] = fmt.Sprintf("10.0.1.%d", 10+i)
	}
	return ips
}

func (r *MachinePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1alpha1.MachinePool{}).
		Named("machinepool").
		Complete(r)
}
