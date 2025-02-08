//go:build e2e

package e2e

import (
	"context"

	"sigs.k8s.io/cluster-api/test/e2e"

	"github.com/neoaggelos/cluster-api-provider-lxc/internal/ptr"
	"github.com/neoaggelos/cluster-api-provider-lxc/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("ClusterUpgrade", func() {
	Context("Default", Label("PRBlocking"), func() {
		e2e.ClusterUpgradeConformanceSpec(context.TODO(), func() e2e.ClusterUpgradeConformanceSpecInput {
			return e2e.ClusterUpgradeConformanceSpecInput{
				E2EConfig:             e2eCtx.E2EConfig,
				ClusterctlConfigPath:  e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy: e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:        e2eCtx.Settings.ArtifactFolder,
				SkipCleanup:           e2eCtx.Settings.SkipCleanup,

				Flavor:               ptr.To(shared.FlavorDevelopment),
				ControlPlaneWaiters:  e2eCtx.DefaultControlPlaneWaiters(),
				PostNamespaceCreated: e2eCtx.DefaultPostNamespaceCreated(),

				SkipConformanceTests: true,
			}
		})
	})
	Context("HAControlPlane", Label("full"), func() {
		e2e.ClusterUpgradeConformanceSpec(context.TODO(), func() e2e.ClusterUpgradeConformanceSpecInput {
			return e2e.ClusterUpgradeConformanceSpecInput{
				E2EConfig:             e2eCtx.E2EConfig,
				ClusterctlConfigPath:  e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy: e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:        e2eCtx.Settings.ArtifactFolder,
				SkipCleanup:           e2eCtx.Settings.SkipCleanup,

				Flavor:                   ptr.To(shared.FlavorDevelopment),
				ControlPlaneMachineCount: ptr.To[int64](3),
				WorkerMachineCount:       ptr.To[int64](1),
				ControlPlaneWaiters:      e2eCtx.DefaultControlPlaneWaiters(),
				PostNamespaceCreated:     e2eCtx.DefaultPostNamespaceCreated(),

				SkipConformanceTests: true,
			}
		})
	})
})
