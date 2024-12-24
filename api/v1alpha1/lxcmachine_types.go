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
	"sigs.k8s.io/cluster-api/util/paused"
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

	// InstanceType is configuration for the instance size (e.g. t3.micro, or c2-m4)
	// Examples:
	//   - `t3.micro` -- match specs of an EC2 t3.micro instance
	//   - `c2-m4` -- 2 cores, 4 GB RAM
	InstanceType string `json:"instanceType,omitempty"`

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

	// Conditions defines current service state of the LXCMachine.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`

	// V1Beta2 groups all status fields that will be added in LXCMachine's status with the v1beta2 version.
	V1Beta2 *LXCMachineV1Beta2Status `json:"v1beta2,omitempty"`
}

// LXCMachineV1Beta2Status groups all the fields that will be added or modified in LXCMachine with the V1Beta2 version.
// See https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md for more context.
type LXCMachineV1Beta2Status struct {
	// conditions represents the observations of a LXCMachine's current state.
	// Known condition types are Ready, LoadBalancerAvailable, Deleting, Paused.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`
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

// GetConditions returns the set of conditions for this object.
func (c *LXCMachine) GetConditions() clusterv1.Conditions {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (c *LXCMachine) SetConditions(conditions clusterv1.Conditions) {
	c.Status.Conditions = conditions
}

// GetV1Beta2Conditions returns the set of conditions for this object.
func (c *LXCMachine) GetV1Beta2Conditions() []metav1.Condition {
	if c.Status.V1Beta2 == nil {
		return nil
	}
	return c.Status.V1Beta2.Conditions
}

// SetV1Beta2Conditions sets conditions for an API object.
func (c *LXCMachine) SetV1Beta2Conditions(conditions []metav1.Condition) {
	if c.Status.V1Beta2 == nil {
		c.Status.V1Beta2 = &LXCMachineV1Beta2Status{}
	}
	c.Status.V1Beta2.Conditions = conditions
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

var (
	_ paused.ConditionSetter = &LXCMachine{}
)
