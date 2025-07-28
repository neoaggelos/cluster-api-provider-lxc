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
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/paused"
)

const (
	// ClusterFinalizer allows LXCClusterReconciler to clean up resources associated with LXCCluster before
	// removing it from the apiserver.
	ClusterFinalizer = "lxccluster.infrastructure.cluster.x-k8s.io"
)

// LXCClusterSpec defines the desired state of LXCCluster.
type LXCClusterSpec struct {
	// ControlPlaneEndpoint represents the endpoint to communicate with the control plane.
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`

	// SecretRef references a secret with credentials to access the LXC (e.g. Incus, LXD) server.
	SecretRef SecretRef `json:"secretRef,omitempty"`

	// LoadBalancer is configuration for provisioning the load balancer of the cluster.
	LoadBalancer LXCClusterLoadBalancer `json:"loadBalancer"`

	// Unprivileged will launch unprivileged LXC containers for the cluster machines.
	//
	// Known limitations apply for unprivileged LXC containers (e.g. cannot use NFS volumes).
	//
	// +optional
	Unprivileged bool `json:"unprivileged"`

	// Do not apply the default kubeadm profile on container instances.
	//
	// In this case, the cluster administrator is responsible to create the
	// profile manually and set the `.spec.template.spec.profiles` field of all
	// LXCMachineTemplate objects.
	//
	// For more details on the default kubeadm profile that is applied, see
	// https://capn.linuxcontainers.org/reference/profile/kubeadm.html
	//
	// +optional
	SkipDefaultKubeadmProfile bool `json:"skipDefaultKubeadmProfile"`

	// TODO(neoaggelos): enable failure domains
	// FailureDomains clusterv1.FailureDomains `json:"failureDomains,omitempty"`
}

// SecretRef is a reference to a secret in the cluster.
type SecretRef struct {
	// Name is the name of the secret to use. The secret must already exist in the same namespace as the parent object.
	Name string `json:"name"`
}

// LXCClusterLoadBalancer is configuration for provisioning the load balancer of the cluster.
//
// +kubebuilder:validation:MaxProperties:=1
// +kubebuilder:validation:MinProperties:=1
type LXCClusterLoadBalancer struct {
	// LXC will spin up a plain Ubuntu instance with haproxy installed.
	//
	// The controller will automatically update the list of backends on the haproxy configuration as control plane nodes are added or removed from the cluster.
	//
	// No other configuration is required for "lxc" mode. The load balancer instance can be configured through the .instanceSpec field.
	//
	// The load balancer container is a single point of failure to access the workload cluster control plane. Therefore, it should only be used for development or evaluation clusters.
	//
	// +optional
	LXC *LXCLoadBalancerInstance `json:"lxc,omitempty"`

	// OCI will spin up an OCI instance running the kindest/haproxy image.
	//
	// The controller will automatically update the list of backends on the haproxy configuration as control plane nodes are added or removed from the cluster.
	//
	// No other configuration is required for "oci" mode. The load balancer instance can be configured through the .instanceSpec field.
	//
	// The load balancer container is a single point of failure to access the workload cluster control plane. Therefore, it should only be used for development or evaluation clusters.
	//
	// Requires server extensions: `instance_oci`
	//
	// +optional
	OCI *LXCLoadBalancerInstance `json:"oci,omitempty"`

	// OVN will create a network load balancer.
	//
	// The controller will automatically update the list of backends for the network load balancer as control plane nodes are added or removed from the cluster.
	//
	// The cluster administrator is responsible to ensure that the OVN network is configured properly and that the LXCMachineTemplate objects have appropriate profiles to use the OVN network.
	//
	// When using the "ovn" mode, the load balancer address must be set in `.spec.controlPlaneEndpoint.host` on the LXCCluster object.
	//
	// Requires server extensions: `network_load_balancer`, `network_load_balancer_health_checks`
	//
	// +optional
	OVN *LXCLoadBalancerOVN `json:"ovn,omitempty"`

	// External will not create a load balancer. It must be used alongside something like kube-vip, otherwise the cluster will fail to provision.
	//
	// When using the "external" mode, the load balancer address must be set in `.spec.controlPlaneEndpoint.host` on the LXCCluster object.
	//
	// +optional
	External *LXCLoadBalancerExternal `json:"external,omitempty"`
}

type LXCLoadBalancerInstance struct {
	// InstanceSpec can be used to adjust the load balancer instance configuration.
	//
	// +optional
	InstanceSpec LXCLoadBalancerMachineSpec `json:"instanceSpec,omitempty"`
}

type LXCLoadBalancerOVN struct {
	// NetworkName is the name of the network to create the load balancer.
	NetworkName string `json:"networkName,omitempty"`
}

type LXCLoadBalancerExternal struct {
}

// LXCLoadBalancerMachineSpec is configuration for the container that will host the cluster load balancer, when using the "lxc" or "oci" load balancer type.
type LXCLoadBalancerMachineSpec struct {
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

	// Image to use for provisioning the load balancer machine. If not set,
	// a default image based on the load balancer type will be used.
	//
	//   - "oci": ghcr.io/lxc/cluster-api-provider-incus/haproxy:v20230606-42a2262b
	//   - "lxc": haproxy from the default simplestreams server
	//
	// +optional
	Image LXCMachineImageSource `json:"image"`

	// Target where the load balancer machine should be provisioned, when
	// infrastructure is a production cluster.
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
	Target string `json:"target,omitempty"`
}

// LXCClusterStatus defines the observed state of LXCCluster.
type LXCClusterStatus struct {
	// Ready denotes that the LXC cluster (infrastructure) is ready.
	//
	// +optional
	Ready bool `json:"ready"`

	// Conditions defines current service state of the LXCCluster.
	//
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`

	// V1Beta2 groups all status fields that will be added in LXCCluster's status with the v1beta2 version.
	//
	// +optional
	V1Beta2 *LXCClusterV1Beta2Status `json:"v1beta2,omitempty"`
}

// LXCClusterV1Beta2Status groups all the fields that will be added or modified in LXCCluster with the V1Beta2 version.
// See https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/proposals/20240916-improve-status-in-CAPI-resources.md for more context.
type LXCClusterV1Beta2Status struct {
	// conditions represents the observations of a LXCCluster's current state.
	// Known condition types are Ready, LoadBalancerAvailable, Deleting, Paused.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=32
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster"
// +kubebuilder:printcolumn:name="Load Balancer",type="string",JSONPath=".spec.controlPlaneEndpoint.host",description="Load Balancer address"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Cluster infrastructure is ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of LXCCluster"
// +kubebuilder:resource:categories=cluster-api

// LXCCluster is the Schema for the lxcclusters API.
type LXCCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LXCClusterSpec   `json:"spec,omitempty"`
	Status LXCClusterStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (c *LXCCluster) GetConditions() clusterv1.Conditions {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (c *LXCCluster) SetConditions(conditions clusterv1.Conditions) {
	c.Status.Conditions = conditions
}

// GetV1Beta2Conditions returns the set of conditions for this object.
func (c *LXCCluster) GetV1Beta2Conditions() []metav1.Condition {
	if c.Status.V1Beta2 == nil {
		return nil
	}
	return c.Status.V1Beta2.Conditions
}

// SetV1Beta2Conditions sets conditions for an API object.
func (c *LXCCluster) SetV1Beta2Conditions(conditions []metav1.Condition) {
	if c.Status.V1Beta2 == nil {
		c.Status.V1Beta2 = &LXCClusterV1Beta2Status{}
	}
	c.Status.V1Beta2.Conditions = conditions
}

// GetLXCSecretNamespacedName returns the client.ObjectKey for the secret containing LXC credentials.
func (c *LXCCluster) GetLXCSecretNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: c.Namespace,
		Name:      c.Spec.SecretRef.Name,
	}
}

// GetLoadBalancerInstanceName returns the instance name for the cluster load balancer.
func (c *LXCCluster) GetLoadBalancerInstanceName() string {
	// NOTE(neoaggelos): use first 5 chars of hex encoded sha256 sum of the namespace name.
	// This is because LXC instance names are limited to 63 characters.
	//
	// TODO(neoaggelos): in the future, consider using a generated name and metadata properties
	// to match the load balancer instance instead, such that we do not rely on magic instance names.
	// Load Balancer instances already have the following properties:
	//    user.cluster-name = Cluster.Name
	//    user.cluster-namespace = Cluster.Namespace
	//    user.role = "loadbalancer"
	hash := sha256.Sum256([]byte(c.Namespace))
	return fmt.Sprintf("%s-%s-lb", c.Name, hex.EncodeToString(hash[:3])[:5])
}

// +kubebuilder:object:root=true

// LXCClusterList contains a list of LXCCluster.
type LXCClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LXCCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LXCCluster{}, &LXCClusterList{})
}

var (
	_ paused.ConditionSetter = &LXCCluster{}
)
