package lxc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// WaitForLaunchInstance attempts to launch and start the specified instance.
// If an instance with the same name already exists, WaitForLaunchInstance will start the instance.
// If an instance create operation is already underway, it will wait for the existing operation and start the instance.
//
// WaitForLaunchInstance will wait for the instance to have a valid host address, and returns a slice of host addresses on success.
func (c *Client) WaitForLaunchInstance(ctx context.Context, name string, opts *LaunchOptions) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, instanceCreateTimeout)
	defer cancel()

	if _, _, err := c.GetInstanceState(name); err == nil {
		log.FromContext(ctx).V(2).Info("Instance already exists")
		return c.WaitForStartInstance(ctx, name)
	} else if err := c.WaitForOperation(ctx, "CreateInstance", func() (incus.Operation, error) {
		if op, err := c.tryFindInstanceCreateOperation(ctx, name); err == nil && op != nil {
			return op, nil
		}

		if err := opts.complete(c.GetServerName()); err != nil {
			return nil, fmt.Errorf("failed to complete launch options: %w", err)
		}

		log.FromContext(ctx).V(2).WithValues(
			"lxc.instance.name", name,
			"lxc.instance.image", opts.image,
			"lxc.instance.type", opts.instanceType,
			"lxc.instance.flavor", opts.flavor,
			"lxc.instance.profiles", opts.profiles,
			"lxc.instance.devices", slices.Collect(maps.Keys(opts.devices)),
		).Info("Creating instance")

		return c.CreateInstance(api.InstancesPost{
			Name:         name,
			Source:       opts.image.(Image).InstanceSource(), // NOTE(neoaggelos): after complete(), this is an Image
			Type:         opts.instanceType,
			InstanceType: opts.flavor,
			InstancePut: api.InstancePut{
				Config:   opts.config,
				Devices:  opts.devices,
				Profiles: opts.profiles,
			},
		})
	}); err != nil {
		return nil, err
	}

	if templates := opts.instanceTemplates; len(templates) > 0 {
		metadata, _, err := c.GetInstanceMetadata(name)
		if err != nil {
			return nil, fmt.Errorf("failed to GetInstanceMetadata: %w", err)
		}

		if metadata.Templates == nil {
			metadata.Templates = make(map[string]*api.ImageMetadataTemplate, len(templates))
		}

		for path, contents := range templates {
			templateName := fmt.Sprintf("%s.tpl", filepath.Base(path))
			if err := c.CreateInstanceTemplateFile(name, templateName, strings.NewReader(contents)); err != nil {
				return nil, fmt.Errorf("failed to CreateInstanceTemplateFile(%s): %w", templateName, err)
			}

			metadata.Templates[path] = &api.ImageMetadataTemplate{
				When:       []string{"create"},
				CreateOnly: true,
				Template:   templateName,
			}
		}
		if err := c.UpdateInstanceMetadata(name, *metadata, ""); err != nil {
			return nil, fmt.Errorf("failed to UpdateInstanceMetadata: %w", err)
		}
	}

	for _, file := range opts.createFiles {
		if err := c.CreateInstanceFile(name, file.path(), file.args()); err != nil {
			return nil, fmt.Errorf("failed to %s: %w", file.action(), err)
		}
	}

	for path, replacer := range opts.replacements {
		reader, resp, err := c.GetInstanceFile(name, path)
		if err != nil {
			return nil, fmt.Errorf("failed to replace text in %q: failed to GetInstanceFile: %w", path, err)
		}

		b, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to replace text in %q: failed to read file: %w", path, err)
		}
		contents := string(b)
		if err := reader.Close(); err != nil {
			return nil, fmt.Errorf("failed to replace text in %q: failed to close reader: %w", path, err)
		}

		// NOTE(neoaggelos): this is slow, but acceptably simple for our use case.
		newContents := contents
		for old, new := range replacer {
			newContents = strings.ReplaceAll(newContents, old, new)
		}

		if newContents == contents {
			continue
		} else if err := c.CreateInstanceFile(name, path, incus.InstanceFileArgs{
			Content:   bytes.NewReader([]byte(newContents)),
			Mode:      resp.Mode,
			UID:       resp.UID,
			GID:       resp.GID,
			WriteMode: "overwrite",
			Type:      resp.Type,
		}); err != nil {
			return nil, fmt.Errorf("failed to replace text in %q: failed to CreateInstanceFile: %w", path, err)
		}
	}

	return c.WaitForStartInstance(ctx, name)
}

// WaitForStartInstance starts an instance, and waits for at least one valid host address.
func (c *Client) WaitForStartInstance(ctx context.Context, name string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, instanceStartTimeout)
	defer cancel()

	state, _, err := c.GetInstanceState(name)
	if err != nil {
		return nil, fmt.Errorf("failed to GetInstanceState: %w", err)
	}
	log := log.FromContext(ctx).WithValues("instance.status", state.Status)

	if state.Status == "Running" {
		log.V(2).Info("Instance is already running")
	} else if err := c.WaitForOperation(ctx, "StartInstance", func() (incus.Operation, error) {
		log.V(2).Info("Starting instance")
		return c.UpdateInstanceState(name, api.InstanceStatePut{Action: "start"}, "")
	}); err != nil {
		return nil, fmt.Errorf("failed to start instance: %w", err)
	}

	return c.waitForInstanceAddress(ctx, name)
}

// WaitForStopInstance stops an instance and waits for the operation to succeed.
// WaitForStopInstance fails if the instance does not exist.
func (c *Client) WaitForStopInstance(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, instanceStopTimeout)
	defer cancel()

	state, _, err := c.GetInstanceState(name)
	if err != nil {
		return fmt.Errorf("failed to GetInstanceState: %w", err)
	}
	log := log.FromContext(ctx).WithValues("instance.status", state.Status, "instance.pid", state.Pid)

	if state.Pid == 0 {
		log.V(2).Info("Instance is not running")
		return nil
	}
	log.V(2).Info("Stopping instance")
	return c.WaitForOperation(ctx, "StopInstance", func() (incus.Operation, error) {
		return c.UpdateInstanceState(name, api.InstanceStatePut{Action: "stop", Force: true}, "")
	})
}

// WaitForDeleteInstance stops and removes an instance.
// WaitForDeleteInstance will not fail if the instance does not exist.
func (c *Client) WaitForDeleteInstance(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, instanceDeleteTimeout)
	defer cancel()

	if err := c.WaitForStopInstance(ctx, name); err != nil && strings.Contains(err.Error(), "Instance not found") {
		log.FromContext(ctx).V(2).Info("Instance does not exist")
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	// delete stopped instance
	log.FromContext(ctx).V(2).Info("Deleting instance")
	return c.WaitForOperation(ctx, "DeleteInstance", func() (incus.Operation, error) {
		return c.DeleteInstance(name)
	})
}
