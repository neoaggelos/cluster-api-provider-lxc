package kini

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/incus/v6/shared/api"
)

func newKiniSetupValidateConfigCmd() *cobra.Command {
	var flags struct {
		configFile string
		remoteName string

		namespace string
	}

	cmd := &cobra.Command{
		Use:           "validate-config",
		Short:         "Validate CLI configuration is usable",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, path, err := lxc.ConfigurationFromLocal(flags.configFile, flags.remoteName, false)
			if err != nil {
				return fmt.Errorf("failed to read local configuration: %w", err)
			}

			log.Info("Loaded configuration", "path", path)
			client, err := lxc.New(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			log.Info("Validating client configuration by listing running instances")
			if _, err := client.GetInstanceNames(api.InstanceTypeAny); err != nil {
				return fmt.Errorf("failed to list instances: %w", err)
			}

			log.Info("Credentials are OK!")

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.configFile, "config-file", "",
		"Read client configuration from file")
	cmd.Flags().StringVar(&flags.remoteName, "remote-name", "",
		"Override remote to use from configuration file")

	return cmd
}
