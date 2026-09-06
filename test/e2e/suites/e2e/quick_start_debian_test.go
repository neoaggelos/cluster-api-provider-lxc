//go:build e2e

package e2e

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/util"

	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("QuickStart", func() {
	Context("Debian", func() {
		BeforeEach(func(ctx context.Context) {
			e2eCtx.OverrideVariables(map[string]string{
				"KUBERNETES_VERSION": e2eCtx.E2EConfig.MustGetVariable(shared.KubernetesVersionInstallKubeadm),
				"LXC_IMAGE_NAME":     shared.DebianImage,
				"INSTALL_KUBEADM":    "true",
			})
		})
		Context("Privileged", Label("PRBlocking"), func() {
			e2e.QuickStartSpec(context.TODO(), func() e2e.QuickStartSpecInput {
				return e2e.QuickStartSpecInput{
					E2EConfig:              e2eCtx.E2EConfig,
					ClusterctlConfigPath:   e2eCtx.Environment.ClusterctlConfigPath,
					BootstrapClusterProxy:  e2eCtx.Environment.BootstrapClusterProxy,
					ArtifactFolder:         e2eCtx.Settings.ArtifactFolder,
					SkipCleanup:            e2eCtx.Settings.SkipCleanup,
					PostNamespaceCreated:   e2eCtx.DefaultPostNamespaceCreated(),
					InfrastructureProvider: new("incus:v0.88.99"),

					Flavor:                   new(shared.FlavorDefault),
					ControlPlaneMachineCount: new(int64(1)),
					WorkerMachineCount:       new(int64(1)),
					ClusterName:              new(fmt.Sprintf("capn-debian-%s", util.RandomString(4))),
				}
			})
		})

		Context("Unprivileged", func() {
			e2e.QuickStartSpec(context.TODO(), func() e2e.QuickStartSpecInput {
				return e2e.QuickStartSpecInput{
					E2EConfig:              e2eCtx.E2EConfig,
					ClusterctlConfigPath:   e2eCtx.Environment.ClusterctlConfigPath,
					BootstrapClusterProxy:  e2eCtx.Environment.BootstrapClusterProxy,
					ArtifactFolder:         e2eCtx.Settings.ArtifactFolder,
					SkipCleanup:            e2eCtx.Settings.SkipCleanup,
					PostNamespaceCreated:   e2eCtx.DefaultPostNamespaceCreated(),
					InfrastructureProvider: new("incus:v0.88.99"),

					Flavor:                   new(shared.FlavorDefault),
					ControlPlaneMachineCount: new(int64(1)),
					WorkerMachineCount:       new(int64(1)),
					ClusterName:              new(fmt.Sprintf("capn-debian-unprivileged-%s", util.RandomString(4))),

					ClusterctlVariables: e2eCtx.UnprivilegedContainersClusterVariables(),
				}
			})
		})
	})
})
