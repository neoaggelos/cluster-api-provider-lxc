package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

type stagePublishHaproxyImage struct{}

func (*stagePublishHaproxyImage) name() string { return "publish-image" }

// incus publish capn-builder capn-builder-image
func (s *stagePublishHaproxyImage) run(ctx context.Context) error {
	now := time.Now()
	serial := fmt.Sprintf("%d%02d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())

	if err := client.PublishImage(ctx, cfg.instanceName, cfg.imageAlias, map[string]string{
		"architecture": runtime.GOARCH,
		"name":         fmt.Sprintf("haproxy %s %s", getUbuntuReleaseName(cfg.ubuntuVersion), runtime.GOARCH),
		"description":  fmt.Sprintf("haproxy %s %s (%s)", getUbuntuReleaseName(cfg.ubuntuVersion), runtime.GOARCH, serial),
		"os":           "haproxy",
		"release":      getUbuntuReleaseName(cfg.ubuntuVersion),
		"variant":      "ubuntu",
		"serial":       serial,
	}); err != nil {
		return fmt.Errorf("failed to create image from instance: %w", err)
	}
	return nil
}
