//go:build e2e

package shared

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"sigs.k8s.io/cluster-api/test/framework/bootstrap"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"

	. "github.com/onsi/gomega"
)

// CreateKindBootstrapClusterAndLoadImages is the same as bootstrap.CreateKindBootstrapClusterAndLoadImages, but does not interact with the docker socket.
func CreateKindBootstrapClusterAndLoadImages(ctx context.Context, input bootstrap.CreateKindBootstrapClusterAndLoadImagesInput) bootstrap.ClusterProvider {
	clusterProvider := bootstrap.CreateKindBootstrapClusterAndLoadImages(ctx, bootstrap.CreateKindBootstrapClusterAndLoadImagesInput{
		Name:               input.Name,
		KubernetesVersion:  input.KubernetesVersion,
		RequiresDockerSock: input.RequiresDockerSock,
		IPFamily:           input.IPFamily,
		LogFolder:          input.LogFolder,
		ExtraPortMappings:  input.ExtraPortMappings,
		CustomNodeImage:    input.CustomNodeImage,
	})

	if err := LoadImagesToKindCluster(ctx, bootstrap.LoadImagesToKindClusterInput{
		Name:   input.Name,
		Images: input.Images,
	}); err != nil {
		clusterProvider.Dispose(ctx)
		Expect(err).ToNot(HaveOccurred(), "Could not load images") // re-surface the error to fail the test
	}

	return clusterProvider
}

// LoadImagesToKindCluster is bootstrap.LoadImagesToKindCluster, but uses the kind CLI.
func LoadImagesToKindCluster(ctx context.Context, input bootstrap.LoadImagesToKindClusterInput) error {
	for _, image := range input.Images {
		if err := loadImage(ctx, input.Name, image.Name); err != nil {
			switch image.LoadBehavior {
			case clusterctl.MustLoadImage:
				return fmt.Errorf("failed to load image %q into the kind cluster %q: %w", image.Name, input.Name, err)
			case clusterctl.TryLoadImage:
				Logf("[WARNING] Unable to load image %q into the kind cluster %q: %v", image.Name, input.Name, err)
			}
		}
	}

	return nil
}

func loadImage(ctx context.Context, clusterName string, image string) error {
	Logf("Loading image %s into the cluster", image)

	cmd := exec.CommandContext(ctx, "kind", "load", "docker-image", "--name", clusterName, image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
