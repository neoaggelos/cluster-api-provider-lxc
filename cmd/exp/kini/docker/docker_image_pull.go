package docker

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

// docker pull kindest/node:v1.31.2@sha256:18fbefc20a7113353c7b75b5c869d7145a6abd6269154825872dc59c1329912e
func newDockerImagePullCmd(env Environment) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pull IMAGE ...",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker pull", "args", args)

			imageName := args[0]
			if !strings.Contains(imageName, "kindest/node") {
				log.V(1).Info("Refusing to pull a non kindest/node image", "image", imageName)
				return nil
			}

			lxcClient, err := env.Client(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			image, err := utils.ParseOCIImage(imageName)
			if err != nil {
				return fmt.Errorf("failed to parse OCI image %q: %w", imageName, err)
			}

			if env.KindInstances(cmd.Context()) {
				if err := lxcClient.PullImage(cmd.Context(), lxc.Image{
					Protocol: lxc.OCI,
					Server:   image.Server(),
					Alias:    image.Alias(),
				}); err != nil {
					return fmt.Errorf("failed to pull %q: %w", imageName, err)
				}

				return nil
			}

			// infer version from tag
			if digest := image.Digest(); digest != "" {
				log.Info("WARNING: Running in LXC mode, ignoring image digest", "image", imageName, "digest", digest)
			}
			tag := image.Tag()

			// TODO: allow using local images with alias `kini/VERSION`
			log.V(1).Info("Detected Kubernetes version from image", "image", image, "version", tag)

			if err := lxcClient.PullImage(cmd.Context(), lxc.CapnImage(fmt.Sprintf("kubeadm/%s", tag))); err != nil {
				return fmt.Errorf("failed to pull kubeadm image for version %q: %w", tag, err)
			}
			return nil
		},
	}

	return cmd
}
