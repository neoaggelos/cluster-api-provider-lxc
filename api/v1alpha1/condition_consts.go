package v1alpha1

import clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

// Conditions and condition Reasons for the LXCCluster object.

const (
	// LoadBalancerAvailableCondition documents the availability of the container that implements the cluster load balancer.
	//
	// NOTE: When the load balancer provisioning starts the process completes almost immediately and within
	// the same reconciliation, so the user will always see a transition from no condition to available without
	// having evidence that the operation is started/is in progress.
	LoadBalancerAvailableCondition clusterv1.ConditionType = "LoadBalancerAvailable"

	// LoadBalancerProvisioningFailedReason (Severity=Warning) documents a LXCCluster controller detecting
	// an error while provisioning the container that provides the cluster load balancer; those kind of
	// errors are usually transient and failed provisioning are automatically re-tried by the controller.
	LoadBalancerProvisioningFailedReason = "LoadBalancerProvisioningFailed"
)
