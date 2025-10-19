package kini

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lxc/cluster-api-provider-incus/cmd/exp/kini/kini/config"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

func newKiniSetupRemotesCmd() *cobra.Command {
	var flags struct {
		configFile string

		dryRun bool
	}

	cmd := &cobra.Command{
		Use:           "remotes",
		Short:         "Setup simplestreams image remotes in local config",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := config.NewManager(cmd.Context(), flags.configFile)
			if err != nil {
				return fmt.Errorf("failed to initialize config manager: %w", err)
			}

			if err := manager.AddSimplestreamsRemoteIfNotExist(cmd.Context(), "capi", lxc.DefaultSimplestreamsServer); err != nil {
				return fmt.Errorf("failed to add capi images remote: %w", err)
			}
			if err := manager.AddSimplestreamsRemoteIfNotExist(cmd.Context(), "capi-stg", lxc.DefaultStagingSimplestreamsServer); err != nil {
				return fmt.Errorf("failed to add capi-stg images remote: %w", err)
			}

			if flags.dryRun {
				log.Info("Not updating config file", "dry-run", true)
				return nil
			}

			if err := manager.Commit(cmd.Context()); err != nil {
				return fmt.Errorf("failed to commit changes: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.configFile, "config-file", "",
		"Read client configuration from file")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false,
		"Do not update any configuration files")

	return cmd
}
