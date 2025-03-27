//go:build e2e

package e2e

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/util"

	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/ptr"
	"github.com/neoaggelos/cluster-api-provider-lxc/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QuickStart", func() {
	Context("OCI", Label("PRBlocking"), func() {
		BeforeEach(func(ctx context.Context) {
			client, err := incus.New(ctx, e2eCtx.Settings.LXCClientOptions)
			Expect(err).ToNot(HaveOccurred())

			err = client.SupportsInstanceOCI()
			Expect(err).To(Or(Succeed(), MatchError(incus.IsTerminalError, "IsTerminalError")))
			if err != nil {
				Skip(fmt.Sprintf("Server does not support OCI instances: %v", err))
			}
		})

		e2e.QuickStartSpec(context.TODO(), func() e2e.QuickStartSpecInput {
			return e2e.QuickStartSpecInput{
				E2EConfig:              e2eCtx.E2EConfig,
				ClusterctlConfigPath:   e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy:  e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:         e2eCtx.Settings.ArtifactFolder,
				SkipCleanup:            e2eCtx.Settings.SkipCleanup,
				PostNamespaceCreated:   e2eCtx.DefaultPostNamespaceCreated(),
				ControlPlaneWaiters:    e2eCtx.DefaultControlPlaneWaiters(),
				InfrastructureProvider: ptr.To("lxc:v0.88.99"),

				Flavor:                   ptr.To(shared.FlavorDefault),
				ControlPlaneMachineCount: ptr.To[int64](3),
				WorkerMachineCount:       ptr.To[int64](0),
				ClusterName:              ptr.To(fmt.Sprintf("quick-start-oci-%s", util.RandomString(6))),

				ClusterctlVariables: map[string]string{
					"LOAD_BALANCER": "oci: {}",
				},
			}
		})
	})
})
