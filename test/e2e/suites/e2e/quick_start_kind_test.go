//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"io"

	"sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/test/framework"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"
	"sigs.k8s.io/cluster-api/util"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func applyDefaultKindCNI(ctx context.Context, input clusterctl.ApplyCustomClusterTemplateAndWaitInput, result *clusterctl.ApplyCustomClusterTemplateAndWaitResult) {
	lxcClient, err := lxc.New(ctx, e2eCtx.Settings.LXCClientOptions)
	Expect(err).ToNot(HaveOccurred())

	instances, err := lxcClient.ListInstances(ctx, lxc.WithConfig(map[string]string{
		"user.cluster-name":      input.ClusterName,
		"user.cluster-namespace": input.Namespace,
		"user.cluster-role":      "control-plane",
	}))
	Expect(err).ToNot(HaveOccurred())
	Expect(instances).ToNot(BeEmpty())

	shared.Logf("Reading default kind CNI from instance %s", instances[0].Name)
	reader, _, err := lxcClient.GetInstanceFile(instances[0].Name, "/kind/manifests/default-cni.yaml")
	Expect(err).ToNot(HaveOccurred())
	defer reader.Close()
	b, err := io.ReadAll(reader)
	Expect(err).ToNot(HaveOccurred())

	shared.Logf("Applying default kind CNI")
	Expect(input.ClusterProxy.GetWorkloadCluster(ctx, input.Namespace, input.ClusterName).CreateOrUpdate(ctx, b)).To(Succeed())

	shared.Logf("Waiting for ControlPlane nodes to become Ready")
	framework.WaitForControlPlaneAndMachinesReady(ctx, framework.WaitForControlPlaneAndMachinesReadyInput{
		GetLister:    input.ClusterProxy.GetClient(),
		Cluster:      result.Cluster,
		ControlPlane: result.ControlPlane,
	}, input.WaitForControlPlaneIntervals...)
}

var _ = Describe("QuickStart", func() {
	Context("Kind", Label("PRBlocking"), func() {
		BeforeEach(func(ctx context.Context) {
			lxcClient, err := lxc.New(ctx, e2eCtx.Settings.LXCClientOptions)
			Expect(err).ToNot(HaveOccurred())

			err = lxcClient.SupportsInstanceOCI()
			Expect(err).To(Or(Succeed(), MatchError(utils.IsTerminalError, "IsTerminalError")))
			if err != nil {
				Skip(fmt.Sprintf("Server does not support OCI instances: %v", err))
			}

			e2eCtx.OverrideVariables(map[string]string{
				"LOAD_BALANCER":              "oci: {}",
				"CONTROL_PLANE_MACHINE_TYPE": "kind",
				"WORKER_MACHINE_TYPE":        "kind",
				"DEPLOY_KUBE_FLANNEL":        "false", // we use kindnet instead
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
				ClusterName:              ptr.To(fmt.Sprintf("capn-kind-%s", util.RandomString(6))),

				ControlPlaneWaiters: clusterctl.ControlPlaneWaiters{
					WaitForControlPlaneMachinesReady: applyDefaultKindCNI,
				},
			}
		})
	})
})
