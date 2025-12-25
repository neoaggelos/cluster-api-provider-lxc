package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/lxc/cluster-api-provider-incus/internal/exp/simplestreams/sync"
)

func newSyncCmd() *cobra.Command {
	var flags struct {
		rootDir    string
		stagingDir string

		manifest string
	}

	cmd := &cobra.Command{
		Use:     "sync",
		Short:   "Sync images from an upstream source into a simplestreams index",
		GroupID: "operations",

		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := os.ReadFile(flags.manifest)
			if err != nil {
				return fmt.Errorf("failed to open manifest: %w", err)
			}

			var manifest sync.Manifest
			if err := yaml.UnmarshalStrict(b, &manifest); err != nil {
				return fmt.Errorf("failed to parse manifest: %w", err)
			}

			mgr, err := sync.NewManager(flags.rootDir, flags.stagingDir, http.DefaultClient)
			if err != nil {
				return fmt.Errorf("failed to initialize manager: %w", err)
			}

			if err := mgr.Sync(cmd.Context(), manifest); err != nil {
				return fmt.Errorf("failed to sync images: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.rootDir, "root-dir", "",
		"Simplestreams index directory")
	cmd.Flags().StringVar(&flags.stagingDir, "staging-dir", "/tmp",
		"Staging directory to download images")
	cmd.Flags().StringVar(&flags.manifest, "manifest", "",
		"Path to manifest of images to sync")

	_ = cmd.MarkPersistentFlagRequired("version")

	return cmd
}
