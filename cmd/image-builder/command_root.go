package main

import (
	"flag"
	"fmt"

	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	logsv1 "k8s.io/component-base/logs/api/v1"

	"github.com/lxc/cluster-api-provider-incus/internal/incus"
)

var (
	rootCmd = &cobra.Command{
		Use:          "image-builder",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := logsv1.ValidateAndApply(logOptions, nil); err != nil {
				return fmt.Errorf("failed to configure logging: %w", err)
			}

			switch cfg.instanceType {
			case "container", "virtual-machine":
			default:
				return fmt.Errorf("invalid value for --instance-type argument %q, must be one of [container, virtual-machine]", cfg.instanceType)
			}

			switch cfg.ubuntuVersion {
			case "22.04", "24.04":
			default:
				return fmt.Errorf("invalid value for --ubuntu-version argument %q, must be one of [22.04, 24.04]", cfg.ubuntuVersion)
			}

			opts, err := incus.NewOptionsFromConfigFile(cfg.configFile, cfg.configRemoteName, false)
			if err != nil {
				return fmt.Errorf("failed to read client credentials: %w", err)
			}

			client, err = incus.New(gCtx, opts)
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

	rootCmd.PersistentFlags().StringVar(&cfg.configFile, "config-file", "",
		"Read client configuration from file")
	rootCmd.PersistentFlags().StringVar(&cfg.configRemoteName, "config-remote-name", "",
		"Override remote to use from configuration file")

	rootCmd.PersistentFlags().StringVar(&cfg.ubuntuVersion, "ubuntu-version", defaultUbuntuVersion,
		"Ubuntu version to use to launch instance (one of 22.04|24.04)")

	rootCmd.PersistentFlags().StringVar(&cfg.instanceName, "instance-name", defaultInstanceName,
		"Name for the builder instance")
	rootCmd.PersistentFlags().StringVar(&cfg.instanceType, "instance-type", defaultInstanceType,
		"Type of image to build (one of container|virtual-machine)")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.instanceProfiles, "instance-profile", defaultInstanceProfiles,
		"Profiles to use to launch the builder instance")

	rootCmd.PersistentFlags().StringVar(&cfg.imageAlias, "image-alias", "",
		"Create image with alias. If not specified, a default is used based on config")

	rootCmd.PersistentFlags().StringVar(&cfg.outputFile, "output", "image.tar.gz",
		"Output file for exported image")

}
