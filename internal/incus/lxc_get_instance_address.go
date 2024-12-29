package incus

import (
	"slices"

	"github.com/lxc/incus/v6/shared/api"
)

// ParseActiveMachineAddresses returns the main IP addresses of the instance.
// It filters for networks that have a host interface name (e.g. vethbbcd39c7), so that CNI addresses are ignored.
// It filters for addresses with global scope, so that IPv6 link-local addresses are ignored.
func (c *Client) ParseActiveMachineAddresses(state *api.InstanceState) []string {
	if state == nil {
		return nil
	}
	var addresses []string
	for _, network := range state.Network {
		switch {
		case network.Type == "loopback":
			// ignore loopback
			continue
		case network.HostName == "":
			// only consider networks with a matching interface name on the host
			continue
		}

		for _, addr := range network.Addresses {
			switch {
			case addr.Scope != "global":
				// ignore addresses without global scope
				continue
			case addr.Family == "inet" && addr.Netmask == "32":
				// ignore /32 IPv4 addresses, this will most likely be a VIP
				continue
			case addr.Family == "inet6" && addr.Netmask == "128":
				// ignore /128 IPv6 addresses, this will most likely be a VIP
				continue
			}

			addresses = append(addresses, addr.Address)
		}
	}

	// sort to ensure stable order across invocations
	slices.Sort(addresses)
	return addresses
}
