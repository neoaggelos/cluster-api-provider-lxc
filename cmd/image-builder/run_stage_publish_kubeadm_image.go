package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

type stagePublishKubeadmImage struct{}

func (*stagePublishKubeadmImage) name() string { return "publish-image" }

// incus publish capn-builder capn-builder-image
func (*stagePublishKubeadmImage) run(ctx context.Context) error {
	now := time.Now()
	serial := fmt.Sprintf("%d%02d%02d%02d%02d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())

	if err := client.PublishImage(ctx, cfg.instanceName, cfg.imageAlias, map[string]string{
		"architecture": runtime.GOARCH,
		"name":         fmt.Sprintf("kubeadm %s ubuntu %s %s", kubeadmCfg.kubernetesVersion, getUbuntuReleaseName(cfg.ubuntuVersion), runtime.GOARCH),
		"description":  fmt.Sprintf("kubeadm %s ubuntu %s %s (%s)", kubeadmCfg.kubernetesVersion, getUbuntuReleaseName(cfg.ubuntuVersion), runtime.GOARCH, serial),
		"os":           "kubeadm",
		"release":      kubeadmCfg.kubernetesVersion,
		"variant":      "ubuntu",
		"serial":       serial,
	}); err != nil {
		return fmt.Errorf("failed to create image from instance: %w", err)
	}
	return nil
}
