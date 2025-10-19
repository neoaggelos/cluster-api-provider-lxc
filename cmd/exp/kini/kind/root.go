package kind

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/cmd/kind"

	"github.com/lxc/cluster-api-provider-incus/internal/exp/kini"
)

var (
	log = ctrl.Log
)

func NewCmd() *cobra.Command {
	cmd := kind.NewCommand(cmd.NewLogger(), cmd.StandardIOStreams())
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.Short = "kind commands for kini"

	kindPreRunE := cmd.PersistentPreRunE
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := kindPreRunE(cmd, args); err != nil {
			return err
		}

		// use the --verbosity/-v flag from the kind command to set log level
		logFlags := &flag.FlagSet{}
		klog.InitFlags(logFlags)
		if verbosity := cmd.Flags().Lookup("verbosity").Value.String(); verbosity != "" {
			_ = logFlags.Set("v", verbosity)
		}

		// configure self for docker commands
		_, cleanup, err := kini.SetupEnvironment(cmd.Context(), true, false)
		if err != nil {
			return fmt.Errorf("failed to configure self as docker: %w", err)
		}
		cmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
			return cleanup()
		}

		log.V(1).Info("kind command invocation", "command", strings.Join(os.Args, " "))
		return nil
	}

	return cmd
}
