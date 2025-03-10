//go:build e2e

package e2e

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/test/e2e"

	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/ptr"
	"github.com/neoaggelos/cluster-api-provider-lxc/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QuickStart", func() {
	Context("OVN", Label("PRBlocking"), func() {
		var (
			lbAddress   string
			networkName string
		)
		BeforeEach(func(ctx context.Context) {
			client, err := incus.New(ctx, e2eCtx.Settings.LXCClientOptions)
			Expect(err).ToNot(HaveOccurred())
			networks, err := client.Client.GetNetworks()
			Expect(err).ToNot(HaveOccurred())

			// find network with the annotations below
			// -- user.capl.e2e.ovn-lb-address = "<ip address>"
			for _, network := range networks {
				if v, ok := network.Config["user.capl.e2e.ovn-lb-address"]; ok {
					networkName = network.Name
					lbAddress = v
					shared.Logf("Using OVN network %q with LoadBalancer address %q", networkName, lbAddress)
					return
				}
			}

			Skip("Did not find any network with configuration 'user.capl.e2e.ovn-lb-address'")
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

				ClusterctlVariables: map[string]string{
					"LOAD_BALANCER":                 fmt.Sprintf("ovn: {host: '%s', networkName: '%s'}", lbAddress, networkName),
					"CONTROL_PLANE_MACHINE_DEVICES": fmt.Sprintf("['eth0,type=nic,network=%s']", networkName),
					"WORKER_MACHINE_DEVICES":        fmt.Sprintf("['eth0,type=nic,network=%s']", networkName),
				},
			}
		})
	})
})
