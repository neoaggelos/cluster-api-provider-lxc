package docker

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func parseImageTag(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	var manifestJSON []struct {
		RepoTags []string
	}

	tarReader := tar.NewReader(f)
	for {
		if hdr, err := tarReader.Next(); err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("no manifest.json found in image tarball")
			}
		} else if hdr.Name == "manifest.json" {
			if err := json.NewDecoder(tarReader).Decode(&manifestJSON); err != nil {
				return "", fmt.Errorf("failed to parse manifest.json: %w", err)
			}
			break
		}
	}

	if len(manifestJSON) >= 1 && len(manifestJSON[0].RepoTags) >= 1 {
		return manifestJSON[0].RepoTags[0], nil
	}
	return "", fmt.Errorf("no image tags found in manifest %#v", manifestJSON)
}

// docker load /tmp/images-tar2590373421/images.tar
func newDockerImageLoadCmd(env Environment) *cobra.Command {
	var flags struct {
		Input string
	}
	cmd := &cobra.Command{
		Use:           "load",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker load", "flags", flags)

			cacheDir := env.CacheDir()
			if cacheDir == "" {
				return fmt.Errorf("load command requires KINI_CACHE to be set")
			}

			// copy STDIN to temporary file -- we will need to read the file either way
			if flags.Input == "" {
				f, err := os.CreateTemp("", "")
				if err != nil {
					return fmt.Errorf("failed to create a temporary file: %w", err)
				}
				defer func() {
					_ = os.Remove(f.Name())
				}()
				if _, err := io.Copy(f, env.Stdin); err != nil {
					return fmt.Errorf("failed to write stdin to temporary file: %w", err)
				} else if err := f.Sync(); err != nil {
					return fmt.Errorf("failed to sync stdin to temporary file: %w", err)
				}

				flags.Input = f.Name()
			}

			// parse image tag or fail
			imageTag, err := parseImageTag(flags.Input)
			if err != nil {
				return fmt.Errorf("no image tag found in %q: %w", flags.Input, err)
			}

			// store in cache folder
			outFileName := filepath.Join(cacheDir, "loaded--"+strings.ReplaceAll(imageTag, "/", "--")+".tar")
			log.V(4).Info("Saving docker image to local cache", "tag", imageTag, "path", outFileName)

			inFile, err := os.Open(flags.Input)
			if err != nil {
				return fmt.Errorf("failed to open %q: %w", flags.Input, err)
			}
			defer func() {
				_ = inFile.Close()
			}()
			outFile, err := os.Create(outFileName)
			if err != nil {
				return fmt.Errorf("failed to create output file %q: %w", outFileName, err)
			}
			defer func() {
				_ = outFile.Close()
			}()
			if _, err := io.Copy(outFile, inFile); err != nil {
				return fmt.Errorf("failed to save tarball in %q: %w", outFileName, err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.Input, "input", "i", "", "input from file instead of STDIN")

	return cmd
}
