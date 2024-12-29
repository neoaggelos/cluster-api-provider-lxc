package incus

import (
	"context"

	"github.com/lxc/incus/v6/shared/api"
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
		clusterName:      lxcCluster.Name,
		clusterNamespace: lxcCluster.Namespace,

		name: lxcCluster.GetLoadBalancerInstanceName(),

		// TODO: make source configurable from lxcCluster spec
		source: api.InstanceSource{
			Type:     "image",
			Protocol: "oci",
			Server:   "https://docker.io",
			Mode:     "pull",
			Alias:    "kindest/haproxy:v20230606-42a2262b",
		},
	}

	// return &loadBalancerLXC{
	// 	lxcClient:        c,
	// 	clusterName:      lxcCluster.Name,
	// 	clusterNamespace: lxcCluster.Namespace,

	// 	name: lxcCluster.GetLoadBalancerInstanceName(),

	// 	// TODO: make source configurable from lxcCluster spec
	// 	source: api.InstanceSource{
	// 		Type:     "image",
	// 		Protocol: "simplestreams",
	// 		Server:   "https://images.linuxcontainers.org",
	// 		Mode:     "pull",
	// 		Alias:    "ubuntu/22.04",
	// 	},
	// }

	// TODO: make configurable from lxcCluster spec
	// return &loadBalancerNetwork{
	// 	lxcClient:        c,
	// 	clusterName:      lxcCluster.Name,
	// 	clusterNamespace: lxcCluster.Namespace,

	// 	networkName:   "ovn1",
	// 	listenAddress: lxcCluster.Spec.ControlPlaneEndpoint.Host,
	// }

	// TODO: make configurable from lxcCluster spec
	// return &loadBalancerExternal{
	// 	address: lxcCluster.Spec.ControlPlaneEndpoint.Host,
	// }
}
