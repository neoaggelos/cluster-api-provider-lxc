package kini

import (
	"fmt"

	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/lxc/cluster-api-provider-incus/cmd/exp/kini/docker"
	"github.com/lxc/cluster-api-provider-incus/cmd/exp/kini/kind"
	"github.com/lxc/cluster-api-provider-incus/internal/exp/kini"
)

var (
	log = ctrl.Log
)

func addCommands(root *cobra.Command, group *cobra.Group, commands ...*cobra.Command) {
	root.AddGroup(group)

	for _, cmd := range commands {
		cmd.GroupID = group.ID
	}

	root.AddCommand(commands...)
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kini",
		Short: `kini creates and manages local Kubernetes clusters using LXC container 'nodes'`,
		Long: `kini creates and manages local Kubernetes clusters using LXC container 'nodes'.

kini is part of the cluster-api-provider-incus project (https://capn.linuxcontainers.org). It is
implemented as a wrapper around kind (https://sigs.k8s.io/kind). It replaces the "docker" CLI with
a shim executable so that it creates LXC containers instead.

kini can be used as a standalone binary and does not require kind or docker to be installed. The
kind sub-command allows running kind commands. For example, to create a single-node development
cluster on a local machine (where Incus is already installed), you can use:

	$ kini kind create cluster

kini can also be used with an existing kind binary. You can do this as follows:

	$ . <(kini setup activate-environment)
	$ docker --version              # should print "docker version kini"
	$ kind create cluster           # run kind commands, it should use kini as a docker CLI shim
`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			_, cleanup, err := kini.SetupEnvironment(cmd.Context(), true, false)
			if err != nil {
				return fmt.Errorf("failed to setup environment: %w", err)
			}
			cmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
				return cleanup()
			}

			return nil
		},
	}

	cmd.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)

	addCommands(cmd,
		&cobra.Group{ID: "setup", Title: "Setup commands:"},
		newKiniSetupCmd(),
	)
	addCommands(cmd,
		&cobra.Group{ID: "commands", Title: "Shim commands:"},
		kind.NewCmd(),
		docker.NewCmd(),
	)

	return cmd
}
