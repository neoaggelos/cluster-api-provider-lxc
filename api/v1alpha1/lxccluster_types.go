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
	// ClusterFinalizer allows LXCClusterReconciler to clean up resources associated with LXCCluster before
	// removing it from the apiserver.
	ClusterFinalizer = "lxccluster.infrastructure.cluster.x-k8s.io"
)

// LXCClusterSpec defines the desired state of LXCCluster.
type LXCClusterSpec struct {
	// ControlPlaneEndpoint represents the endpoint to communicate with the control plane.
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`

	// TODO(neoaggelos): enable failure domains
	// FailureDomains clusterv1.FailureDomains `json:"failureDomains,omitempty"`
}

// LXCClusterStatus defines the observed state of LXCCluster.
type LXCClusterStatus struct {
	// Ready denotes that the LXC cluster (infrastructure) is ready.
	Ready bool `json:"ready"`

	// Conditions defines current service state of the LXCCluster.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Cluster infrastructure is ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of LXCCluster"

// LXCCluster is the Schema for the lxcclusters API.
type LXCCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LXCClusterSpec   `json:"spec,omitempty"`
	Status LXCClusterStatus `json:"status,omitempty"`
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
