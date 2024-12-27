package incus

import "github.com/lxc/incus/v6/shared/api"

// ParseMachineAddressIfExists returns the main IP address of the instance.
// It filters for networks that have a host interface name (e.g. vethbbcd39c7), so that CNI addresses are ignored.
// It filters for addresses with global scope, so that IPv6 link-local addresses are ignored.
func (c *Client) ParseMachineAddressIfExists(state *api.InstanceState) string {
	if state == nil {
		return ""
	}
	for _, network := range state.Network {
		switch {
		case network.Type == "loopback":
			// ignore loopback address
			continue
		case network.HostName == "":
			// only consider networks with a matching interface name on the host
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
