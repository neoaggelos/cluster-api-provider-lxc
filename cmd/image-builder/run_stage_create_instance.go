package main

import (
	"context"
	"fmt"

	"github.com/lxc/incus/v6/shared/api"
)

type stageCreateInstance struct{}

func (*stageCreateInstance) name() string { return "create-instance" }

// incus launch image:ubuntu/24.04/cloud
// incus launch image:ubuntu/24.04/cloud --vm -d root,size=5GiB
func (*stageCreateInstance) run(ctx context.Context) error {
	// Fetch server information
	server, _, err := client.Client.GetServer()
	if err != nil {
		return fmt.Errorf("failed to get server information: %w", err)
	}

	// Incus and LXD have diverged image servers for Ubuntu images, making it easy to confuse users.
	// To address the issue, we allow a special prefix `ubuntu:VERSION` for image names:
	var image api.InstanceSource
	switch server.Environment.Server {
	case "incus":
		image = api.InstanceSource{
			Alias:    fmt.Sprintf("ubuntu/%s/cloud", cfg.ubuntuVersion),
			Server:   "https://images.linuxcontainers.org",
			Protocol: "simplestreams",
			Type:     "image",
		}
	case "lxd":
		image = api.InstanceSource{
			Alias:    cfg.ubuntuVersion,
			Server:   "https://cloud-images.ubuntu.com/releases/",
			Protocol: "simplestreams",
			Type:     "image",
		}
	default:
		return fmt.Errorf("unknown server name %q, must be one of [incus, lxd]", server.Environment.Server)
	}

	// LXD needs security.nesting=true for containers to be able to pull images
	var config map[string]string
	if cfg.instanceType == "container" {
		config = map[string]string{
			"security.nesting": "true",
		}
	}

	instanceConfig := api.InstancesPost{
		Name:   cfg.instanceName,
		Source: image,
		Type:   api.InstanceType(cfg.instanceType),
		InstancePut: api.InstancePut{
			Config:   config,
			Profiles: cfg.instanceProfiles,
		},
	}

	// set size of root volume to 5GB for virtual machines
	if cfg.instanceType == "virtual-machine" {
		pools, err := client.Client.GetStoragePools()
		if err != nil {
			return fmt.Errorf("failed to list storage pools: %w", err)
		}

		for _, pool := range pools {
			if pool.Status == api.StoragePoolStatusCreated {
				instanceConfig.InstancePut.Devices = map[string]map[string]string{
					"root": {
						"type": "disk",
						"pool": pool.Name,
						"path": "/",
						"size": "5GiB",
					},
				}
			}
		}
	}

	if err := client.CreateAndWaitForInstance(ctx, instanceConfig); err != nil {
		return fmt.Errorf("failed to launch builder instance: %w", err)
	}

	return nil
}
