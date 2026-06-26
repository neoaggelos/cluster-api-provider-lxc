package lxc

import (
	"context"
	"fmt"

	"github.com/lxc/incus/v7/shared/api"
)

type ListInstanceFilter func(i api.InstanceFull) bool

// WithConfig filters for instances that have a set of configuration keys.
func WithConfig(kv map[string]string) ListInstanceFilter {
	return func(i api.InstanceFull) bool {
		for k, v := range kv {
			if i.Config[k] != v {
				return false
			}
		}

		return true
	}
}

// WithConfigKeys filters for instances that have configuration keys set (regardless of values).
func WithConfigKeys(keys ...string) ListInstanceFilter {
	return func(i api.InstanceFull) bool {
		for _, k := range keys {
			if _, ok := i.Config[k]; !ok {
				return false
			}
		}

		return true
	}
}

// Containers filters for instances of type container.
func Containers() ListInstanceFilter {
	return func(i api.InstanceFull) bool {
		return i.Type == Container
	}
}

// VirtualMachines filters for instances of type virtual-machine.
func VirtualMachines() ListInstanceFilter {
	return func(i api.InstanceFull) bool {
		return i.Type == VirtualMachine
	}
}

// ListInstances returns a list of instances that match the specified filters.
// The filters are currently applied on the client side, due to observed issues with server side filtering.
func (c *Client) ListInstances(ctx context.Context, filters ...ListInstanceFilter) ([]api.InstanceFull, error) {
	allInstances, err := c.GetInstancesFull(api.InstanceTypeAny)
	if err != nil {
		return nil, fmt.Errorf("failed to GetInstancesFull: %w", err)
	}

	var instances []api.InstanceFull
nextInstance:
	for _, instance := range allInstances {
		for _, filter := range filters {
			if !filter(instance) {
				continue nextInstance
			}
		}

		instances = append(instances, instance)
	}

	return instances, nil
}
