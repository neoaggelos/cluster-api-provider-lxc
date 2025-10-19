package docker

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/cobra"
)

// docker image inspect -f '{{ .Id }}' registry.k8s.io/cluster-api/cluster-api-controller:v1.9.3
func newDockerImageInspectCmd(env Environment) *cobra.Command {
	var flags struct {
		Format string
	}

	cmd := &cobra.Command{
		Use:           "inspect IMAGE",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker image inspect", "flags", flags)

			if flags.Format != "{{ .Id }}" {
				return fmt.Errorf("invalid format %q", flags.Format)
			}

			var (
				tag = args[0]
				img v1.Image
				err error
			)

			// if image has been `docker load`ed, use the local tarball
			loadedFileName := filepath.Join(env.CacheDir(), "loaded--"+strings.ReplaceAll(tag, "/", "--")+".tar")
			if img, err = crane.Load(loadedFileName); err == nil {
				log.V(4).Info("Using local image", "tag", tag, "path", loadedFileName)
			} else if img, err = crane.Pull(tag); err != nil {
				return fmt.Errorf("could not pull image %q: %w", tag, err)
			}

			if manifest, err := img.Manifest(); err != nil {
				return fmt.Errorf("could not get manifest of image: %w", err)
			} else {
				fmt.Println(manifest.Config.Digest.String())
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.Format, "format", "f", "", "Output format")

	return cmd
}
