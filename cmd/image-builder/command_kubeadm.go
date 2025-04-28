package main

import (
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	kubeadmCmd = &cobra.Command{
		Use:          "kubeadm",
		GroupID:      "build",
		Short:        "Build kubeadm images for cluster-api-provider-incus",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := semver.ParseTolerant(kubeadmCfg.kubernetesVersion); err != nil {
				return fmt.Errorf("--kubernetes-version %q is not valid semver: %w", kubeadmCfg.kubernetesVersion, err)
			}

			if cfg.imageAlias == "" {
				cfg.imageAlias = fmt.Sprintf("kubeadm-%s-u%s-%s", kubeadmCfg.kubernetesVersion, cfg.ubuntuVersion, cfg.instanceType)
			}

			log.FromContext(gCtx).WithValues(
				"kubernetes-version", kubeadmCfg.kubernetesVersion,
				"ubuntu-version", cfg.ubuntuVersion,
				"instance-type", cfg.instanceType,
				"image-alias", cfg.imageAlias,
			).Info("Building kubeadm image")

			return runStages(
				&stageCreateInstance{},
				&stagePreRunCommands{},
				&stageInstallKubeadm{},
				&stagePullExtraImages{},
				&stageGenerateManifest{},
				&stagePostRunCommands{},
				&stageCleanupInstance{},
				&stageStopInstance{},
				&stagePublishKubeadmImage{},
				&stageExportImage{},
				&stageRemoveInstance{},
			)
		},
	}
)

func init() {
	kubeadmCmd.Flags().StringVar(&kubeadmCfg.kubernetesVersion, "kubernetes-version", "",
		"Kubernetes version to create image for")

	kubeadmCmd.Flags().StringSliceVar(&kubeadmCfg.pullExtraImages, "pull-extra-images", defaultPullExtraImages,
		"Extra OCI images to pull in the image")

	_ = kubeadmCmd.MarkFlagRequired("kubernetes-version")
}
