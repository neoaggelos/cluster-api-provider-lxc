//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/test/framework"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SelfHosted", func() {
	var (
		preloadImages string
	)
	BeforeEach(func(ctx context.Context) {
		preloadImages = e2eCtx.E2EConfig.GetVariableOrEmpty(shared.PreloadImages)
		switch preloadImages {
		case "":
			Skip(fmt.Sprintf("%s is not set", shared.PreloadImages))
		case "auto":
			shared.Logf("Generating tarball with preload images for nodes of self-hosted cluster")

			preloadImages = fmt.Sprintf("%s/images.tar", GinkgoT().TempDir())

			dockerSaveCommand := make([]string, 0, len(e2eCtx.E2EConfig.Images)+3)
			dockerSaveCommand = append(dockerSaveCommand, "save")
			for _, image := range e2eCtx.E2EConfig.Images {
				dockerSaveCommand = append(dockerSaveCommand, image.Name)
			}
			dockerSaveCommand = append(dockerSaveCommand, "-o", preloadImages)

			shared.Logf("docker %v", strings.Join(dockerSaveCommand, " "))
			cmd := exec.CommandContext(ctx, "docker", dockerSaveCommand...)
			cmd.Stderr = os.Stderr
			Expect(exec.CommandContext(ctx, "docker", dockerSaveCommand...).Run()).To(Succeed())

			shared.Logf("Tarball generated at %s", preloadImages)
		default:
			// expected to be an absolute path to an existing tarball
			Expect(filepath.IsAbs(preloadImages), fmt.Sprintf("%s must be an absolute path", shared.PreloadImages))
		}
	})
	e2e.SelfHostedSpec(context.TODO(), func() e2e.SelfHostedSpecInput {
		return e2e.SelfHostedSpecInput{
			E2EConfig:              e2eCtx.E2EConfig,
			ClusterctlConfigPath:   e2eCtx.Environment.ClusterctlConfigPath,
			BootstrapClusterProxy:  e2eCtx.Environment.BootstrapClusterProxy,
			ArtifactFolder:         e2eCtx.Settings.ArtifactFolder,
			SkipCleanup:            e2eCtx.Settings.SkipCleanup,
			InfrastructureProvider: ptr.To("incus:v0.88.99"),

			Flavor:                   shared.FlavorDefault,
			SkipUpgrade:              true,
			ControlPlaneMachineCount: ptr.To[int64](1),
			WorkerMachineCount:       ptr.To[int64](1),

			// We hijack PostNamespaceCreated to allow pre-loading images into the workload cluster
			// This is required for the CAPN provider e2e image, which is not published
			PostNamespaceCreated: func(managementClusterProxy framework.ClusterProxy, workloadClusterNamespace string) {
				if managementClusterProxy.GetKubeconfigPath() != e2eCtx.Environment.BootstrapClusterProxy.GetKubeconfigPath() {
					// we are running for the workload cluster, load the images
					shared.Logf("Preloading images into workload cluster")

					lxcClient, err := lxc.New(context.TODO(), e2eCtx.Settings.LXCClientOptions)
					Expect(err).ToNot(HaveOccurred())

					instances, err := lxcClient.ListInstances(context.TODO(), lxc.WithConfig(map[string]string{
						"user.cluster-namespace": workloadClusterNamespace,
						"user.cluster-role":      "worker",
					}))
					Expect(err).ToNot(HaveOccurred())
					Expect(instances).To(HaveLen(1)) // must match number of worker nodes

					for _, instance := range instances {
						shared.Logf("Preloading images into %s", instance.Name)

						f, err := os.Open(preloadImages)
						Expect(err).ToNot(HaveOccurred())

						// command to load image tarball adapted from kind
						var stderr bytes.Buffer
						command := []string{"ctr", "--namespace=k8s.io", "images", "import", "--all-platforms", "--digests", "-"}
						err = lxcClient.RunCommand(context.TODO(), instance.Name, command, f, nil, &stderr)
						Expect(err).ToNot(HaveOccurred(), "Failed to load images into %s. command=%q. stderr=%q", instance.Name, strings.Join(command, " "), stderr.String())
						Expect(f.Close()).To(Succeed())
					}
				} else {
					// we are running for the management cluster, create the infrastructure secret
					shared.FixupNamespace(e2eCtx, workloadClusterNamespace, true, true)
				}
			},
		}
	})
})
