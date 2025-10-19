package docker

import (
	"fmt"
	"os"
	"strings"

	incus "github.com/lxc/incus/v6/client"
	"github.com/spf13/cobra"
)

// docker cp /tmp/k8s-tar-extract-2873489325/kubernetes/server/bin/kubeadm kind-build-1760304352-1984474374:/usr/bin/kubeadm
func newDockerCpCmd(env Environment) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "cp",
		Args:          cobra.ExactArgs(2),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker cp", "args", args)

			src, dst := args[0], args[1]

			_, _, srcOk := strings.Cut(src, ":")
			dstInstance, dstPath, dstOk := strings.Cut(dst, ":")

			if srcOk {
				return fmt.Errorf("source %q not supported", src)
			}
			if !dstOk {
				return fmt.Errorf("destination %q not supported", dst)
			}

			srcStat, err := os.Stat(src)
			if err != nil {
				return fmt.Errorf("failed to stat source: %w", err)
			}

			lxcClient, err := env.Client(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			f, err := os.Open(src)
			if err != nil {
				return fmt.Errorf("failed to read source: %w", err)
			}

			if err := lxcClient.CreateInstanceFile(dstInstance, dstPath, incus.InstanceFileArgs{
				Content: f,
				Mode:    int(srcStat.Mode().Perm()),
			}); err != nil {
				return fmt.Errorf("failed to copy file to instance: %w", err)
			}

			return nil
		},
	}

	return cmd
}
