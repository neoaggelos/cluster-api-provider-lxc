package incus

import (
	"context"
	"fmt"

	"github.com/lxc/incus/v6/shared/api"
	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CreateLoadBalancer creates the LoadBalancer container and returns the instance address.
func (c *Client) CreateLoadBalancer(ctx context.Context, cluster *infrav1.LXCCluster) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerCreateTimeout)
	defer cancel()

	if !c.Client.HasExtension("instance_oci") {
		// TODO(neoaggelos): find an alternative for LXD or Incus without OCI support
		return "", fmt.Errorf("server missing required 'instance_oci' extension, cannot create loadbalancer container")
	}
	name := fmt.Sprintf("%s-%s-lb", cluster.Namespace, cluster.Name)
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", name))
	if err := c.createInstanceIfNotExists(ctx, name, api.InstancesPost{
		Name: name,
		Type: api.InstanceTypeContainer,
		Source: api.InstanceSource{
			Type:     "image",
			Protocol: "oci",
			Server:   loadBalancerDefaultHaproxyImageRegistry,
			Mode:     "pull",
			Alias:    loadBalancerDefaultHaproxyImage,
		},
		InstancePut: api.InstancePut{
			Config: map[string]string{
				configClusterNameKey:      cluster.Name,
				configClusterNamespaceKey: cluster.Namespace,
				configInstanceRoleKey:     "loadbalancer",
			},
		},
	}); err != nil {
		return "", fmt.Errorf("failed to ensure loadbalancer instance exists: %w", err)
	}

	if err := c.ensureInstanceRunning(ctx, name); err != nil {
		return "", fmt.Errorf("failed to ensure loadbalancer instance is running: %w", err)
	}

	address, err := c.waitForInstanceAddress(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to get loadbalancer instance address: %w", err)
	}
	return address, nil
}
