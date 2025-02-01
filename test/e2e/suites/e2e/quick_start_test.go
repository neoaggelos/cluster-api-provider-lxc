//go:build e2e

package e2e

import (
	"context"

	"sigs.k8s.io/cluster-api/test/e2e"

	"github.com/neoaggelos/cluster-api-provider-lxc/internal/ptr"
	"github.com/neoaggelos/cluster-api-provider-lxc/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("QuickStart", Label("PRBlocking"), func() {
	Context("SingleNode", func() {
		e2e.QuickStartSpec(context.TODO(), func() e2e.QuickStartSpecInput {
			return e2e.QuickStartSpecInput{
				E2EConfig:                e2eCtx.E2EConfig,
				ClusterctlConfigPath:     e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy:    e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:           e2eCtx.Settings.ArtifactFolder,
				Flavor:                   ptr.To(shared.FlavorDevelopment),
				SkipCleanup:              e2eCtx.Settings.SkipCleanup,
				ControlPlaneMachineCount: ptr.To[int64](1),
				WorkerMachineCount:       ptr.To[int64](0),
				PostNamespaceCreated:     e2eCtx.DefaultPostNamespaceCreated(),
				ControlPlaneWaiters:      e2eCtx.DefaultControlPlaneWaiters(),
			}
		})
	})
	Context("Full", func() {
		e2e.QuickStartSpec(context.TODO(), func() e2e.QuickStartSpecInput {
			return e2e.QuickStartSpecInput{
				E2EConfig:                e2eCtx.E2EConfig,
				ClusterctlConfigPath:     e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy:    e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:           e2eCtx.Settings.ArtifactFolder,
				Flavor:                   ptr.To(shared.FlavorDevelopment),
				SkipCleanup:              e2eCtx.Settings.SkipCleanup,
				ControlPlaneMachineCount: ptr.To[int64](3),
				WorkerMachineCount:       ptr.To[int64](3),
				PostNamespaceCreated:     e2eCtx.DefaultPostNamespaceCreated(),
				ControlPlaneWaiters:      e2eCtx.DefaultControlPlaneWaiters(),
			}
		})
	})
})
