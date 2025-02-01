package incus

import (
	"testing"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"

	. "github.com/onsi/gomega"
)

type mockClient_supportsServerExtensions struct {
	incus.InstanceServer

	supportsExtensions []string
}

func (c *mockClient_supportsServerExtensions) GetServer() (*api.Server, string, error) {
	return &api.Server{ServerUntrusted: api.ServerUntrusted{
		APIExtensions: c.supportsExtensions,
	}}, "", nil
}

func Test_serverSupportsExtensions(t *testing.T) {
	for _, tc := range []struct {
		name              string
		supports          []string
		want              []string
		expectUnsupported []string
	}{
		{
			name: "Nil",
		},
		{
			name:     "WantNone",
			supports: []string{"instance_a", "instance_b"},
		},
		{
			name:              "SupportsOne",
			supports:          []string{"instance_oci", "instance_a", "instance_b"},
			want:              []string{"instance_oci"},
			expectUnsupported: nil,
		},
		{
			name:              "SupportsMany",
			supports:          []string{"network_load_balancer", "network_load_balancer_health_check", "instance_a", "instance_b"},
			want:              []string{"network_load_balancer", "network_load_balancer_health_check"},
			expectUnsupported: nil,
		},
		{
			name:              "SupportsSome",
			supports:          []string{"network_load_balancer", "instance_a", "instance_b"},
			want:              []string{"network_load_balancer", "network_load_balancer_health_check"},
			expectUnsupported: []string{"network_load_balancer_health_check"},
		},
		{
			name:              "SupportsNotOne",
			supports:          []string{"instance_oci", "instance_a", "instance_b"},
			want:              []string{"network_load_balancer"},
			expectUnsupported: []string{"network_load_balancer"},
		},
		{
			name:              "SupportsNotMany",
			supports:          []string{"instance_oci", "instance_a", "instance_b"},
			want:              []string{"network_load_balancer", "network_load_balancer_health_check"},
			expectUnsupported: []string{"network_load_balancer", "network_load_balancer_health_check"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			c := &Client{Client: &mockClient_supportsServerExtensions{supportsExtensions: tc.supports}}
			unsupported, err := c.serverSupportsExtensions(tc.want...)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(unsupported).To(ConsistOf(tc.expectUnsupported))
		})
	}
}
