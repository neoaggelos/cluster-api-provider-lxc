package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lxc/cluster-api-provider-incus/internal/exp/simplestreams/index"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

func newImportImageCmd() *cobra.Command {
	var flags struct {
		rootDir string

		imagePath    string
		imageAliases []string
		imageType    string // "virtual-machine" or "container"

		incus bool
		lxd   bool
	}
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Import a single image into a simplestreams index",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !flags.incus && !flags.lxd {
				return fmt.Errorf("at least one of --incus or --lxd must be set")
			}
			switch flags.imageType {
			case lxc.VirtualMachine, lxc.Container:
			default:
				return fmt.Errorf("invalid value for --image-type argument %q, must be one of [container, virtual-machine]", flags.imageType)
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			index, err := index.GetOrCreateIndex(flags.rootDir)
			if err != nil {
				return fmt.Errorf("failed to read simplestreams index: %w", err)
			}

			if err := index.ImportImage(cmd.Context(), flags.imageType, flags.imagePath, flags.imageAliases, flags.incus, flags.lxd); err != nil {
				return fmt.Errorf("failed to import image into simplestreams index: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.rootDir, "root-dir", "",
		"Simplestreams index directory")
	cmd.Flags().StringVar(&flags.imagePath, "image-path", "",
		"Path to image unified tarball to add in the simplestreams index directory")
	cmd.Flags().StringVar(&flags.imageType, "image-type", "",
		"Type of image. Must be one [container, virtual-machine]")
	cmd.Flags().StringSliceVar(&flags.imageAliases, "image-alias", nil,
		"List of aliases to add to the image. This is ignored if the product exists already")
	cmd.Flags().BoolVar(&flags.incus, "incus", false,
		"Import an image for Incus")
	cmd.Flags().BoolVar(&flags.lxd, "lxd", false,
		"Import an image for Canonical LXD")

	_ = cmd.MarkPersistentFlagRequired("image-path")
	_ = cmd.MarkPersistentFlagRequired("server-type")

	return cmd
}
