package incus

import (
	"context"
	"fmt"

	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

// CreateInstance creates the LXC instance based on configuration from the machine.
func (c *Client) CreateInstance(ctx context.Context, machine *clusterv1.Machine, lxcMachine *infrav1.LXCMachine, lxcCluster *infrav1.LXCCluster, cloudInit string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, instanceCreateTimeout)
	defer cancel()

	name := lxcMachine.GetInstanceName()

	role := "control-plane"
	if !util.IsControlPlaneMachine(machine) {
		role = "worker"
	}
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", name, "role", role))

	if err := c.createInstanceIfNotExists(ctx, api.InstancesPost{
		Name:         name,
		Type:         api.InstanceType(lxcMachine.Spec.Type),
		Source:       c.instanceSourceFromAPI(lxcMachine.Spec.Image),
		InstanceType: lxcMachine.Spec.Flavor,
		InstancePut: api.InstancePut{
			Profiles: lxcMachine.Spec.Profiles,
			Config: map[string]string{
				configClusterNameKey:      lxcCluster.Name,
				configClusterNamespaceKey: lxcCluster.Namespace,
				configInstanceRoleKey:     role,
				configCloudInitKey:        cloudInit,
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to ensure instance exists: %w", err)
	}

	if err := c.ensureInstanceRunning(ctx, name); err != nil {
		return nil, fmt.Errorf("failed to ensure loadbalancer instance is running: %w", err)
	}

	addrs, err := c.waitForInstanceAddress(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get loadbalancer instance address: %w", err)
	}
	return addrs, nil
}
