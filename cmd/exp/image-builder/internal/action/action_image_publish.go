package action

import (
	"context"
	"fmt"
	"runtime"
	"time"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

type PublishImageInfo struct {
	Name            string
	OperatingSystem string
	Release         string
	Variant         string

	LXCRequireCgroupv2 bool
}

// PublishImage is an Action that publishes an image from an existing instance.
func PublishImage(lxcClient *lxc.Client, instanceName string, imageAliasName string, info PublishImageInfo) Action {
	return func(ctx context.Context) error {
		instance, _, err := lxcClient.GetInstance(instanceName)
		if err != nil {
			return fmt.Errorf("failed to get instance: %w", err)
		}

		// if image alias already exists:
		// - test if image is newer than the instance last used timestamp
		// - otherwise, attempt to delete alias
		if alias, _, err := lxcClient.GetImageAlias(imageAliasName); err == nil {
			if image, _, err := lxcClient.GetImage(alias.Target); err == nil && instance.LastUsedAt.Before(image.CreatedAt) {
				log.FromContext(ctx).V(1).Info("Skipping image publish, as alias exists and is newer than instance")
				return nil
			}

			log.FromContext(ctx).V(1).Info("Deleting existing image alias")
			if err := lxcClient.DeleteImageAlias(imageAliasName); err != nil {
				return fmt.Errorf("failed to delete existing image alias %q: %w", imageAliasName, err)
			}
		}

		now := time.Now()
		serial := fmt.Sprintf("%d%02d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
		properties := map[string]string{
			"architecture": runtime.GOARCH,
			"name":         info.Name,
			"description":  fmt.Sprintf("%s (%s)", info.Name, serial),
			"os":           info.OperatingSystem,
			"release":      info.Release,
			"variant":      info.Variant,
			"serial":       serial,
		}
		if info.LXCRequireCgroupv2 && instance.Type == lxc.Container {
			properties["requirements.cgroupv2"] = "true"
		}

		log.FromContext(ctx).V(1).Info("Publishing image")
		return lxcClient.WaitForOperation(ctx, "PublishImage", func() (incus.Operation, error) {
			return lxcClient.CreateImage(api.ImagesPost{
				ImagePut: api.ImagePut{
					Properties: properties,
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
}
