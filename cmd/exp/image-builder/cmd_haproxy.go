package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	haproxyCmd = &cobra.Command{
		Use:          "haproxy",
		GroupID:      "build",
		Short:        "Build haproxy images for cluster-api-provider-incus",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.imageAlias == "" {
				cfg.imageAlias = "haproxy"
			}

			log.FromContext(gCtx).WithValues(
				"base-image", cfg.baseImage,
				"instance-type", cfg.instanceType,
				"image-alias", cfg.imageAlias,
			).Info("Building haproxy image")

			return runStages(
				&stageCreateInstance{},
				&stagePreRunCommands{},
				&stageInstallHaproxy{},
				&stagePostRunCommands{},
				&stageCleanupInstance{},
				&stageStopInstance{},
				&stagePublishImage{
					info: publishImageInfo{
						name:    fmt.Sprintf("haproxy %s %s", wellKnownBaseImages[cfg.baseImage].fullName, runtime.GOARCH),
						os:      "haproxy",
						release: wellKnownBaseImages[cfg.baseImage].releaseName,
						variant: wellKnownBaseImages[cfg.baseImage].variantName,
					},
				},
				&stageExportImage{},
				&stageRemoveInstance{},
			)
		},
	}
)
