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
	"fmt"

	corev1 "k8s.io/api/core/v1"
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

	// SecretRef is a reference to a secret with credentials to access LXC (e.g. Incus, LXD) server.
	SecretRef corev1.SecretReference `json:"secretRef,omitempty"`

	// LoadBalancer is configuration for provisioning the load balancer of the cluster.
	LoadBalancer LXCClusterLoadBalancer `json:"loadBalancer"`

	// Running Kubernetes on LXC requires an LXC profile enabling privileged
	// containers and similar configuration. By default, a profile with name
	// "cluster-api-$namespace-$name" profile is created and associated with
	// all created Machines automatically.
	//
	// This option can be used to disable this behavior. In that case, the cluster
	// administrator is responsible to create the LXC profile and specify it in the
	// .spec.template.spec.profiles field of the LXCMachineTemplate objects.
	//
	// This is useful in cases where a limited project is used, which does not
	// allow privileged containers.
	//
	// +optional
	SkipDefaultKubeadmProfile bool `json:"skipDefaultKubeadmProfile"`

	// TODO(neoaggelos): enable failure domains
	// FailureDomains clusterv1.FailureDomains `json:"failureDomains,omitempty"`
}

// LXCClusterLoadBalancer is configuration for provisioning the load balancer of the cluster.
type LXCClusterLoadBalancer struct {
	// Type of load balancer to provision for the cluster.
	//
	//   - "lxc" will spin up a plain Ubuntu instance and install haproxy.
	//
	//     The controller will automatically update the list of backends on the
	//     haproxy configuration control plane nodes are added or removed from
	//     the cluster.
	//
	//     No other configuration is required for "lxc" mode. The load balancer
	//     instance can be configured through .spec.loadBalancer.instanceSpec.
	//
	//   - "external" will not create any load balancer. Should be used alongside
	//     something like kube-vip, otherwise the cluster will fail to provision.
	//
	//     When using "external" mode, the load balancer address must be set in
	//     .spec.controlPlaneEndpoint.host on the LXCCluster object.
	//
	//   - "oci" will spin up an OCI instance running haproxy using the kind
	//     haproxy image.
	//
	//     The controller will automatically update the list of backends on the
	//     haproxy configuration control plane nodes are added or removed from
	//     the cluster.
	//
	//     No other configuration is required for "oci" mode. The load balancer
	//     instance can be configured through .spec.loadBalancer.instanceSpec.
	//
	//     Requires server extensions: "instance_oci"
	//
	//   - "network" will create a network load balancer.
	//
	//     The controller will automatically update the list of backends on the
	//     haproxy configuration control plane nodes are added or removed from
	//     the cluster.
	//
	//     When using "network" mode, the load balancer address must be set in
	//     .spec.controlPlaneEndpoint.host on the LXCCluster object. In addition,
	//     the ovn network to use must be set in .spec.loadBalancer.ovnNetworkName.
	//     The cluster administrator is responsible to ensure that the OVN network
	//     is configured and that the LXCMachineTemplate objects have appropriate
	//     profiles to use the OVN network.
	//
	//     Requires server extensions: "network_load_balancer"
	//
	//     Optional server extensions: "network_load_balancer_health_checks"
	//
	// +kubebuilder:validation:Enum:=lxc;external;oci;network
	Type string `json:"type,omitempty"`

	// InstanceSpec can be used to adjust the load balancer instance when using the "lxc" or "oci" load balancer type.
	// +optional
	InstanceSpec LXCMachineSpec `json:"instanceSpec,omitempty"`

	// OVNNetworkName is the name of the OVN network to use when using the "network" load balancer type.
	// +optional
	OVNNetworkName string `json:"ovnNetworkName,omitempty"`
}

// LXCClusterStatus defines the observed state of LXCCluster.
type LXCClusterStatus struct {
	// Ready denotes that the LXC cluster (infrastructure) is ready.
	// +optional
	Ready bool `json:"ready"`

	// Conditions defines current service state of the LXCCluster.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`

	// FailureReason will be set in the event that there is a terminal problem reconciling the LXCCluster
	// and will contain a succinct value suitable for machine interpretation.
	//
	// This field should not be set for transitive errors that can be fixed automatically or with manual intervention,
	// but instead indicate that something is fundamentally wrong with the LXCCluster and that it cannot be recovered.
	// +optional
	FailureReason *string `json:"failureReason,omitempty"`

	// FailureMessage will be set in the event that there is a terminal problem reconciling the LXCCluster
	// and will contain a more verbose string suitable for logging and human consumption.
	//
	// This field should not be set for transitive errors that can be fixed automatically or with manual intervention,
	// but instead indicate that something is fundamentally wrong with the LXCCluster and that it cannot be recovered.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// V1Beta2 groups all status fields that will be added in LXCCluster's status with the v1beta2 version.
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
// It defaults to the namespace of the cluster, if that is not set in the secretRef.
func (c *LXCCluster) GetLXCSecretNamespacedName() types.NamespacedName {
	key := types.NamespacedName{
		Namespace: c.Spec.SecretRef.Namespace,
		Name:      c.Spec.SecretRef.Name,
	}
	if key.Namespace == "" {
		key.Namespace = c.Namespace
	}
	return key
}

// GetLoadBalancerInstanceName returns the instance name for the cluster load balancer.
func (c *LXCCluster) GetLoadBalancerInstanceName() string {
	return fmt.Sprintf("%s-%s-lb", c.Namespace, c.Name)
}

// GetProfileName returns the profile name for the cluster LXC machines.
func (c *LXCCluster) GetProfileName() string {
	return fmt.Sprintf("cluster-api-%s-%s", c.Namespace, c.Name)
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
