package docker

import (
	"fmt"

	"github.com/spf13/cobra"
)

// docker rm -f -v c1-control-plane
// docker rm -f -v c1-control-plane c1-control-plane2
func newDockerRmCmd(env Environment) *cobra.Command {
	var flags struct {
		Force   bool
		Volumes bool
	}

	cmd := &cobra.Command{
		Use:           "rm INSTANCE",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker rm", "flags", flags, "args", args)

			lxcClient, err := env.Client(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			for _, arg := range args {
				if err := lxcClient.WaitForDeleteInstance(cmd.Context(), arg); err != nil {
					return fmt.Errorf("failed to delete instance %q: %w", arg, err)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&flags.Force, "force", "f", false, "Force delete")
	cmd.Flags().BoolVarP(&flags.Volumes, "volumes", "v", false, "Delete volumes")

	return cmd
}
