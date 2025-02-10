//go:build e2e

package e2e

import (
	"context"

	"sigs.k8s.io/cluster-api/test/e2e"

	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/ptr"
	"github.com/neoaggelos/cluster-api-provider-lxc/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QuickStart", Label("PRBlocking"), func() {
	Context("SingleNode", func() {
		e2e.QuickStartSpec(context.TODO(), func() e2e.QuickStartSpecInput {
			return e2e.QuickStartSpecInput{
				E2EConfig:             e2eCtx.E2EConfig,
				ClusterctlConfigPath:  e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy: e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:        e2eCtx.Settings.ArtifactFolder,
				SkipCleanup:           e2eCtx.Settings.SkipCleanup,
				PostNamespaceCreated:  e2eCtx.DefaultPostNamespaceCreated(),
				ControlPlaneWaiters:   e2eCtx.DefaultControlPlaneWaiters(),

				Flavor:                   ptr.To(shared.FlavorDefault),
				ControlPlaneMachineCount: ptr.To[int64](1),
				WorkerMachineCount:       ptr.To[int64](0),
			}
		})
	})
	Context("Full", func() {
		e2e.QuickStartSpec(context.TODO(), func() e2e.QuickStartSpecInput {
			return e2e.QuickStartSpecInput{
				E2EConfig:             e2eCtx.E2EConfig,
				ClusterctlConfigPath:  e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy: e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:        e2eCtx.Settings.ArtifactFolder,
				SkipCleanup:           e2eCtx.Settings.SkipCleanup,
				PostNamespaceCreated:  e2eCtx.DefaultPostNamespaceCreated(),
				ControlPlaneWaiters:   e2eCtx.DefaultControlPlaneWaiters(),

				Flavor:                   ptr.To(shared.FlavorDefault),
				ControlPlaneMachineCount: ptr.To[int64](3),
				WorkerMachineCount:       ptr.To[int64](3),
			}
		})
	})
	Context("Ubuntu", func() {
		BeforeEach(func() {
			e2eCtx.OverrideVariables(map[string]string{
				"LXC_IMAGE_NAME":  "ubuntu:24.04",
				"INSTALL_KUBEADM": "true",
			})
		})
		e2e.QuickStartSpec(context.TODO(), func() e2e.QuickStartSpecInput {
			return e2e.QuickStartSpecInput{
				E2EConfig:             e2eCtx.E2EConfig,
				ClusterctlConfigPath:  e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy: e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:        e2eCtx.Settings.ArtifactFolder,
				SkipCleanup:           e2eCtx.Settings.SkipCleanup,
				PostNamespaceCreated:  e2eCtx.DefaultPostNamespaceCreated(),
				ControlPlaneWaiters:   e2eCtx.DefaultControlPlaneWaiters(),

				Flavor:                   ptr.To(shared.FlavorDefault),
				ControlPlaneMachineCount: ptr.To[int64](1),
				WorkerMachineCount:       ptr.To[int64](1),
			}
		})
	})
	Context("OCI", func() {
		BeforeEach(func() {
			client, err := incus.New(context.TODO(), e2eCtx.Settings.LXCClientOptions)
			Expect(err).ToNot(HaveOccurred())

			err = client.SupportsInstanceOCI()
			Expect(err).To(Or(Succeed(), MatchError(incus.IsTerminalError, "IsTerminalError")))
			if err != nil {
				Skip("Server does not support OCI instances")
			}

			e2eCtx.OverrideVariables(map[string]string{
				"LOAD_BALANCER": "oci: {}",
			})
		})

		e2e.QuickStartSpec(context.TODO(), func() e2e.QuickStartSpecInput {
			return e2e.QuickStartSpecInput{
				E2EConfig:             e2eCtx.E2EConfig,
				ClusterctlConfigPath:  e2eCtx.Environment.ClusterctlConfigPath,
				BootstrapClusterProxy: e2eCtx.Environment.BootstrapClusterProxy,
				ArtifactFolder:        e2eCtx.Settings.ArtifactFolder,
				SkipCleanup:           e2eCtx.Settings.SkipCleanup,
				PostNamespaceCreated:  e2eCtx.DefaultPostNamespaceCreated(),
				ControlPlaneWaiters:   e2eCtx.DefaultControlPlaneWaiters(),

				Flavor:                   ptr.To(shared.FlavorDefault),
				ControlPlaneMachineCount: ptr.To[int64](3),
				WorkerMachineCount:       ptr.To[int64](0),
			}
		})
	})
})
