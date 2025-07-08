//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"slices"

	"sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/util"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QuickStart", func() {
	Context("KVM", Label("PRBlocking"), Label("Flaky"), func() {
		BeforeEach(func(ctx context.Context) {
			lxcClient, err := lxc.New(ctx, e2eCtx.Settings.LXCClientOptions)
			Expect(err).ToNot(HaveOccurred())

			// skip if server cannot launch kvm instances
			if err := lxcClient.SupportsInstanceKVM(); err != nil {
				Skip(fmt.Sprintf("Server cannot launch kvm instances: %v", err))
			}

			// skip if server is not amd64 (kvm images are only available for amd64)
			if archs := lxcClient.SupportsArchitectures(); !slices.Contains(archs, "x86_64") {
				Skip(fmt.Sprintf("QuickStart KVM test requires amd64, but server only supports the following architectures: %v", archs))
			}

			e2eCtx.OverrideVariables(map[string]string{
				"WORKER_MACHINE_TYPE": lxc.VirtualMachine,
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
				ControlPlaneMachineCount: ptr.To[int64](1),
				WorkerMachineCount:       ptr.To[int64](1),
				ClusterName:              ptr.To(fmt.Sprintf("capn-kvm-%s", util.RandomString(6))),
			}
		})
	})
})
