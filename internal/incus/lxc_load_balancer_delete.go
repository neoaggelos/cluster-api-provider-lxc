package incus

import (
	"context"
	"fmt"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DeleteLoadBalancer deletes the LoadBalancer container if it is running.
func (c *Client) DeleteLoadBalancer(ctx context.Context, cluster *infrav1.LXCCluster) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerDeleteTimeout)
	defer cancel()

	name := fmt.Sprintf("%s-%s-lb", cluster.Namespace, cluster.Name)
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", name))

	return c.forceRemoveInstanceIfExists(ctx, name)
}
