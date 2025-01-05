package incus

import (
	"context"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
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
	switch lxcCluster.Spec.LoadBalancer.Type {
	case "lxc", "":
		return &loadBalancerLXC{
			lxcClient:        c,
			clusterName:      lxcCluster.Name,
			clusterNamespace: lxcCluster.Namespace,
			serverType:       lxcCluster.Spec.ServerType,

			name: lxcCluster.GetLoadBalancerInstanceName(),
			spec: lxcCluster.Spec.LoadBalancer.InstanceSpec,
		}
	case "oci":
		return &loadBalancerOCI{
			lxcClient:        c,
			clusterName:      lxcCluster.Name,
			clusterNamespace: lxcCluster.Namespace,

			name: lxcCluster.GetLoadBalancerInstanceName(),
			spec: lxcCluster.Spec.LoadBalancer.InstanceSpec,
		}
	case "network":
		return &loadBalancerNetwork{
			lxcClient:        c,
			clusterName:      lxcCluster.Name,
			clusterNamespace: lxcCluster.Namespace,

			networkName:   lxcCluster.Spec.LoadBalancer.OVNNetworkName,
			listenAddress: lxcCluster.Spec.ControlPlaneEndpoint.Host,
		}
	case "external":
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
