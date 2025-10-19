package docker

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

// docker save -o /tmp/images-tar2590373421/images.tar registry.k8s.io/cluster-api/cluster-api-controller:v1.9.3
func newDockerImageSaveCmd(env Environment) *cobra.Command {
	var flags struct {
		Output   string
		Platform string
	}
	cmd := &cobra.Command{
		Use:           "save IMAGE ...",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker save", "flags", flags, "args", args)

			cacheDir := env.CacheDir()
			var opts []remote.Option
			if flags.Platform != "" {
				p, err := v1.ParsePlatform(flags.Platform)
				if err != nil {
					return fmt.Errorf("invalid platform %q: %w", flags.Platform, err)
				}
				opts = append(opts, remote.WithPlatform(*p))
			}

			imgMap := make(map[string]v1.Image, len(args))
			for _, arg := range args {
				// if image has been `docker load`ed, use the local tarball
				loadedFileName := filepath.Join(cacheDir, "loaded--"+strings.ReplaceAll(arg, "/", "--")+".tar")
				if img, err := crane.Load(loadedFileName); err == nil {
					log.V(4).Info("Using local image", "tag", arg, "path", loadedFileName)
					imgMap[arg] = img
					continue
				}

				// otherwise, attempt to pull remote image
				ref, err := name.ParseReference(arg)
				if err != nil {
					return fmt.Errorf("failed to parse image %q: %w", arg, err)
				}
				img, err := remote.Image(ref, opts...)
				if err != nil {
					return fmt.Errorf("failed to get image %q: %w", arg, err)
				}
				if cacheDir != "" {
					img = cache.Image(img, cache.NewFilesystemCache(cacheDir))
				}

				imgMap[arg] = img
			}

			if flags.Output == "" {
				return fmt.Errorf("--output flag is required")
			}

			if err := crane.MultiSave(imgMap, flags.Output); err != nil {
				return fmt.Errorf("failed to save image: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.Output, "output", "o", "", "write to file")
	cmd.Flags().StringVarP(&flags.Platform, "platform", "p", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH), "platform")

	return cmd
}
