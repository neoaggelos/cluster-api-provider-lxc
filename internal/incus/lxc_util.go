package incus

import (
	"context"
	"fmt"
	"strings"
	"time"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

// wait executes an Incus API call that returns an Operation, and waits for the operation to complete.
// Returns an error if anything failed.
func (c *Client) wait(ctx context.Context, name string, f func() (incus.Operation, error)) error {
	if op, err := f(); err != nil {
		return fmt.Errorf("failed to %s: %q", name, err)
	} else if err := op.WaitContext(ctx); err != nil && !strings.Contains(err.Error(), "Operation not found") {
		return fmt.Errorf("failed to wait for %s operation: %w", name, err)
	}
	return nil
}

func (c *Client) waitForInstanceAddress(ctx context.Context, name string) ([]string, error) {
	for {
		log.FromContext(ctx).V(4).Info("Checking for instance address")
		if state, _, err := c.Client.GetInstanceState(name); err != nil {
			return nil, fmt.Errorf("failed to GetInstanceState: %w", err)
		} else if addrs := c.ParseActiveMachineAddresses(state); len(addrs) > 0 {
			return addrs, nil
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for instance address: %w", ctx.Err())
		case <-time.After(time.Second):
		}
	}
}

func (c *Client) forceRemoveInstanceIfExists(ctx context.Context, name string) error {
	state, _, err := c.Client.GetInstanceState(name)
	if err != nil {
		if strings.Contains(err.Error(), "Instance not found") {
			log.FromContext(ctx).V(4).Info("Instance does not exist")
			return nil
		}
		return fmt.Errorf("failed to GetInstanceState: %w", err)
	}

	// stop instance if running
	if state.Pid != 0 {
		log.FromContext(ctx).WithValues("status", state.Status, "pid", state.Pid).V(4).Info("Stopping instance")
		if err := c.wait(ctx, "UpdateInstanceState", func() (incus.Operation, error) {
			return c.Client.UpdateInstanceState(name, api.InstanceStatePut{Action: "stop", Force: true}, "")
		}); err != nil {
			return err
		}
	}

	// delete stopped instance
	log.FromContext(ctx).V(4).Info("Deleting instance")
	if err := c.wait(ctx, "DeleteInstance", func() (incus.Operation, error) {
		return c.Client.DeleteInstance(name)
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) createInstanceIfNotExists(ctx context.Context, instance api.InstancesPost) error {
	state, _, err := c.Client.GetInstanceState(instance.Name)
	if err != nil {
		if !strings.Contains(err.Error(), "Instance not found") {
			return fmt.Errorf("failed to GetInstanceState: %w", err)
		}
	} else if state.Status == "Error" || state.StatusCode.IsFinal() {
		log.FromContext(ctx).V(4).Info("Deleting old failed instance", "state", state)

		if err := c.wait(ctx, "DeleteInstance", func() (incus.Operation, error) { return c.Client.DeleteInstance(instance.Name) }); err != nil {
			return err
		}
	} else {
		log.FromContext(ctx).V(4).WithValues("status", state.Status).Info("Instance exists")
		return nil
	}

	log.FromContext(ctx).V(4).Info("Creating instance")
	return c.wait(ctx, "CreateInstance", func() (incus.Operation, error) { return c.Client.CreateInstance(instance) })
}

func (c *Client) ensureInstanceRunning(ctx context.Context, name string) error {
	state, _, err := c.Client.GetInstanceState(name)
	if err != nil {
		return fmt.Errorf("failed to GetInstanceState: %w", err)
	}

	action := "start"
	if state.Status == "Running" {
		log.FromContext(ctx).V(4).WithValues("status", state.Status).Info("Instance is already running")
		return nil
	} else if state.Status == "Frozen" {
		action = "unfreeze"
	}

	log.FromContext(ctx).V(4).WithValues("status", state.Status, "action", action).Info("Starting instance")
	return c.wait(ctx, "UpdateInstanceState", func() (incus.Operation, error) {
		return c.Client.UpdateInstanceState(name, api.InstanceStatePut{Action: action}, "")
	})
}

func (c *Client) getInstancesWithFilter(ctx context.Context, instanceType api.InstanceType, filters map[string]string) ([]api.InstanceFull, error) {
	// TODO(neoaggelos): use server-side filters
	// instances, err := c.GetInstancesFullWithFilter(api.InstanceTypeAny, []string{"user.capi.cluster eq c1 and user.capi.role eq control-plane"})
	log.FromContext(ctx).V(4).WithValues("filters", filters).Info("Retrieving LXC instances with specified filter")
	unfiltereredInstances, err := c.Client.GetInstancesFull(instanceType)
	if err != nil {
		return nil, fmt.Errorf("failed to GetInstancesFull: %w", err)
	}

	var instances []api.InstanceFull
nextInstance:
	for _, instance := range unfiltereredInstances {
		log := log.FromContext(ctx).WithValues("instance", instance.Name)
		for k, v := range filters {
			if instance.Config[k] != v {
				log.V(4).WithValues("key", k, "want", v, "have", instance.Config[k]).Info("Ignoring instance")
				continue nextInstance
			}
		}
		log.V(4).Info("Found instance")
		instances = append(instances, instance)
	}

	return instances, nil
}

func (c *Client) instanceSourceFromAPI(source infrav1.LXCMachineImageSource) api.InstanceSource {
	result := api.InstanceSource{
		Type:        "image",
		Alias:       source.Name,
		Server:      source.Server,
		Protocol:    source.Protocol,
		Source:      source.Snapshot,
		Fingerprint: source.Fingerprint,
	}

	return result
}
