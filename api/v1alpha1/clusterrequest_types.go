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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterRequestSpec defines the desired state of ClusterRequest
type ClusterRequestSpec struct {
	// Version is the Kubernetes version for this cluster
	// e.g. "1.34.3" — like ClusterVersion in OCP
	// +kubebuilder:validation:Pattern=`^\d+\.\d+\.\d+$`
	Version string `json:"version"`

	// ControlPlane describes the control plane configuration
	// Similar to ControlPlaneMachineSet in OCP
	ControlPlane ControlPlaneSpec `json:"controlPlane"`

	// Workers describes the worker machine pools
	// Similar to MachineSets in OCP — we can have multiple worker machine pools, each with its own configuration
	// +kubebuilder:validation:MinItems=1
	Workers []MachinePoolSpec `json:"workers"`
}

type ControlPlaneSpec struct {
	// Replicas is always 1 or 3 in real clusters (etcd quorum)
	// +kubebuilder:validation:Enum=1;3
	Replicas int32 `json:"replicas"`

	// MachineType is a fake "instance type" — like AWS m5.xlarge in IPI
	// +kubebuilder:validation:Enum=small;medium;large
	MachineType string `json:"machineType"`
}

// ClusterRequestStatus defines the observed state of ClusterRequest.
type ClusterRequestStatus struct {
	// Phase is the high-level lifecycle state
	// Similar to the "phase" field on a Machine object in OCP
	// +kubebuilder:validation:Enum=Pending;Provisioning;Ready;Failed;Deleting
	Phase string `json:"phase,omitempty"`

	// ControlPlaneReady mirrors what you see on Cluster.status.controlPlaneReady in CAPI
	ControlPlaneReady bool `json:"controlPlaneReady,omitempty"`

	// ReadyWorkers is how many worker Machines are Ready
	// Similar to checking MachineSet .status.readyReplicas
	ReadyWorkers int32 `json:"readyWorkers,omitempty"`

	// TotalWorkers is the sum of desired replicas across all pools
	TotalWorkers int32 `json:"totalWorkers,omitempty"`

	// Conditions is the standard Kubernetes conditions array
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ClusterRequest is the Schema for the clusterrequests API
type ClusterRequest struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ClusterRequest
	// +required
	Spec ClusterRequestSpec `json:"spec"`

	// status defines the observed state of ClusterRequest
	// +optional
	Status ClusterRequestStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ClusterRequestList contains a list of ClusterRequest
type ClusterRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ClusterRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &ClusterRequest{}, &ClusterRequestList{})
		return nil
	})
}
