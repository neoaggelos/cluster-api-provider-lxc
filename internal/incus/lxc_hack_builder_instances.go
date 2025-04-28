package incus

import (
	"context"
	"fmt"
	"strings"
	"time"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *Client) CreateAndWaitForInstance(ctx context.Context, instance api.InstancesPost) error {
	if err := c.createInstanceIfNotExists(ctx, instance); err != nil {
		return fmt.Errorf("failed to ensure instance exists: %w", err)
	}
	if err := c.ensureInstanceRunning(ctx, instance.Name); err != nil {
		return fmt.Errorf("failed to ensure instance is running: %w", err)
	}
	if _, err := c.waitForInstanceAddress(ctx, instance.Name); err != nil {
		return fmt.Errorf("failed to wait for instance address: %w", err)
	}
	return nil
}

func (c *Client) StopInstance(ctx context.Context, name string) error {
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

	return nil
}

func (c *Client) CreateInstanceSnapshot(ctx context.Context, instanceName string, snapshotName string) error {
	if _, _, err := c.Client.GetInstanceSnapshot(instanceName, snapshotName); err == nil {
		log.FromContext(ctx).V(2).Info("Instance snapshot already exists")
		return nil
	}
	return c.wait(ctx, "CreateInstanceSnapshot", func() (incus.Operation, error) {
		return c.Client.CreateInstanceSnapshot(instanceName, api.InstanceSnapshotsPost{
			Name: snapshotName,
		})
	})
}

func (c *Client) PublishImage(ctx context.Context, instanceName string, imageAliasName string, imageProperties map[string]string) error {
	if _, _, err := c.Client.GetImageAlias(imageAliasName); err == nil {
		log.FromContext(ctx).V(2).Info("Image alias already exists")
		return nil
	}

	return c.wait(ctx, "CreateImage", func() (incus.Operation, error) {
		return c.Client.CreateImage(api.ImagesPost{
			ImagePut: api.ImagePut{
				Properties: imageProperties,
				Public:     true,
				ExpiresAt:  time.Now().AddDate(10, 0, 0),
			},
			Source: &api.ImagesPostSource{
				Type: "instance",
				Name: instanceName,
			},
			Aliases: []api.ImageAlias{
				{Name: imageAliasName},
			},
		}, nil)
	})
}

func (c *Client) ForceRemoveInstance(ctx context.Context, instanceName string) error {
	return c.forceRemoveInstanceIfExists(ctx, instanceName)
}
