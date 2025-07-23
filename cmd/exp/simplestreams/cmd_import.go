package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

var (
	importCfg struct {
		imagePath    string
		imageAliases []string

		serverType string // "lxd" or "incus"
	}

	importCmd = &cobra.Command{
		Use:     "import",
		GroupID: "operations",
		Short:   "Import images into a simplestreams index",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			switch importCfg.serverType {
			case lxc.Incus, lxc.LXD:
			default:
				return fmt.Errorf("invalid value for --server-type argument %q, must be one of [incus, lxd]", importCfg.serverType)
			}

			log.Info("Ensure simplestreams directories", "rootDir", cfg.rootDir)
			if err := os.MkdirAll(filepath.Join(cfg.rootDir, "streams", "v1"), 0755); err != nil {
				return fmt.Errorf("failed to create streams/v1 directory: %w", err)
			}
			if err := os.MkdirAll(filepath.Join(cfg.rootDir, "images"), 0755); err != nil {
				return fmt.Errorf("failed to create images directory: %w", err)
			}

			return nil
		},
	}
)

func init() {
	importCmd.AddCommand(importContainerCmd, importVirtualMachineCmd)

	importCmd.PersistentFlags().StringVar(&importCfg.imagePath, "image-path", "",
		"Path to image unified tarball to add in the simplestreams index directory")
	importCmd.PersistentFlags().StringSliceVar(&importCfg.imageAliases, "image-alias", nil,
		"List of aliases to add to the image. This is ignored if the product exists already")
	importCmd.PersistentFlags().StringVar(&importCfg.serverType, "server-type", lxc.Incus,
		"Server to create simplestreams index for. Must be one of [incus, lxd]")

	_ = importCmd.MarkPersistentFlagRequired("image-path")
}
