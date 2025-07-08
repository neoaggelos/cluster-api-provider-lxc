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
	Context("OVN", Label("PRBlocking"), func() {
		BeforeEach(func(ctx context.Context) {
			lxcClient, err := lxc.New(ctx, e2eCtx.Settings.LXCClientOptions)
			Expect(err).ToNot(HaveOccurred())

			err = lxcClient.SupportsNetworkLoadBalancers()
			Expect(err).To(Or(Succeed(), MatchError(utils.IsTerminalError, "IsTerminalError")))
			if err != nil {
				Skip(fmt.Sprintf("Server does not support network load balancer: %v", err))
			}

			networks, err := lxcClient.GetNetworks()
			Expect(err).ToNot(HaveOccurred())

			// find network with the annotations below
			// -- user.capn.e2e.ovn-lb-address = "<ip address>"
			for _, network := range networks {
				if lbAddress, ok := network.Config["user.capn.e2e.ovn-lb-address"]; ok {
					shared.Logf("Using OVN network %q with LoadBalancer address %q", network.Name, lbAddress)

					e2eCtx.OverrideVariables(map[string]string{
						"LOAD_BALANCER":                 fmt.Sprintf("ovn: {host: '%s', networkName: '%s'}", lbAddress, network.Name),
						"CONTROL_PLANE_MACHINE_DEVICES": fmt.Sprintf("['eth0,type=nic,network=%s']", network.Name),
						"WORKER_MACHINE_DEVICES":        fmt.Sprintf("['eth0,type=nic,network=%s']", network.Name),
					})
					return
				}
			}

			Skip("Did not find any network with configuration 'user.capn.e2e.ovn-lb-address'")
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
				WorkerMachineCount:       ptr.To[int64](0),
				ClusterName:              ptr.To(fmt.Sprintf("capn-ovn-%s", util.RandomString(6))),
			}
		})
	})
})
