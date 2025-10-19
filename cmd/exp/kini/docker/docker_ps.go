package docker

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// docker ps -a --filter label=io.x-k8s.kind.cluster=$name --format {{.Names}}
// docker ps -a --filter label=io.x-k8s.kind.cluster --format {{.Label "io.x-k8s.kind.cluster"}}
func newDockerPsCmd(env Environment) *cobra.Command {
	var flags struct {
		Format string
		Filter string
		All    bool
	}

	cmd := &cobra.Command{
		Use:           "ps",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker ps", "flags", flags)

			lxcClient, err := env.Client(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			var filter lxc.ListInstanceFilter
			if clusterName, hasPrefix := strings.CutPrefix(flags.Filter, "label=io.x-k8s.kind.cluster="); hasPrefix {
				filter = lxc.WithConfig(map[string]string{"user.io.x-k8s.kind.cluster": clusterName})
			} else if flags.Filter == "label=io.x-k8s.kind.cluster" {
				filter = lxc.WithConfigKeys("user.io.x-k8s.kind.cluster")
			} else {
				return fmt.Errorf("unknown filter %q", flags.Filter)
			}

			instances, err := lxcClient.ListInstances(cmd.Context(), filter)
			if err != nil {
				return fmt.Errorf("failed to list instances: %w", err)
			}

			switch flags.Format {
			case `{{.Names}}`:
				for _, instance := range instances {
					fmt.Println(instance.Name)
				}
			case `{{.Label "io.x-k8s.kind.cluster"}}`:
				clusterNames := sets.New[string]()
				for _, instance := range instances {
					if v := instance.Config["user.io.x-k8s.kind.cluster"]; len(v) > 0 {
						clusterNames.Insert(v)
					}
				}

				fmt.Println(strings.Join(clusterNames.UnsortedList(), "\n"))
			default:
				return fmt.Errorf("unknown format %q", flags.Format)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&flags.All, "all", "a", false, "Show all containers")
	cmd.Flags().StringVar(&flags.Format, "format", "", "Output format")
	cmd.Flags().StringVar(&flags.Filter, "filter", "", "Filter rules")

	return cmd
}
