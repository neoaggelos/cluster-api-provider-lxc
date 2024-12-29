package incus

import (
	"context"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

// LoadBalancerManager can be used to interact with the cluster load balancer.
type LoadBalancerManager interface {
	// Create provisions the load balancer instance.
	// Implementations can indicate non-retriable failures (e.g. because of Incus not having the required extensions).
	// Callers must check these with IsLoadBalancerTerminalError and treat them as terminal failures.
	Create(context.Context) ([]string, error)
	// Delete cleans up any load balancer resources.
	Delete(context.Context) error
	// Reconfigure updates the load balancer configuration based on the currently running control plane instances.
	Reconfigure(context.Context) error
}

// LoadBalancerManagerForCluster returns the proper LoadBalancerManager based on the lxcCluster spec.
func (c *Client) LoadBalancerManagerForCluster(lxcCluster *infrav1.LXCCluster) LoadBalancerManager {
	return &loadBalancerOCI{
		lxcClient:        c,
		instanceName:     lxcCluster.GetLoadBalancerInstanceName(),
		clusterName:      lxcCluster.Name,
		clusterNamespace: lxcCluster.Namespace,
	}
}
