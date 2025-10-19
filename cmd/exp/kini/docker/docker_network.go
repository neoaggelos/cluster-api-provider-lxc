package docker

import (
	"github.com/spf13/cobra"
)

// docker network ls --filter=name=^kind$ --format={{.ID}}
func newDockerNetworkCmd(env Environment) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "network",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newDockerNetworkLsCmd(env))
	return cmd
}
