package cmd

import (
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "import",
		Short:   "Import images in a simplestreams index",
		GroupID: "operations",
	}

	cmd.AddCommand(newImportImageCmd())
	cmd.AddCommand(newImportReleaseCmd())

	return cmd
}
