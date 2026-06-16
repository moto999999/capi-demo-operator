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

// MachinePoolSpec defines the desired state of MachinePool
type MachinePoolSpec struct {
	// ClusterName references the parent ClusterRequest
	ClusterName string `json:"clusterName"`

	// Replicas is the desired number of instances in the pool
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas"`

	// MachineType is the instance size
	// +kubebuilder:validation:Enum=small;medium;large
	MachineType string `json:"machineType"`

	// MachineImage is the OS image to use for this pool
	// +kubebuilder:validation:Enum=ubuntu;rhel
	MachineImage string `json:"machineImage"`

	// AvailabilityZone is where the pool lives in the cloud
	AvailabilityZone string `json:"availabilityZone"`
}

// MachinePoolStatus defines the observed state of MachinePool.
type MachinePoolStatus struct {
	// Phase is the high-level lifecycle state
	// +kubebuilder:validation:Enum=Pending;Provisioning;Ready;Failed;Deleting
	Phase string `json:"phase,omitempty"`

	// ProviderID is the cloud's native identifier for this server
	// group (set AFTER cloud API responds)
	// +optional
	ProviderID string `json:"providerID,omitempty"`

	// ReadyReplicas is what the cloud reports as actually running
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// NetworkID is the cloud network this pool was placed in
	// Like the networkName you see in vSphere providerSpec
	// +optional
	NetworkID string `json:"networkID,omitempty"`

	// Addresses are the IPs of instances in the pool
	// Like Machine.status.addresses in OCP
	// +optional
	Addresses []string `json:"addresses,omitempty"`

	// Conditions is the standard conditions array
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MachinePool is the Schema for the machinepools API
type MachinePool struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of MachinePool
	// +required
	Spec MachinePoolSpec `json:"spec"`

	// status defines the observed state of MachinePool
	// +optional
	Status MachinePoolStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// MachinePoolList contains a list of MachinePool
type MachinePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []MachinePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &MachinePool{}, &MachinePoolList{})
		return nil
	})
}
