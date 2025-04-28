package main

import (
	"fmt"

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
				cfg.imageAlias = fmt.Sprintf("haproxy-u%s", cfg.ubuntuVersion)
			}

			log.FromContext(gCtx).WithValues(
				"ubuntu-version", cfg.ubuntuVersion,
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
				&stagePublishHaproxyImage{},
				&stageExportImage{},
				&stageRemoveInstance{},
			)
		},
	}
)
