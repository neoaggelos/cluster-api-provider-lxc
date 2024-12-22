/*
Copyright 2024 Angelos Kolaitis.

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

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// MachineFinalizer allows ReconcileLXCMachine to clean up resources associated with LXCMachine before
	// removing it from the apiserver.
	MachineFinalizer = "lxcmachine.infrastructure.cluster.x-k8s.io"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LXCMachineSpec defines the desired state of LXCMachine.
type LXCMachineSpec struct {
	// ProviderID is the container name in ProviderID format (lxc:///<containername>)
	// +optional
	ProviderID string `json:"providerID,omitempty"`

	// Type is the type of instance to create (container or virtual machine).
	// +kubebuilder:validation:Enum:=container;virtual-machine
	Type string `json:"type,omitempty"`

	// Profiles is a list of profiles to attach to the instance.
	Profiles []string `json:"profiles,omitempty"`
}

// LXCMachineStatus defines the observed state of LXCMachine.
type LXCMachineStatus struct {
	// Ready denotes that the LXC machine is ready.
	Ready bool `json:"ready,omitempty"`

	// State is the LXC machine state.
	State string `json:"state,omitempty"`

	// Addresses is the list of addresses of the LXC machine.
	Addresses []clusterv1.MachineAddress `json:"addresses"`

	// Conditions defines current service state of the DockerMachine.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels['cluster\\.x-k8s\\.io/cluster-name']",description="Cluster"
// +kubebuilder:printcolumn:name="Machine",type="string",JSONPath=".metadata.ownerReferences[?(@.kind==\"Machine\")].name",description="Machine object which owns this LXCMachine"
// +kubebuilder:printcolumn:name="ProviderID",type="string",JSONPath=".spec.providerID",description="Provider ID"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Machine ready status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of LXCMachine"

// LXCMachine is the Schema for the lxcmachines API.
type LXCMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LXCMachineSpec   `json:"spec,omitempty"`
	Status LXCMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LXCMachineList contains a list of LXCMachine.
type LXCMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LXCMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LXCMachine{}, &LXCMachineList{})
}
