package incus

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

// DeleteInstance deletes the matching LXC instance, if any.
func (c *Client) DeleteInstance(ctx context.Context, lxcMachine *infrav1.LXCMachine) error {
	ctx, cancel := context.WithTimeout(ctx, instanceDeleteTimeout)
	defer cancel()

	name := lxcMachine.GetInstanceName()
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", name))

	return c.forceRemoveInstanceIfExists(ctx, name)
}
