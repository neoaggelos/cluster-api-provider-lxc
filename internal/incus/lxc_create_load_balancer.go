package incus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

func (c *Client) createLoadBalancer_createInstance(ctx context.Context, cluster *infrav1.LXCCluster, name string) error {
	state, _, err := c.Client.GetInstanceState(name)
	if err != nil {
		if !strings.Contains(err.Error(), "Instance not found") {
			return fmt.Errorf("failed to retrieve current state: %w", err)
		}
	} else if state.Status == "Error" || state.StatusCode.IsFinal() {
		log.FromContext(ctx).V(4).Info("Deleting failed instance", "state", state)

		if op, err := c.Client.DeleteInstance(name); err != nil {
			return fmt.Errorf("failed to DeleteInstance: %w", err)
		} else if err := op.WaitContext(ctx); err != nil {
			return fmt.Errorf("failed to wait for DeleteInstance operation: %w", err)
		}
	} else {
		log.FromContext(ctx).V(4).WithValues("status", state.Status).Info("Instance exists")
		return nil
	}

	log.FromContext(ctx).V(4).Info("Creating instance")

	if op, err := c.Client.CreateInstance(api.InstancesPost{
		Name: name,
		Type: api.InstanceTypeContainer,
		Source: api.InstanceSource{
			Type:     "image",
			Protocol: "oci",
			Server:   "https://docker.io",
			Mode:     "pull",
			Alias:    "kindest/haproxy:v20230606-42a2262b", // TODO(neoaggelos): mirror and use our own image
		},
		InstancePut: api.InstancePut{
			Config: map[string]string{
				"user.capi.name":      cluster.Name,
				"user.capi.namespace": cluster.Namespace,
				"user.capi.role":      "loadbalancer",
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to CreateInstance: %w", err)
	} else if err := op.WaitContext(ctx); err != nil {
		return fmt.Errorf("failed to wait for CreateInstance operation: %w", err)
	}

	return nil
}

func (c *Client) createLoadBalancer_startInstance(ctx context.Context, name string) error {
	state, _, err := c.Client.GetInstanceState(name)
	if err != nil {
		return fmt.Errorf("failed to retrieve current container state: %w", err)
	}

	action := "start"
	if state.Status == "Running" {
		log.FromContext(ctx).V(4).WithValues("status", state.Status).Info("Instance is running")
		return nil
	} else if state.Status == "Frozen" {
		action = "unfreeze"
	}

	log.FromContext(ctx).V(4).WithValues("status", state.Status, "action", action).Info("Starting instance")
	if op, err := c.Client.UpdateInstanceState(name, api.InstanceStatePut{Action: action}, ""); err != nil {
		return fmt.Errorf("failed to UpdateInstanceState: %w", err)
	} else if err := op.WaitContext(ctx); err != nil && !strings.Contains(err.Error(), "Operation not found") {
		return fmt.Errorf("failed to wait for UpdateInstanceState operation: %w", err)
	}

	return nil
}

// CreateLoadBalancer creates the LoadBalancer container and returns the instance address.
func (c *Client) CreateLoadBalancer(ctx context.Context, cluster *infrav1.LXCCluster) (string, error) {
	name := fmt.Sprintf("%s-%s-lb", cluster.Namespace, cluster.Name)

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("container", name))

	if err := c.createLoadBalancer_createInstance(ctx, cluster, name); err != nil {
		return "", fmt.Errorf("failed to create loadbalancer instance: %w", err)
	}
	if err := c.createLoadBalancer_startInstance(ctx, name); err != nil {
		return "", fmt.Errorf("failed to start loadbalancer instance: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for {
		log.FromContext(ctx).V(4).Info("Waiting for loadbalancer instance address")
		if state, _, err := c.Client.GetInstanceState(name); err != nil {
			return "", fmt.Errorf("failed to GetInstanceState: %w", err)
		} else {
		nextNetwork:
			for _, network := range state.Network {
				if network.Type == "loopback" {
					continue nextNetwork
				}

				for _, addr := range network.Addresses {
					// TODO(neoaggelos): care for addr.Family ipv4 vs ipv6
					if addr.Scope != "global" {
						continue nextNetwork
					}

					return addr.Address, nil
				}
			}
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("timed out waiting for loadbalancer address: %w", ctx.Err())
		case <-time.After(time.Second):
		}
	}
}
