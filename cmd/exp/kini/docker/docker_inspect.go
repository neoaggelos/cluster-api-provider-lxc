package docker

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// docker inspect c1-control-plane
// docker inspect --format '{{ index .Config.Labels "io.x-k8s.kind.role"}}' c1-control-plane
// docker inspect --format '{{ index .Config.Labels "desktop.docker.io/ports/6443/tcp" }}' c1-control-plane
// docker inspect --format '{{ with (index (index .NetworkSettings.Ports "6443/tcp") 0) }}{{ printf "%s\t%s" .HostIp .HostPort }}{{ end }}' c1-control-plane
// docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}},{{.GlobalIPv6Address}}{{end}}' c1-control-plane
func newDockerInspectCmd(env Environment) *cobra.Command {
	var flags struct {
		Format string
	}

	cmd := &cobra.Command{
		Use:           "inspect INSTANCE",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker inspect", "flags", flags, "args", args)

			lxcClient, err := env.Client(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			instance, _, err := lxcClient.GetInstanceFull(args[0])
			if err != nil {
				return fmt.Errorf("failed to retrieve instance %q: %w", args[0], err)
			}

			switch flags.Format {
			case ``:
				b, err := yaml.Marshal(instance)
				if err != nil {
					return fmt.Errorf("failed to marshal YAML: %w", err)
				}
				fmt.Println(string(b))
				return nil
			case `{{ index .Config.Labels "io.x-k8s.kind.role"}}`:
				fmt.Println(instance.Config["user.io.x-k8s.kind.role"])
				return nil
			case `{{ index .Config.Labels "desktop.docker.io/ports/6443/tcp" }}`:
				// TODO: if we are in environment where we have to (e.g: remote server, etc)
				// print the apiserver address e.g. "10.0.1.7:6443"
				return nil
			case "{{ with (index (index .NetworkSettings.Ports \"6443/tcp\") 0) }}{{ printf \"%s\t%s\" .HostIp .HostPort }}{{ end }}", `test`:
				for _, device := range instance.Devices {
					if device["type"] != "proxy" {
						continue
					}
					if device["bind"] != "host" {
						continue
					}
					if device["connect"] != "tcp::6443" {
						continue
					}

					parts := strings.Split(device["listen"], ":")
					if parts[0] != "tcp" || len(parts) != 3 {
						continue
					}
					fmt.Printf("%s\t%s\n", parts[1], parts[2])
					break
				}
				return nil
			case `{{range .NetworkSettings.Networks}}{{.IPAddress}},{{.GlobalIPv6Address}}{{end}}`, `t2`:
				addrs := lxc.ParseHostAddresses(instance.State)
				var ipv4, ipv6 string
				for _, addr := range addrs {
					if ip, err := netip.ParseAddr(addr); err == nil {
						if ip.Is4() {
							ipv4 = addr
						} else if ip.Is6() {
							ipv6 = addr
						}
					}
				}
				fmt.Printf("%s,%s\n", ipv4, ipv6)
				return nil
			default:
				return fmt.Errorf("unknown format %q", flags.Format)
			}
		},
	}

	cmd.Flags().StringVarP(&flags.Format, "format", "f", "", "Output format")

	return cmd
}
