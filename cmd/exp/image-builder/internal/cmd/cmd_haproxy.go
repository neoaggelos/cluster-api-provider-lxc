package cmd

import (
	"fmt"
	"runtime"

	"github.com/lxc/incus/v6/shared/api"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/cmd/exp/image-builder/internal/action"
	"github.com/lxc/cluster-api-provider-incus/cmd/exp/image-builder/internal/stage"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/static"
)

func newHaproxyCmd() *cobra.Command {
	var flags struct {
		// client configuration
		configFile       string
		configRemoteName string

		// base image configuration
		baseImage string

		// builder configuration
		instanceName     string
		instanceProfiles []string

		// image alias configuration
		imageAlias string

		// build step configuration
		skipStages []string
		onlyStages []string
		dryRun     bool

		// output
		outputFile string
	}

	cmd := &cobra.Command{
		Use:     "haproxy",
		GroupID: "build",
		Short:   "Build haproxy images for cluster-api-provider-incus",

		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch flags.baseImage {
			case "debian":
				flags.baseImage = "debian:13"
			case "ubuntu":
				flags.baseImage = "ubuntu:24.04"
			case "debian:12", "debian:13", "ubuntu:22.04", "ubuntu:24.04":
			default:
				return fmt.Errorf("invalid value for --base-image argument %q, must be one of [ubuntu:22.04, ubuntu:24.04, debian:12, debian:13]", flags.baseImage)
			}

			if flags.imageAlias == "" {
				flags.imageAlias = fmt.Sprintf("haproxy-%s", wellKnownBaseImages[flags.baseImage].variantName)
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := lxc.ConfigurationFromLocal(flags.configFile, flags.configRemoteName, false)
			if err != nil {
				return fmt.Errorf("failed to read client credentials: %w", err)
			}

			lxcClient, err := lxc.New(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("failed to create incus client: %w", err)
			}

			image, _, err := lxc.TryParseImageSource(lxcClient.GetServerName(), flags.baseImage)
			if err != nil {
				return fmt.Errorf("failed to parse base image: %w", err)
			}

			stages := []stage.Stage{
				{Name: "create-instance", Action: action.LaunchInstance(lxcClient, flags.instanceName, (&lxc.LaunchOptions{}).
					WithInstanceType(api.InstanceTypeContainer).
					WithProfiles(flags.instanceProfiles).
					MaybeWithImage(image),
				)},
				// {Name: "pre-run-commands", Action: action.ExecInstance(lxcClient, flags.instanceName, <TODO>, <TODO>)},
				{Name: "install-haproxy", Action: action.ExecInstance(lxcClient, flags.instanceName, static.InstallHaproxyScript())},
				// {Name: "post-run-commands", Action: action.ExecInstance(lxcClient, flags.instanceName, <TODO>, <TODO>)},
				{Name: "prepare-instance", Action: action.ExecInstance(lxcClient, flags.instanceName, static.CleanupInstanceScript())},
				{Name: "stop-instance", Action: action.StopInstance(lxcClient, flags.instanceName)},
				{Name: "publish-image", Action: action.PublishImage(lxcClient, flags.instanceName, flags.imageAlias, action.PublishImageInfo{
					Name:            fmt.Sprintf("haproxy %s %s", wellKnownBaseImages[flags.baseImage].fullName, runtime.GOARCH),
					OperatingSystem: "haproxy",
					Release:         wellKnownBaseImages[flags.baseImage].releaseName,
					Variant:         wellKnownBaseImages[flags.baseImage].variantName,
				})},
				{Name: "export-image", Action: action.ExportImage(lxcClient, flags.imageAlias, flags.outputFile)},
				{Name: "delete-instance", Action: action.DeleteInstance(lxcClient, flags.instanceName)},
			}

			log.FromContext(cmd.Context()).WithValues(
				"base-image", flags.baseImage,
				"instance-type", "container",
				"image-alias", flags.imageAlias,
			).Info("Building haproxy image")

			if err := stage.Run(cmd.Context(), flags.skipStages, flags.onlyStages, flags.dryRun, stages...); err != nil {
				return fmt.Errorf("failed to run kubeadm stages: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.configFile, "config-file", "",
		"Read client configuration from file")
	cmd.Flags().StringVar(&flags.configRemoteName, "config-remote-name", "",
		"Override remote to use from configuration file")

	cmd.Flags().StringVar(&flags.baseImage, "base-image", defaultBaseImage,
		"Base image for launching builder instance (one of ubuntu:22.04|ubuntu:24.04|debian:12|debian:13)")

	cmd.Flags().StringVar(&flags.instanceName, "instance-name", defaultInstanceName,
		"Name for the builder instance")
	cmd.Flags().StringSliceVar(&flags.instanceProfiles, "instance-profile", defaultInstanceProfiles,
		"Profiles to use to launch the builder instance")

	cmd.Flags().StringVar(&flags.imageAlias, "image-alias", "",
		"Create image with alias. If not specified, a default is used based on config")

	cmd.Flags().StringSliceVar(&flags.skipStages, "skip", nil,
		"Skip stages while building the image")
	cmd.Flags().StringSliceVar(&flags.onlyStages, "only", nil,
		"Run specific stages while building the image")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false,
		"Dry run stages")
	cmd.Flags().StringVar(&flags.outputFile, "output", "image.tar.gz",
		"Output file for exported image")

	return cmd
}
