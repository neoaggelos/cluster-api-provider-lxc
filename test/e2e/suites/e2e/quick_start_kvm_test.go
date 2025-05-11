//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/util"

	"github.com/lxc/cluster-api-provider-incus/internal/incus"
	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QuickStart", func() {
	Context("KVM", Label("PRBlocking"), Label("Flaky"), func() {
		BeforeEach(func(ctx context.Context) {
			client, err := incus.New(ctx, e2eCtx.Settings.LXCClientOptions)
			Expect(err).ToNot(HaveOccurred())
			info, _, err := client.Client.GetServer()
			Expect(err).ToNot(HaveOccurred())

			// skip if server cannot launch kvm instances
			if !slices.Contains(strings.Split(info.Environment.Driver, " | "), "qemu") {
				Skip(fmt.Sprintf("Server is missing driver qemu, supported drivers are: %q", info.Environment.Driver))
			}

			// skip if server is not amd64 (kvm images are only available for amd64)
			if !slices.Contains(info.Environment.Architectures, "x86_64") {
				Skip(fmt.Sprintf("QuickStart KVM test requires amd64, but server only supports the following architectures: %v", info.Environment.Architectures))
			}

			e2eCtx.OverrideVariables(map[string]string{
				"WORKER_MACHINE_TYPE": "virtual-machine",
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
				ClusterName:              ptr.To(fmt.Sprintf("quick-start-kvm-%s", util.RandomString(6))),
			}
		})
	})
})
