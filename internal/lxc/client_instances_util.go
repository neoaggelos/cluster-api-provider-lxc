package lxc

import (
	"context"
	"fmt"
	"strings"
	"time"

	incus "github.com/lxc/incus/v6/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// List existing server operations to find if any CreateInstance operations are pending for the target instance name.
// If anything fails, it will return the error that occurred.
// If any is active, it will return a waitable incus.Operation object.
// If no operations are active, it will return nil.
//
// NOTE(neoaggelos/2025-02-15): Reference instance create operation metadata (for incus):
// I0213 13:22:34.766849 3032651 client_test.go:39] "Starting" operation={"id":"b0fb6039-b45f-4ee0-89f5-d7b1b1dd164f","class":"task","description":"Creating instance","created_at":"2025-02-13T23:22:31.608899984+02:00","updated_at":"2025-02-13T23:22:34.764554068+02:00","status":"Running","status_code":103,"resources":{"instances":["/1.0/instances/t2"]},"metadata":{"create_instance_from_image_unpack_progress":"Unpacking image: 88%","progress":{"percent":"88","speed":"0","stage":"create_instance_from_image_unpack"}},"may_cancel":false,"err":"","location":"damocles"}
func (c *Client) tryFindInstanceCreateOperation(ctx context.Context, instanceName string) (incus.Operation, error) {
	ops, err := c.GetOperations()
	if err != nil {
		return nil, fmt.Errorf("failed to GetOperations: %w", err)
	}

	instancePath := fmt.Sprintf("/1.0/instances/%s", instanceName)

	for _, op := range ops {
		if op.Class != "task" && op.Description != "Creating instance" {
			continue
		}
		if instances := op.Resources["instances"]; len(instances) != 1 || instances[0] != instancePath {
			continue
		}

		apiOperation, _, err := c.RawOperation("GET", fmt.Sprintf("/operations/%s", op.ID), nil, "")
		if err != nil {
			return nil, fmt.Errorf("failed to GetOperation(%s): %w", op.ID, err)
		}
		log.FromContext(ctx).V(4).WithValues("operation.uuid", op.ID, "operation.metadata", op.Metadata, "operation.status", op.Status).Info("Found existing instance create operation")

		return apiOperation, nil
	}
	return nil, nil
}

func (c *Client) waitForInstanceAddress(ctx context.Context, name string) ([]string, error) {
	for {
		log.FromContext(ctx).V(4).Info("Waiting for instance address")
		if state, _, err := c.GetInstanceState(name); err != nil {
			return nil, fmt.Errorf("failed to GetInstanceState: %w", err)
		} else if addrs := ParseHostAddresses(state); len(addrs) > 0 {
			return addrs, nil
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for instance address: %w", ctx.Err())
		case <-time.After(time.Second):
		}
	}
}

// LaunchOptions describe additional provisioning actions for machines.
type LaunchOptions struct {
	// SeedFiles are "<file>"="<contents>" template files that will be created on the machine.
	// Supported by all instance types.
	SeedFiles map[string]string
	// Symlinks are "<path>"="<target>" symbolic links to that will be created on the machine.
	// Not supported by virtual-machine instance types.
	Symlinks map[string]string
	// Replacements are a list of string replacements to perform on files on the machine.
	// The replacer is expected to be idempotent.
	// Not supported by virtual-machine instance types.
	Replacements map[string]*strings.Replacer
}

func (o *LaunchOptions) GetSeedFiles() map[string]string {
	if o == nil {
		return nil
	}
	return o.SeedFiles
}

func (o *LaunchOptions) GetSymlinks() map[string]string {
	if o == nil {
		return nil
	}
	return o.Symlinks
}

func (o *LaunchOptions) GetReplacements() map[string]*strings.Replacer {
	if o == nil {
		return nil
	}
	return o.Replacements
}
