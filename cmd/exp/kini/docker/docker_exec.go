package docker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	incus "github.com/lxc/incus/v6/client"
	"github.com/spf13/cobra"

	"github.com/lxc/cluster-api-provider-incus/internal/loadbalancer"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// docker exec --privileged c1-control-plane cat /etc/kubernetes/admin.conf
// docker exec --privileged -i c1-control-plane cp /dev/stdin /kind/kubeadm.yaml
// docker exec --privileged -i c1-control-plane kubectl create --kubeconfig=/etc/kubernetes/admin.conf -f -
// docker exec --privileged -i c1-control-plane kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f -
// docker exec --privileged -i c1-control-plane ctr --namespace=k8s.io images import --all-platforms --digests --snapshotter=overlayfs -
func newDockerExecCmd(env Environment) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "exec INSTANCE COMMAND ...",
		Args:               cobra.MinimumNArgs(2),
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableFlagParsing: true, // do not parse flags, as they will passed through as command-line to the instance
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker exec", "args", args)

			// ignore --privileged and -i flags
			for args[0] == "--privileged" || args[0] == "-i" {
				if len(args) == 1 {
					return fmt.Errorf("instance name not specified")
				}
				args = args[1:]
			}

			lxcClient, err := env.Client(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			instanceName := args[0]

			// docker exec $instance cp /dev/stdin $destination
			if len(args) == 4 && args[1] == "cp" && args[2] == "/dev/stdin" {
				b, err := io.ReadAll(env.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}

				// HACK: "docker exec --privileged -i test-external-load-balancer cp /dev/stdin /usr/local/etc/haproxy/haproxy.cfg"
				// Default haproxy configuration requires DNS names, which does not work as expected in Incus.
				if args[3] == "/usr/local/etc/haproxy/haproxy.cfg" {
					if b, err = loadbalancer.GenerateHaproxyLoadBalancerConfiguration(cmd.Context(), lxcClient, lxc.WithConfig(map[string]string{
						"user.io.x-k8s.kind.cluster": strings.TrimSuffix(instanceName, "-external-load-balancer"),
						"user.io.x-k8s.kind.role":    "control-plane",
					})); err != nil {
						return fmt.Errorf("failed to generate hack haproxy configuration: %w", err)
					}
				}

				if err := lxcClient.CreateInstanceFile(instanceName, args[3], incus.InstanceFileArgs{
					Content: bytes.NewReader(b),
					Type:    "file",
					Mode:    0o644,
				}); err != nil {
					return fmt.Errorf("failed to create file %q on instance %q: %w", args[3], instanceName, err)
				}

				return nil
			}

			// docker exec $instance $command...
			return lxcClient.RunCommand(cmd.Context(), instanceName, args[1:], env.Stdin, os.Stdout, os.Stderr)
		},
	}

	return cmd
}
