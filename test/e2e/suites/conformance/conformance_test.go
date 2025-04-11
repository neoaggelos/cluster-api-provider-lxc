//go:build e2e

package conformance

import (
	"context"

	"sigs.k8s.io/cluster-api/test/e2e"

	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Conformance", Label("conformance"), func() {
	e2e.K8SConformanceSpec(context.TODO(), func() e2e.K8SConformanceSpecInput {
		return e2e.K8SConformanceSpecInput{
			E2EConfig:             e2eCtx.E2EConfig,
			ClusterctlConfigPath:  e2eCtx.Environment.ClusterctlConfigPath,
			BootstrapClusterProxy: e2eCtx.Environment.BootstrapClusterProxy,
			ArtifactFolder:        e2eCtx.Settings.ArtifactFolder,
			SkipCleanup:           e2eCtx.Settings.SkipCleanup,
			PostNamespaceCreated:  e2eCtx.DefaultPostNamespaceCreated(),
			ControlPlaneWaiters:   e2eCtx.DefaultControlPlaneWaiters(),

			Flavor: shared.FlavorDefault,
		}
	})
})
