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
	Context("KubeVIP", Label("PRBlocking"), func() {
		var (
			clusterctlVariables map[string]string
		)
		BeforeEach(func(ctx context.Context) {
			if v := e2eCtx.E2EConfig.GetVariableBestEffort(shared.KubeVIPAddress); v != "" {
				shared.Logf("Using kube-vip address %q (from environment variable KUBE_VIP_ADDRESS)", v)
				clusterctlVariables = map[string]string{
					"LOAD_BALANCER": fmt.Sprintf("kube-vip: {host: %q}", v),
				}
				return
			}

			// KUBE_VIP_ADDRESS is not set, look for a network
			client, err := incus.New(ctx, e2eCtx.Settings.LXCClientOptions)
			Expect(err).ToNot(HaveOccurred())
			networks, err := client.Client.GetNetworks()
			Expect(err).ToNot(HaveOccurred())

			// find network with the annotations below
			// -- user.capl.e2e.kube-vip-address = "<ip address>"
			for _, network := range networks {
				if v, ok := network.Config["user.capl.e2e.kube-vip-address"]; ok {
					shared.Logf("Using kube-vip address %q (from network %q)", v, network.Name)
					clusterctlVariables = map[string]string{
						"LOAD_BALANCER":                 fmt.Sprintf("kube-vip: {host: '%s'}", v),
						"CONTROL_PLANE_MACHINE_DEVICES": fmt.Sprintf("['eth0,type=nic,network=%s']", network.Name),
						"WORKER_MACHINE_DEVICES":        fmt.Sprintf("['eth0,type=nic,network=%s']", network.Name),
					}
					return
				}
			}

			Skip("Did not find any network with configuration 'user.capl.e2e.kube-vip-address', and KUBE_VIP_ADDRESS is not set")
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
				ClusterName:              ptr.To(fmt.Sprintf("quick-start-kube-vip-%s", util.RandomString(6))),

				ClusterctlVariables: clusterctlVariables,
			}
		})
	})
})
