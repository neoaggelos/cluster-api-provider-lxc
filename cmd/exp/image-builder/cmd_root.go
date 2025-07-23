package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	logsv1 "k8s.io/component-base/logs/api/v1"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

var (
	cfg struct {
		// client configuration
		configFile       string
		configRemoteName string

		// base image configuration
		baseImage string

		// builder configuration
		instanceName     string
		instanceProfiles []string
		instanceType     string

		// image alias configuration
		imageAlias string

		// build step configuration
		skipStages          []string
		instanceGracePeriod time.Duration

		// output
		outputFile string
	}

	// runtime configuration
	lxcClient *lxc.Client

	rootCmd = &cobra.Command{
		Use:          "image-builder",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := logsv1.ValidateAndApply(logOptions, nil); err != nil {
				return fmt.Errorf("failed to configure logging: %w", err)
			}

			switch cfg.instanceType {
			case lxc.Container, lxc.VirtualMachine:
			default:
				return fmt.Errorf("invalid value for --instance-type argument %q, must be one of [container, virtual-machine]", cfg.instanceType)
			}

			switch cfg.baseImage {
			case "debian":
				cfg.baseImage = "debian:12"
			case "ubuntu":
				cfg.baseImage = "ubuntu:24.04"
			case "debian:12", "ubuntu:22.04", "ubuntu:24.04":
			default:
				return fmt.Errorf("invalid value for --base-image argument %q, must be one of [ubuntu:22.04, ubuntu:24.04, debian:12]", cfg.baseImage)
			}

			opts, err := lxc.ConfigurationFromLocal(cfg.configFile, cfg.configRemoteName, false)
			if err != nil {
				return fmt.Errorf("failed to read client credentials: %w", err)
			}

			lxcClient, err = lxc.New(gCtx, opts)
			if err != nil {
				return fmt.Errorf("failed to create incus client: %w", err)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddGroup(&cobra.Group{ID: "build", Title: "Available Image Types:"})
	rootCmd.AddCommand(kubeadmCmd, haproxyCmd)

	logsv1.AddFlags(logOptions, rootCmd.PersistentFlags())
	rootCmd.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	_ = rootCmd.PersistentFlags().MarkHidden("kubeconfig")
	_ = rootCmd.PersistentFlags().MarkHidden("log-text-info-buffer-size")
	_ = rootCmd.PersistentFlags().MarkHidden("log-flush-frequency")
	_ = rootCmd.PersistentFlags().MarkHidden("log-text-split-stream")
	_ = rootCmd.PersistentFlags().MarkHidden("logging-format")

	rootCmd.PersistentFlags().StringVar(&cfg.configFile, "config-file", "",
		"Read client configuration from file")
	rootCmd.PersistentFlags().StringVar(&cfg.configRemoteName, "config-remote-name", "",
		"Override remote to use from configuration file")

	rootCmd.PersistentFlags().StringVar(&cfg.baseImage, "base-image", defaultBaseImage,
		"Base image for launching builder instance (one of ubuntu:22.04|ubuntu:24.04|debian:12)")

	rootCmd.PersistentFlags().StringVar(&cfg.instanceName, "instance-name", defaultInstanceName,
		"Name for the builder instance")
	rootCmd.PersistentFlags().StringVar(&cfg.instanceType, "instance-type", defaultInstanceType,
		"Type of image to build (one of container|virtual-machine)")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.instanceProfiles, "instance-profile", defaultInstanceProfiles,
		"Profiles to use to launch the builder instance")

	rootCmd.PersistentFlags().StringVar(&cfg.imageAlias, "image-alias", "",
		"Create image with alias. If not specified, a default is used based on config")

	rootCmd.PersistentFlags().StringSliceVar(&cfg.skipStages, "skip", nil,
		"Skip stages while building the image")
	_ = rootCmd.PersistentFlags().MarkHidden("skip")

	rootCmd.PersistentFlags().DurationVar(&cfg.instanceGracePeriod, "instance-grace-period", defaultInstanceGracePeriod,
		"[advanced] Grace period before stopping instance, such that all disk writes complete")

	rootCmd.PersistentFlags().StringVar(&cfg.outputFile, "output", "image.tar.gz",
		"Output file for exported image")
}
