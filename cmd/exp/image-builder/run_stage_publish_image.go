package main

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

type publishImageInfo struct {
	name, os, release, variant string

	lxcRequireCgroupv2 bool
}

type stagePublishImage struct {
	info publishImageInfo
}

func (s *stagePublishImage) name() string { return fmt.Sprintf("publish-%s-image", s.info.os) }

// incus publish capn-builder capn-builder-image
func (s *stagePublishImage) run(ctx context.Context) error {
	// if image alias already exists:
	// - test if image is newer than the instance last used timestamp
	// - otherwise, attempt to delete alias
	if alias, _, err := lxcClient.GetImageAlias(cfg.imageAlias); err == nil {
		if image, _, err := lxcClient.GetImage(alias.Target); err == nil {
			if instance, _, err := lxcClient.GetInstance(cfg.instanceName); err == nil && instance.LastUsedAt.Before(image.CreatedAt) {
				log.FromContext(ctx).V(1).Info("Skipping image publish, as alias exists and is newer than instance")
				return nil
			}
		}

		log.FromContext(ctx).V(1).Info("Deleting existing image alias")
		if err := lxcClient.DeleteImageAlias(cfg.imageAlias); err != nil {
			return fmt.Errorf("failed to delete existing image alias %q: %w", cfg.imageAlias, err)
		}
	}

	now := time.Now()
	serial := fmt.Sprintf("%d%02d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
	properties := map[string]string{
		"architecture": runtime.GOARCH,
		"name":         s.info.name,
		"description":  fmt.Sprintf("%s (%s)", s.info.name, serial),
		"os":           s.info.os,
		"release":      s.info.release,
		"variant":      s.info.variant,
		"serial":       serial,
	}
	if s.info.lxcRequireCgroupv2 && cfg.instanceType == lxc.Container {
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
				Name: cfg.instanceName,
			},
			Aliases: []api.ImageAlias{
				{Name: cfg.imageAlias},
			},
		}, nil)
	})
}
