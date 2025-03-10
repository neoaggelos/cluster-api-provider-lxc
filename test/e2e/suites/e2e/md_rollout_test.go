//go:build e2e

package e2e

import (
	"context"

	"sigs.k8s.io/cluster-api/test/e2e"

	"github.com/neoaggelos/cluster-api-provider-lxc/internal/ptr"
	"github.com/neoaggelos/cluster-api-provider-lxc/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("MDRollout", func() {
	e2e.MachineDeploymentRolloutSpec(context.TODO(), func() e2e.MachineDeploymentRolloutSpecInput {
		return e2e.MachineDeploymentRolloutSpecInput{
			E2EConfig:              e2eCtx.E2EConfig,
			ClusterctlConfigPath:   e2eCtx.Environment.ClusterctlConfigPath,
			BootstrapClusterProxy:  e2eCtx.Environment.BootstrapClusterProxy,
			ArtifactFolder:         e2eCtx.Settings.ArtifactFolder,
			SkipCleanup:            e2eCtx.Settings.SkipCleanup,
			PostNamespaceCreated:   e2eCtx.DefaultPostNamespaceCreated(),
			ControlPlaneWaiters:    e2eCtx.DefaultControlPlaneWaiters(),
			InfrastructureProvider: ptr.To("lxc:v0.88.99"),

			Flavor: shared.FlavorDevelopment,
		}
	})
})
