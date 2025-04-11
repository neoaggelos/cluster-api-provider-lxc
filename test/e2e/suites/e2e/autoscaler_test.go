//go:build e2e

package e2e

import (
	"context"

	"sigs.k8s.io/cluster-api/test/e2e"

	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Autoscaler", func() {
	// TODO: investigate why ScaleToAndFromZero is failing to scale back up from zero
	e2e.AutoscalerSpec(context.TODO(), func() e2e.AutoscalerSpecInput {
		return e2e.AutoscalerSpecInput{
			E2EConfig:              e2eCtx.E2EConfig,
			ClusterctlConfigPath:   e2eCtx.Environment.ClusterctlConfigPath,
			BootstrapClusterProxy:  e2eCtx.Environment.BootstrapClusterProxy,
			ArtifactFolder:         e2eCtx.Settings.ArtifactFolder,
			SkipCleanup:            e2eCtx.Settings.SkipCleanup,
			PostNamespaceCreated:   e2eCtx.DefaultPostNamespaceCreated(),
			ControlPlaneWaiters:    e2eCtx.DefaultControlPlaneWaiters(),
			InfrastructureProvider: ptr.To("incus:v0.88.99"),

			Flavor: ptr.To(shared.FlavorAutoscaler),

			InfrastructureMachineTemplateKind: "lxcmachinetemplates",
			AutoscalerVersion:                 "v1.31.1",
			InstallOnManagementCluster:        true,
		}
	})
})
