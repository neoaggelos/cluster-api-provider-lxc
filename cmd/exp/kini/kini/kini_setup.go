package kini

import (
	"flag"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func newKiniSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup configuration",
	}

	klogFlags := &flag.FlagSet{}
	klog.InitFlags(klogFlags)
	klogFlags.VisitAll(func(f *flag.Flag) {
		f.Usage = "[logging] " + f.Usage
	})
	cmd.PersistentFlags().AddGoFlagSet(klogFlags)

	cmd.AddCommand(newKiniSetupRemotesCmd())
	cmd.AddCommand(newKiniSetupActivateEnvironmentCmd())
	cmd.AddCommand(newKiniSetupGenerateSecretCmd())

	return cmd
}
