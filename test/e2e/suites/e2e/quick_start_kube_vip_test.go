//go:build e2e

package e2e

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/util"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/test/e2e/shared"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QuickStart", func() {
	Context("KubeVIP", Label("PRBlocking"), func() {
		BeforeEach(func(ctx context.Context) {
			if v := e2eCtx.E2EConfig.GetVariableOrEmpty(shared.KubeVIPAddress); v != "" {
				shared.Logf("Using kube-vip address %q (from environment variable KUBE_VIP_ADDRESS)", v)
				e2eCtx.OverrideVariables(map[string]string{
					"LOAD_BALANCER": fmt.Sprintf("kube-vip: {host: %q}", v),
				})
				return
			}

			// KUBE_VIP_ADDRESS is not set, look for a network
			lxcClient, err := lxc.New(ctx, e2eCtx.Settings.LXCClientOptions)
			Expect(err).ToNot(HaveOccurred())
			networks, err := lxcClient.GetNetworks()
			Expect(err).ToNot(HaveOccurred())

			// find network with the annotations below
			// -- user.capn.e2e.kube-vip-address = "<ip address>"
			for _, network := range networks {
				if v, ok := network.Config["user.capn.e2e.kube-vip-address"]; ok {
					shared.Logf("Using kube-vip address %q (from network %q)", v, network.Name)
					e2eCtx.OverrideVariables(map[string]string{
						"LOAD_BALANCER":                 fmt.Sprintf("kube-vip: {host: '%s'}", v),
						"CONTROL_PLANE_MACHINE_DEVICES": fmt.Sprintf("['eth0,type=nic,network=%s']", network.Name),
						"WORKER_MACHINE_DEVICES":        fmt.Sprintf("['eth0,type=nic,network=%s']", network.Name),
					})
					return
				}
			}

			Skip("Did not find any network with configuration 'user.capn.e2e.kube-vip-address', and KUBE_VIP_ADDRESS is not set")
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
				ClusterName:              ptr.To(fmt.Sprintf("capn-kube-vip-%s", util.RandomString(6))),
			}
		})
	})
})
