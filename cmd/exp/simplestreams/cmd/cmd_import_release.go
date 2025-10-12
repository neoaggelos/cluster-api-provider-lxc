package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lxc/cluster-api-provider-incus/internal/exp/simplestreams/index"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

func newImportReleaseCmd() *cobra.Command {
	var flags struct {
		rootDir string

		alias []string

		containerImage      string
		virtualMachineImage string
	}

	cmd := &cobra.Command{
		Use:   "release",
		Short: "Import kubeadm images for a Kubernetes release",

		RunE: func(cmd *cobra.Command, args []string) error {
			index, err := index.GetOrCreateIndex(flags.rootDir)
			if err != nil {
				return fmt.Errorf("failed to read simplestreams index: %w", err)
			}

			if flags.containerImage != "" {
				if err := index.ImportImage(cmd.Context(), lxc.Container, flags.containerImage, flags.alias, true, true); err != nil {
					return fmt.Errorf("failed to import container image: %w", err)
				}
			}

			if flags.virtualMachineImage != "" {
				if err := index.ImportImage(cmd.Context(), lxc.VirtualMachine, flags.virtualMachineImage, flags.alias, true, true); err != nil {
					return fmt.Errorf("failed to import Incus VM image: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.rootDir, "root-dir", "",
		"Simplestreams index directory")
	cmd.Flags().StringSliceVar(&flags.alias, "alias", nil,
		"alias to add to the images, e.g. 'kubeadm/v1.33.0,kubeadm/v1.33.0/ubuntu'")
	cmd.Flags().StringVar(&flags.containerImage, "container", "",
		"Path to kubeadm image for containers")
	cmd.Flags().StringVar(&flags.virtualMachineImage, "vm", "",
		"Path to kubeadm image for virtual machines")

	_ = cmd.MarkPersistentFlagRequired("version")

	return cmd
}
