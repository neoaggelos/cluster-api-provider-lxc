package kini

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lxc/cluster-api-provider-incus/internal/exp/kini"
)

func newKiniSetupActivateEnvironmentCmd() *cobra.Command {
	var flags struct {
		docker bool
		kind   bool
	}
	cmd := &cobra.Command{
		Use:           "activate-environment",
		Short:         "Shadow kind and docker commands with kini",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if path, _, err := kini.SetupEnvironment(cmd.Context(), flags.docker, flags.kind); err != nil {
				return fmt.Errorf("failed to setup environment: %w", err)
			} else {
				fmt.Printf("export PATH=%s:$PATH\n", path)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.docker, "docker", true, "create symlink for docker binary")
	cmd.Flags().BoolVar(&flags.kind, "kind", true, "create symlink for kind binary")

	return cmd
}
