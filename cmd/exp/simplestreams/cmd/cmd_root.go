package cmd

import (
	"flag"

	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "simplestreams",
		SilenceUsage: true,
	}

	// logging flags
	klog.InitFlags(nil)
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		f.Usage = "[logging] " + f.Usage
	})
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	cmd.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)

	cmd.AddGroup(&cobra.Group{ID: "operations", Title: "Available operations:"})
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newImportCmd())

	return cmd
}
