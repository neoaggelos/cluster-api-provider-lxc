package docker

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// docker info
// docker info --format '{{json .}}'
// docker info --format '{{json .SecurityOptions}}'
func newDockerInfoCmd(env Environment) *cobra.Command {
	var (
		flags struct {
			Format string
		}

		securityOptionsByFormatAndPrivileged = map[string]map[bool]string{
			"{{json .}}": {
				true:  `{"CgroupDriver":"systemd","CGroupVersion":"2","MemoryLimit":true,"CPUShares":true,"PidsLimit":true,"SecurityOptions":[]}`,
				false: `{"CgroupDriver":"systemd","CGroupVersion":"2","MemoryLimit":true,"CPUShares":true,"PidsLimit":true,"SecurityOptions":["name=userns","name=rootless"]}`,
			},
			"'{{json .SecurityOptions}}'": {
				true:  `[]`,
				false: `["name=userns","name=rootless"]`,
			},
		}
	)

	cmd := &cobra.Command{
		Use:           "info",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker info", "flags", flags)

			if flags.Format == "" {
				lxcClient, err := env.Client(cmd.Context())
				if err != nil {
					return fmt.Errorf("failed to initialize client: %w", err)
				}

				info, _, err := lxcClient.GetServer()
				if err != nil {
					return fmt.Errorf("failed to get server info: %w", err)
				}

				b, err := yaml.Marshal(info)
				if err != nil {
					return fmt.Errorf("failed to marshal YAML: %w", err)
				}
				fmt.Println(string(b))
				return nil
			}

			opts, ok := securityOptionsByFormatAndPrivileged[flags.Format]
			if !ok {
				return fmt.Errorf("unknown format %q", flags.Format)
			}

			fmt.Println(opts[env.Privileged()])

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.Format, "format", "", "Output format")

	return cmd
}
