package cmd

import (
	"fmt"
	"runtime"
	"time"

	"github.com/blang/semver/v4"
	"github.com/lxc/incus/v6/shared/api"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/cmd/exp/image-builder/internal/action"
	"github.com/lxc/cluster-api-provider-incus/cmd/exp/image-builder/internal/stage"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/static"
)

func newKubeadmCmd() *cobra.Command {
	var flags struct {
		// client configuration
		configFile       string
		configRemoteName string

		// base image configuration
		baseImage string

		// builder configuration
		instanceName           string
		instanceProfiles       []string
		instanceType           string
		validationInstanceName string

		// image alias configuration
		imageAlias string

		// build step configuration
		skipStages              []string
		onlyStages              []string
		dryRun                  bool
		instanceStopGracePeriod time.Duration

		// output
		outputFile         string
		outputManifestFile string

		// kubeadm configuration
		kubernetesVersion string
		pullExtraImages   []string
	}

	cmd := &cobra.Command{
		Use:     "kubeadm",
		GroupID: "build",
		Short:   "Build kubeadm images for cluster-api-provider-incus",

		PreRunE: func(cmd *cobra.Command, args []string) error {
			switch flags.instanceType {
			case lxc.Container, lxc.VirtualMachine:
			default:
				return fmt.Errorf("invalid value for --instance-type argument %q, must be one of [container, virtual-machine]", flags.instanceType)
			}

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
				flags.imageAlias = fmt.Sprintf("kubeadm-%s-%s", flags.kubernetesVersion, flags.instanceType)
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := semver.ParseTolerant(flags.kubernetesVersion); err != nil {
				return fmt.Errorf("--kubernetes-version %q is not valid semver: %w", flags.kubernetesVersion, err)
			}
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
					WithProfiles(flags.instanceProfiles).
					MaybeWithImage(image).
					WithInstanceType(api.InstanceType(flags.instanceType)),
				)},
				// {Name: "pre-run-commands", Action: action.ExecInstance(lxcClient, flags.instanceName, <TODO>, <TODO>)},
				{Name: "install-kubeadm", Action: action.ExecInstance(lxcClient, flags.instanceName, static.InstallKubeadmScript(), flags.kubernetesVersion)},
				{Name: "pull-extra-images", Action: action.ExecInstance(lxcClient, flags.instanceName, static.PullImagesScript(), flags.pullExtraImages...)},
				{Name: "generate-manifest", Action: action.ExecInstance(lxcClient, flags.instanceName, static.GenerateManifestScript())},
				{Name: "export-manifest", Action: action.CopyInstanceFile(lxcClient, flags.instanceName, "/opt/manifest.txt", flags.outputManifestFile)},
				// {Name: "post-run-commands", Action: action.ExecInstance(lxcClient, flags.instanceName, <TODO>, <TODO>)},
				{Name: "prepare-instance", Action: action.ExecInstance(lxcClient, flags.instanceName, static.CleanupInstanceScript())},
				{Name: "stop-grace-period", Action: action.Wait(flags.instanceStopGracePeriod)},
				{Name: "stop-instance", Action: action.StopInstance(lxcClient, flags.instanceName)},
				{Name: "publish-image", Action: action.PublishImage(lxcClient, flags.instanceName, flags.imageAlias, action.PublishImageInfo{
					Name:               fmt.Sprintf("kubeadm %s %s %s", flags.kubernetesVersion, wellKnownBaseImages[flags.baseImage].fullName, runtime.GOARCH),
					OperatingSystem:    "kubeadm",
					Release:            flags.kubernetesVersion,
					Variant:            wellKnownBaseImages[flags.baseImage].variantName,
					LXCRequireCgroupv2: true,
				})},
				{Name: "export-image", Action: action.ExportImage(lxcClient, flags.imageAlias, flags.outputFile)},
				{Name: "delete-instance", Action: action.DeleteInstance(lxcClient, flags.instanceName)},
				{Name: "validate-image", Action: action.Chain(
					action.LaunchInstance(lxcClient, flags.validationInstanceName, (&lxc.LaunchOptions{}).
						MaybeWithImage(api.InstanceSource{Type: "image", Alias: flags.imageAlias}).
						WithInstanceType(api.InstanceType(flags.instanceType)).
						WithProfiles(flags.instanceProfiles),
					),
					action.ExecInstance(lxcClient, flags.validationInstanceName, static.ValidateKubeadmImageScript()),
					action.DeleteInstance(lxcClient, flags.validationInstanceName),
				)},
			}

			log.FromContext(cmd.Context()).WithValues(
				"kubernetes-version", flags.kubernetesVersion,
				"base-image", flags.baseImage,
				"instance-type", flags.instanceType,
				"image-alias", flags.imageAlias,
			).Info("Building kubeadm image")

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
	cmd.Flags().StringVar(&flags.instanceType, "instance-type", defaultInstanceType,
		"Type of image to build (one of container|virtual-machine)")
	cmd.Flags().StringSliceVar(&flags.instanceProfiles, "instance-profile", defaultInstanceProfiles,
		"Profiles to use to launch the builder instance")
	cmd.Flags().StringVar(&flags.validationInstanceName, "validation-instance-name", defaultValidationInstanceName,
		"Name for the builder instance")

	cmd.Flags().StringVar(&flags.imageAlias, "image-alias", "",
		"Create image with alias. If not specified, a default is used based on config")

	cmd.Flags().StringSliceVar(&flags.skipStages, "skip", nil,
		"Skip stages while building the image")
	cmd.Flags().StringSliceVar(&flags.onlyStages, "only", nil,
		"Run specific stages while building the image")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false,
		"Dry run stages")
	cmd.Flags().DurationVar(&flags.instanceStopGracePeriod, "instance-stop-grace-period", defaultInstanceStopGracePeriod,
		"[advanced] Grace period before stopping instance, such that all disk writes complete")

	cmd.Flags().StringVar(&flags.outputFile, "output", "image.tar.gz",
		"Output file for exported image")
	cmd.Flags().StringVar(&flags.outputManifestFile, "manifest", "image.txt",
		"Output file for exported image manifest")

	cmd.Flags().StringVar(&flags.kubernetesVersion, "kubernetes-version", "",
		"Kubernetes version to create image for")
	cmd.Flags().StringSliceVar(&flags.pullExtraImages, "pull-extra-images", defaultPullExtraImages,
		"Extra OCI images to pull in the image")

	_ = cmd.MarkFlagRequired("kubernetes-version")

	return cmd
}
