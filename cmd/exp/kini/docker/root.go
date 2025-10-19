package docker

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

var (
	log = ctrl.Log
)

func NewCmd() *cobra.Command {
	env := Environment{
		Stdin: os.Stdin,

		Client: func(ctx context.Context) (*lxc.Client, error) {
			opts, _, err := lxc.ConfigurationFromLocal(os.Getenv("KINI_CONFIG"), os.Getenv("KINI_REMOTE"), false)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve local configuration: %w", err)
			}
			opts.Project = os.Getenv("KINI_PROJECT")

			return lxc.New(ctx, opts)
		},

		Getenv: os.Getenv,
	}

	cmd := &cobra.Command{
		Use:           "docker",
		Short:         "docker commands for kini",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logFlags := &flag.FlagSet{}
			klog.InitFlags(logFlags)

			if logFile := os.Getenv("KINI_DOCKER_LOG"); logFile != "" {
				_ = logFlags.Set("logtostderr", "false")
				_ = logFlags.Set("log_file", logFile)
				_ = logFlags.Set("alsologtostderr", "true")
				_ = logFlags.Set("skip_log_headers", "true")
				_ = logFlags.Set("v", "5")
			}
			if verbosity := os.Getenv("KINI_DOCKER_LOGV"); verbosity != "" {
				_ = logFlags.Set("v", verbosity)
			}

			log.V(1).Info("docker command invocation", "command", strings.Join(os.Args, " "))

			return nil
		},
		Version: "kini",
	}

	cmd.AddCommand(newDockerCpCmd(env))
	cmd.AddCommand(newDockerExecCmd(env))
	cmd.AddCommand(newDockerImageCmd(env))
	cmd.AddCommand(newDockerImageLoadCmd(env)) // "docker load" same as "docker image load"
	cmd.AddCommand(newDockerImagePullCmd(env)) // "docker pull" same as "docker image pull"
	cmd.AddCommand(newDockerImageSaveCmd(env)) // "docker save" same as "docker image save"
	cmd.AddCommand(newDockerInfoCmd(env))
	cmd.AddCommand(newDockerInspectCmd(env))
	cmd.AddCommand(newDockerLogsCmd(env))
	cmd.AddCommand(newDockerNetworkCmd(env))
	cmd.AddCommand(newDockerPsCmd(env))
	cmd.AddCommand(newDockerRmCmd(env))
	cmd.AddCommand(newDockerRunCmd(env))

	return cmd
}
