package incus

import (
	"context"
	"fmt"
	"slices"

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

	instanceType := c.instanceTypeFromAPI(lxcMachine.Spec.Type)

	profiles := lxcMachine.Spec.Profiles
	if instanceType == api.InstanceTypeContainer && !lxcCluster.Spec.SkipDefaultKubeadmProfile && !slices.Contains(lxcMachine.Spec.Profiles, lxcCluster.GetProfileName()) {
		// for containers, include the default kubeadm profile
		profiles = append(lxcMachine.Spec.Profiles, lxcCluster.GetProfileName())
	}

	if err := c.createInstanceIfNotExists(ctx, api.InstancesPost{
		Name:         name,
		Type:         c.instanceTypeFromAPI(lxcMachine.Spec.Type),
		Source:       c.instanceSourceFromAPI(lxcMachine.Spec.Image),
		InstanceType: lxcMachine.Spec.Flavor,
		InstancePut: api.InstancePut{
			Profiles: profiles,
			Config: map[string]string{
				configClusterNameKey:      lxcCluster.Name,
				configClusterNamespaceKey: lxcCluster.Namespace,
				configInstanceRoleKey:     role,
				configCloudInitKey:        cloudInit,
			},
		},
	}); err != nil {

		// TODO: Handle the below situation as a terminalError.
		//
		// E1230 21:42:45.170291 1388422 controller.go:316] "Reconciler error" err="failed to create instance: failed to ensure instance exists: failed to wait for CreateInstance operation: Requested image's type \"container\" doesn't match instance type \"virtual-machine\"" controller="lxcmachine" controllerGroup="infrastructure.cluster.x-k8s.io" controllerKind="LXCMachine" LXCMachine="default/c1-control-plane-kprl9" namespace="default" name="c1-control-plane-kprl9" reconcileID="d40dfec7-ce45-4585-9a1e-5974efbeb925"
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
