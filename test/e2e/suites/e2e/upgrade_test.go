//go:build e2e

package e2e

import (
	"context"

	"sigs.k8s.io/cluster-api/test/e2e"

	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("ClusterUpgrade", func() {
	Context("Default", Label("PRBlocking"), func() {
		e2e.ClusterUpgradeConformanceSpec(context.TODO(), func() e2e.ClusterUpgradeConformanceSpecInput {
			return e2e.ClusterUpgradeConformanceSpecInput{
				E2EConfig:              e2eCtx.E2EConfig,
				ClusterctlConfigPath:   e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy:  e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:         e2eCtx.Settings.ArtifactFolder,
				SkipCleanup:            e2eCtx.Settings.SkipCleanup,
				PostNamespaceCreated:   e2eCtx.DefaultPostNamespaceCreated(),
				InfrastructureProvider: ptr.To("incus:v0.88.99"),

				Flavor:                   ptr.To(shared.FlavorDefault),
				ControlPlaneMachineCount: ptr.To[int64](1),
				WorkerMachineCount:       ptr.To[int64](1),

				SkipConformanceTests: true,
			}
		})
	})
	Context("HAControlPlane", func() {
		e2e.ClusterUpgradeConformanceSpec(context.TODO(), func() e2e.ClusterUpgradeConformanceSpecInput {
			return e2e.ClusterUpgradeConformanceSpecInput{
				E2EConfig:              e2eCtx.E2EConfig,
				ClusterctlConfigPath:   e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy:  e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:         e2eCtx.Settings.ArtifactFolder,
				SkipCleanup:            e2eCtx.Settings.SkipCleanup,
				PostNamespaceCreated:   e2eCtx.DefaultPostNamespaceCreated(),
				InfrastructureProvider: ptr.To("incus:v0.88.99"),

				Flavor:                   ptr.To(shared.FlavorDefault),
				ControlPlaneMachineCount: ptr.To[int64](3),
				WorkerMachineCount:       ptr.To[int64](2),

				SkipConformanceTests: true,
			}
		})
	})
})
