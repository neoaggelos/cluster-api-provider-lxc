package incus

import "github.com/lxc/incus/v6/shared/api"

func (c *Client) GetAddressIfExists(state *api.InstanceState) string {
	if state == nil {
		return ""
	}
	for _, network := range state.Network {
		if network.Type == "loopback" {
			continue
		}

		for _, addr := range network.Addresses {
			// TODO(neoaggelos): care for addr.Family ipv4 vs ipv6
			if addr.Scope == "global" {
				return addr.Address
			}
		}
	}

	return ""
}
