package kini

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/defaults"
	kindversion "sigs.k8s.io/kind/pkg/cmd/kind/version"
)

// Set with -X cmd/exp/kini/kini/kini_version.Version=v0.8.2
var version = "dev"

func newKiniVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print kini version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("kini version %s (%s %s/%s)\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
			fmt.Printf("kind version v%s (default image %s)\n", kindversion.Version(), defaults.Image)

			return nil
		},
	}
}
