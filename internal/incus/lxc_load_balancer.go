package incus

import (
	"context"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha2"
)

// LoadBalancerManager can be used to interact with the cluster load balancer.
type LoadBalancerManager interface {
	// Create provisions the load balancer instance.
	// Implementations can indicate non-retriable failures (e.g. because of Incus not having the required extensions).
	// Callers must check these with incus.IsTerminalError() and treat them as terminal failures.
	Create(context.Context) ([]string, error)
	// Delete cleans up any load balancer resources.
	Delete(context.Context) error
	// Reconfigure updates the load balancer configuration based on the currently running control plane instances.
	Reconfigure(context.Context) error
}

// LoadBalancerManagerForCluster returns the proper LoadBalancerManager based on the lxcCluster spec.
func (c *Client) LoadBalancerManagerForCluster(lxcCluster *infrav1.LXCCluster) LoadBalancerManager {
	switch {
	case lxcCluster.Spec.LoadBalancer.LXC != nil:
		return &loadBalancerLXC{
			lxcClient:        c,
			clusterName:      lxcCluster.Name,
			clusterNamespace: lxcCluster.Namespace,

			name: lxcCluster.GetLoadBalancerInstanceName(),
			spec: lxcCluster.Spec.LoadBalancer.LXC.InstanceSpec,
		}
	case lxcCluster.Spec.LoadBalancer.OCI != nil:
		return &loadBalancerOCI{
			lxcClient:        c,
			clusterName:      lxcCluster.Name,
			clusterNamespace: lxcCluster.Namespace,

			name: lxcCluster.GetLoadBalancerInstanceName(),
			spec: lxcCluster.Spec.LoadBalancer.OCI.InstanceSpec,
		}
	case lxcCluster.Spec.LoadBalancer.OVN != nil:
		return &loadBalancerNetwork{
			lxcClient:        c,
			clusterName:      lxcCluster.Name,
			clusterNamespace: lxcCluster.Namespace,

			networkName:   lxcCluster.Spec.LoadBalancer.OVN.NetworkName,
			listenAddress: lxcCluster.Spec.ControlPlaneEndpoint.Host,
		}
	case lxcCluster.Spec.LoadBalancer.External != nil:
		return &loadBalancerExternal{
			lxcClient:        c,
			clusterName:      lxcCluster.Name,
			clusterNamespace: lxcCluster.Namespace,

			address: lxcCluster.Spec.ControlPlaneEndpoint.Host,
		}
	default:
		// TODO: handle this more gracefully.
		// If only Go had enums.
		return &loadBalancerExternal{
			lxcClient:        c,
			clusterName:      lxcCluster.Name,
			clusterNamespace: lxcCluster.Namespace,

			address: lxcCluster.Spec.ControlPlaneEndpoint.Host,
		}
	}
}
