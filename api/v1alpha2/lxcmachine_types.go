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

package v1alpha2

import (
	"fmt"

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
	// ProviderID is the container name in ProviderID format (lxc:///<containername>).
	//
	// +optional
	ProviderID *string `json:"providerID,omitempty"`

	// InstanceType is "container" or "virtual-machine". Empty defaults to "container".
	//
	// InstanceType may also be set to "kind", in which case OCI containers using the kindest/node
	// images will be created. This requires server extensions: "instance_oci", "instance_oci_entrypoint".
	//
	// +kubebuilder:validation:Enum:=container;virtual-machine;kind;""
	// +optional
	InstanceType string `json:"instanceType,omitempty"`

	// Flavor is configuration for the instance size (e.g. t3.micro, or c2-m4).
	//
	// Examples:
	//
	//   - `t3.micro` -- match specs of an EC2 t3.micro instance
	//   - `c2-m4` -- 2 cores, 4 GB RAM
	//
	// +optional
	Flavor string `json:"flavor,omitempty"`

	// Profiles is a list of profiles to attach to the instance.
	//
	// +optional
	Profiles []string `json:"profiles,omitempty"`

	// Devices allows overriding the configuration of the instance disk or network.
	//
	// Device configuration must be formatted using the syntax "<device>,<key>=<value>".
	//
	// For example, to specify a different network for an instance, you can use:
	//
	// ```yaml
	//   # override device "eth0", to be of type "nic" and use network "my-network"
	//   devices:
	//   - eth0,type=nic,network=my-network
	// ```
	//
	// +optional
	Devices []string `json:"devices,omitempty"`

	// Config allows overriding instance configuration keys.
	//
	// Note that the provider will always set the following configuration keys:
	//
	// - "cloud-init.user-data": cloud-init config data
	// - "user.cluster-name": name of owning cluster
	// - "user.cluster-namespace": namespace of owning cluster
	// - "user.cluster-role": instance role (e.g. control-plane, worker)
	// - "user.machine-name": name of machine (should match instance hostname)
	//
	// See https://linuxcontainers.org/incus/docs/main/reference/instance_options/#instance-options
	// for details.
	//
	// +optional
	Config map[string]string `json:"config,omitempty"`

	// Image to use for provisioning the machine. If not set, a kubeadm image
	// from the default upstream simplestreams source will be used, based on
	// the version of the machine.
	//
	// Note that the default source does not support images for all Kubernetes
	// versions, refer to the documentation for more details on which versions
	// are supported and how to build a base image for any version.
	//
	// +optional
	Image LXCMachineImageSource `json:"image"`

	// Target where the machine should be provisioned, when infrastructure
	// is a production cluster.
	//
	// Can be one of:
	//
	//   - `name`: where `name` is the name of a cluster member.
	//   - `@name`: where `name` is the name of a cluster group.
	//
	// Target is ignored when infrastructure is single-node (e.g. for
	// development purposes).
	//
	// For more information on cluster groups, you can refer to https://linuxcontainers.org/incus/docs/main/explanation/clustering/#cluster-groups
	//
	// +optional
	Target string `json:"target"`
}

type LXCMachineImageSource struct {
	// Name is the image name or alias.
	//
	// Note that Incus and Canonical LXD use incompatible image servers. To help
	// mitigate this issue, the following image names are automatically recognized:
	//
	//   - "ubuntu:VERSION" (e.g. "ubuntu:24.04")
	//     - Incus: "images:ubuntu/VERSION/cloud" (from https://images.linuxcontainers.org)
	//     - LXD: 	"ubuntu:VERSION" (from https://cloud-images.ubuntu.com/releases)
	//
	//   - "debian:VERSION" (e.g. "debian:12")
	//     - Incus: "images:debian/VERSION/cloud" (from https://images.linuxcontainers.org)
	//     - LXD: 	"images:debian/VERSION/cloud" (from https://images.lxd.canonical.com)
	//
	//   - "images:IMAGE" (e.g. "images:almalinux/9/cloud")
	//     - Incus: "images:IMAGE" (from https://images.linuxcontainers.org)
	//     - LXD: 	"images:IMAGE" (from https://images.lxd.canonical.com)
	//
	// +optional
	Name string `json:"name"`

	// Fingerprint is the image fingerprint.
	//
	// +optional
	Fingerprint string `json:"fingerprint"`

	// Server is the remote server, e.g. "https://images.linuxcontainers.org"
	//
	// +optional
	Server string `json:"server,omitempty"`

	// Protocol is the protocol to use for fetching the image, e.g. "simplestreams".
	//
	// +optional
	Protocol string `json:"protocol,omitempty"`
}

func (s *LXCMachineImageSource) IsZero() bool {
	return s == nil || *s == LXCMachineImageSource{}
}

// LXCMachineStatus defines the observed state of LXCMachine.
type LXCMachineStatus struct {
	// Ready denotes that the LXC machine is ready.
	//
	// +optional
	Ready bool `json:"ready,omitempty"`

	// LoadBalancerConfigured will be set to true once for each control plane node, after the load balancer instance is reconfigured.
	//
	// +optional
	LoadBalancerConfigured bool `json:"loadBalancerConfigured,omitempty"`

	// Addresses is the list of addresses of the LXC machine.
	//
	// +optional
	Addresses []clusterv1.MachineAddress `json:"addresses"`

	// Conditions defines current service state of the LXCMachine.
	//
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`

	// V1Beta2 groups all status fields that will be added in LXCMachine's status with the v1beta2 version.
	//
	// +optional
	V1Beta2 *LXCMachineV1Beta2Status `json:"v1beta2,omitempty"`
}

// LXCMachineV1Beta2Status groups all the fields that will be added or modified in LXCMachine with the V1Beta2 version.
// See https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md for more context.
type LXCMachineV1Beta2Status struct {
	// conditions represents the observations of a LXCMachine's current state.
	// Known condition types are Ready, InstanceProvisioned, Deleting, Paused.
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
// +kubebuilder:resource:categories=cluster-api

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

func (c *LXCMachine) GetInstanceName() string {
	return c.Name
}

// GetExpectedProviderID returns the expected providerID that the Kubernetes node should have.
func (c *LXCMachine) GetExpectedProviderID() string {
	return fmt.Sprintf("lxc:///%s", c.GetInstanceName())
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
