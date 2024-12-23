package incus

import (
	"context"
	"fmt"
	"strings"

	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

func (c *Client) DeleteLoadBalancer(ctx context.Context, cluster *infrav1.LXCCluster) error {
	name := fmt.Sprintf("%s-%s-lb", cluster.Namespace, cluster.Name)

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("container", name))

	state, _, err := c.Client.GetInstanceState(name)
	if err != nil {
		log.FromContext(ctx).Info("Instance does not exist")
		if strings.Contains(err.Error(), "Instance not found") {
			return nil
		}
		return fmt.Errorf("failed to GetInstanceState: %w", err)
	}

	if state.Pid != 0 {
		log.FromContext(ctx).WithValues("status", state.Status, "pid", state.Pid).Info("Stopping instance")
		if op, err := c.Client.UpdateInstanceState(name, api.InstanceStatePut{Action: "stop", Force: true}, ""); err != nil {
			return fmt.Errorf("failed to UpdateInstanceState: %w", err)
		} else if err := op.WaitContext(ctx); err != nil && !strings.Contains(err.Error(), "Operation not found") {
			return fmt.Errorf("failed to wait for UpdateInstanceState operation: %w", err)
		}
	}

	log.FromContext(ctx).Info("Deleting instance")
	if op, err := c.Client.DeleteInstance(name); err != nil && !strings.Contains(err.Error(), "Instance not found") {
		return fmt.Errorf("failed to DeleteInstance: %w", err)
	} else if op == nil {
		return nil
	} else if op.WaitContext(ctx); err != nil {
		return fmt.Errorf("failed to wait for DeleteInstance operation: %w", err)
	}

	return nil
}
