package docker

import (
	"fmt"

	"github.com/spf13/cobra"
)

// docker network ls --filter=name=^kind$ --format={{.ID}}
func newDockerNetworkLsCmd(_ Environment) *cobra.Command {
	var flags struct {
		Filter string
		Format string
	}

	cmd := &cobra.Command{
		Use:           "ls NETWORK",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker network ls", "flags", flags)

			if flags.Filter != "name=^kind$" {
				return fmt.Errorf("invalid filter %q", flags.Filter)
			}

			if flags.Format != "{{.ID}}" {
				return fmt.Errorf("invalid format %q", flags.Format)
			}

			fmt.Println("kind")
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.Format, "format", "", "Output format")
	cmd.Flags().StringVar(&flags.Filter, "filter", "", "Filter rules")

	return cmd
}
