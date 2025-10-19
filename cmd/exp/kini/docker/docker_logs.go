package docker

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// docker logs -f c1-control-plane
// docker logs c1-control-plane
func newDockerLogsCmd(env Environment) *cobra.Command {
	var flags struct {
		Follow bool
	}
	cmd := &cobra.Command{
		Use:           "logs INSTANCE",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker logs", "flags", flags, "args", args)

			lxcClient, err := env.Client(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			command := []string{"journalctl", "--boot", "--no-tail"}
			if flags.Follow {
				command = append(command, "--follow")
			}

			return lxcClient.RunCommand(cmd.Context(), args[0], command, env.Stdin, os.Stdout, os.Stderr)
		},
	}

	cmd.Flags().BoolVarP(&flags.Follow, "follow", "f", false, "follow logs")

	return cmd
}
