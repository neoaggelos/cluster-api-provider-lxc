//go:build e2e

package e2e

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/util"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QuickStart", func() {
	Context("OCI", Label("PRBlocking"), func() {
		BeforeEach(func(ctx context.Context) {
			lxcClient, err := lxc.New(ctx, e2eCtx.Settings.LXCClientOptions)
			Expect(err).ToNot(HaveOccurred())

			err = lxcClient.SupportsInstanceOCI()
			Expect(err).To(Or(Succeed(), MatchError(utils.IsTerminalError, "IsTerminalError")))
			if err != nil {
				Skip(fmt.Sprintf("Server does not support OCI instances: %v", err))
			}

			e2eCtx.OverrideVariables(map[string]string{
				"LOAD_BALANCER": "oci: {}",
			})
		})

		e2e.QuickStartSpec(context.TODO(), func() e2e.QuickStartSpecInput {
			return e2e.QuickStartSpecInput{
				E2EConfig:              e2eCtx.E2EConfig,
				ClusterctlConfigPath:   e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy:  e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:         e2eCtx.Settings.ArtifactFolder,
				SkipCleanup:            e2eCtx.Settings.SkipCleanup,
				PostNamespaceCreated:   e2eCtx.DefaultPostNamespaceCreated(),
				InfrastructureProvider: ptr.To("incus:v0.88.99"),

				Flavor:                   ptr.To(shared.FlavorDefault),
				ControlPlaneMachineCount: ptr.To[int64](3),
				WorkerMachineCount:       ptr.To[int64](1),
				ClusterName:              ptr.To(fmt.Sprintf("capn-oci-%s", util.RandomString(4))),
			}
		})
	})
})
