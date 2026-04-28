package utils

import clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"

func ClusterFirstPodNetworkCIDR(in *clusterv1.Cluster) string {
	if pods := in.Spec.ClusterNetwork.Pods.CIDRBlocks; len(pods) > 0 {
		return pods[0]
	}
	return ""
}
