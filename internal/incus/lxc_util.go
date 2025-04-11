package incus

import (
	"context"
	"fmt"
	"strings"
	"time"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/loadbalancer"
)

// wait executes an Incus API call that returns an Operation, and waits for the operation to complete.
// Returns an error if anything failed.
func (c *Client) wait(ctx context.Context, name string, f func() (incus.Operation, error)) error {
	op, err := f()
	if err != nil {
		return fmt.Errorf("failed to %s: %w", name, err)
	}

	// log progress of LXC operation. Note that this will be very verbose and but will be very useful to troubleshoot potential issues
	operationLogger := log.FromContext(ctx).V(2).WithValues("operation.name", name)
	target, _ := op.AddHandler(func(o api.Operation) {
		log := operationLogger.WithValues("operation.uuid", o.ID, "operation.metadata", o.Metadata, "operation.status", o.Status)
		if o.Err != "" {
			log = log.WithValues("operation.err", o.Err)
		}
		switch {
		case o.StatusCode == api.Failure:
			log.Error(err, "Operation failed")
		case o.StatusCode.IsFinal():
			log.Info("Operation finished")
		default:
			log.Info("Operation in progress")
		}
	})
	defer func() {
		_ = op.RemoveHandler(target)
	}()

	if err := op.WaitContext(ctx); err != nil && !strings.Contains(err.Error(), "Operation not found") {
		return fmt.Errorf("failed to wait for %s operation: %w", name, err)
	}
	return nil
}

func (c *Client) waitForInstanceAddress(ctx context.Context, name string) ([]string, error) {
	for {
		log.FromContext(ctx).V(2).Info("Waiting for instance address")
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
			log.FromContext(ctx).V(2).Info("Instance does not exist")
			return nil
		}
		return fmt.Errorf("failed to GetInstanceState: %w", err)
	}

	// stop instance if running
	if state.Pid != 0 {
		log.FromContext(ctx).WithValues("status", state.Status, "pid", state.Pid).V(2).Info("Stopping instance")
		if err := c.wait(ctx, "UpdateInstanceState", func() (incus.Operation, error) {
			return c.Client.UpdateInstanceState(name, api.InstanceStatePut{Action: "stop", Force: true}, "")
		}); err != nil {
			return err
		}
	}

	// delete stopped instance
	log.FromContext(ctx).V(2).Info("Deleting instance")
	if err := c.wait(ctx, "DeleteInstance", func() (incus.Operation, error) {
		return c.Client.DeleteInstance(name)
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) createInstanceIfNotExists(ctx context.Context, instance api.InstancesPost) error {
	state, _, err := c.Client.GetInstanceState(instance.Name)
	if err != nil && !strings.Contains(err.Error(), "Instance not found") {
		return fmt.Errorf("failed to GetInstanceState: %w", err)
	} else if err == nil {
		log.FromContext(ctx).V(2).WithValues("status", state.Status).Info("Instance exists")
		return nil
	}

	log.FromContext(ctx).V(2).Info("Creating instance")
	return c.wait(ctx, "CreateInstance", func() (incus.Operation, error) {
		op, err := c.tryFindInstanceCreateOperation(ctx, instance.Name)
		if err != nil {
			log.FromContext(ctx).Error(err, "Warning: failed to check for existing instance create operation")
		} else if op != nil {
			log.FromContext(ctx).V(2).Info("Found existing create operation")
			return op, nil
		}
		return c.Client.CreateInstance(instance)
	})
}

func (c *Client) ensureInstanceRunning(ctx context.Context, name string) error {
	state, _, err := c.Client.GetInstanceState(name)
	if err != nil {
		return fmt.Errorf("failed to GetInstanceState: %w", err)
	}

	action := "start"
	if state.Status == "Running" {
		log.FromContext(ctx).V(2).WithValues("status", state.Status).Info("Instance is already running")
		return nil
	} else if state.Status == "Frozen" {
		action = "unfreeze"
	}

	log.FromContext(ctx).V(2).WithValues("status", state.Status, "action", action).Info("Starting instance")
	return c.wait(ctx, "UpdateInstanceState", func() (incus.Operation, error) {
		return c.Client.UpdateInstanceState(name, api.InstanceStatePut{Action: action}, "")
	})
}

func (c *Client) getInstancesWithFilter(ctx context.Context, instanceType api.InstanceType, filters map[string]string) ([]api.InstanceFull, error) {
	// TODO(neoaggelos): use server-side filters
	// instances, err := c.GetInstancesFullWithFilter(api.InstanceTypeAny, []string{"user.capi.cluster eq c1 and user.capi.role eq control-plane"})
	log.FromContext(ctx).V(2).WithValues("filters", filters).Info("Retrieving LXC instances with specified filter")
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
				log.V(2).WithValues("key", k, "want", v, "have", instance.Config[k]).Info("Ignoring instance")
				continue nextInstance
			}
		}
		log.V(2).Info("Found instance")
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
		Fingerprint: source.Fingerprint,
	}

	return result
}

func (c *Client) instanceTypeFromAPI(instanceType string) api.InstanceType {
	if instanceType == "" {
		return api.InstanceTypeContainer
	}
	return api.InstanceType(instanceType)
}

func (c *Client) getLoadBalancerConfiguration(ctx context.Context, clusterName string, clusterNamespace string) (*loadbalancer.ConfigData, error) {
	instances, err := c.getInstancesWithFilter(ctx, api.InstanceTypeAny, map[string]string{
		configClusterNameKey:      clusterName,
		configClusterNamespaceKey: clusterNamespace,
		configInstanceRoleKey:     "control-plane",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cluster control plane instances: %w", err)
	}

	config := &loadbalancer.ConfigData{
		FrontendControlPlanePort: "6443",
		BackendControlPlanePort:  "6443",
		BackendServers:           make(map[string]loadbalancer.BackendServer, len(instances)),
	}
	for _, instance := range instances {
		if addresses := c.ParseActiveMachineAddresses(instance.State); len(addresses) > 0 {
			// TODO(neoaggelos): care about the instance weight (e.g. for deleted machines)
			// TODO(neoaggelos): care about ipv4 vs ipv6 addresses
			config.BackendServers[instance.Name] = loadbalancer.BackendServer{Address: addresses[0], Weight: 100}
		}
	}

	return config, nil
}

// The built-in Client.HasExtension() from Incus cannot be trusted, as it returns true if we skip the GetServer call.
// Return the list of extensions that are NOT supported by the server, if any.
func (c *Client) serverSupportsExtensions(extensions ...string) ([]string, error) {
	if server, _, err := c.Client.GetServer(); err != nil {
		return nil, fmt.Errorf("failed to retrieve server information: %w", err)
	} else {
		return sets.New(extensions...).Difference(sets.New(server.APIExtensions...)).UnsortedList(), nil
	}
}

// List existing server operations to find if any CreateInstance operations are pending for the target instance name.
// If anything fails, it will return the error that occurred.
// If any is active, it will return a waitable incus.Operation object.
// If no operations are active, it will return nil.
//
// NOTE(neoaggelos/2025-02-15): Reference instance create operation metadata (for incus):
// I0213 13:22:34.766849 3032651 client_test.go:39] "Starting" operation={"id":"b0fb6039-b45f-4ee0-89f5-d7b1b1dd164f","class":"task","description":"Creating instance","created_at":"2025-02-13T23:22:31.608899984+02:00","updated_at":"2025-02-13T23:22:34.764554068+02:00","status":"Running","status_code":103,"resources":{"instances":["/1.0/instances/t2"]},"metadata":{"create_instance_from_image_unpack_progress":"Unpacking image: 88%","progress":{"percent":"88","speed":"0","stage":"create_instance_from_image_unpack"}},"may_cancel":false,"err":"","location":"damocles"}
func (c *Client) tryFindInstanceCreateOperation(ctx context.Context, instanceName string) (incus.Operation, error) {
	ops, err := c.Client.GetOperations()
	if err != nil {
		return nil, fmt.Errorf("failed to GetOperations: %w", err)
	}

	instancePath := fmt.Sprintf("/1.0/instances/%s", instanceName)

	for _, metadata := range ops {
		if metadata.Class != "task" && metadata.Description != "Creating instance" {
			continue
		}
		if instances := metadata.Resources["instances"]; len(instances) != 1 || instances[0] != instancePath {
			continue
		}

		op, _, err := c.Client.RawOperation("GET", fmt.Sprintf("/operations/%s", metadata.ID), nil, "")
		if err != nil {
			return nil, fmt.Errorf("failed to GetOperation(%s): %w", metadata.ID, err)
		}
		log.FromContext(ctx).V(4).WithValues("operation", metadata).Info("Found existing instance create operation")

		return op, nil
	}
	return nil, nil
}
