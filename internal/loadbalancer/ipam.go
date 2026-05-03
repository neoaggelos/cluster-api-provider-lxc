package loadbalancer

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

type ipam struct {
	lxcClient *lxc.Client

	networkName string

	rangesKey   string
	volatileKey func(s string) string
}

func (a *ipam) Allocate(ctx context.Context, identity string) (raddress string, rerr error) {
	network, etag, err := a.lxcClient.GetNetwork(a.networkName)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve network %q: %w", a.networkName, err)
	}

	volatileIdentityKey := a.volatileKey(identity)

	defer func() {
		// if we have an address, patch network to ensure no one else allocates it
		// if the network has changed since we retrieved, then the etag value will be changed and this will fail
		if rerr == nil && len(raddress) > 0 {
			volatileAddrKey := a.volatileKey(raddress)

			if network.Config[volatileAddrKey] == identity && network.Config[volatileIdentityKey] == raddress {
				return
			}

			network.Config[volatileAddrKey] = identity
			network.Config[volatileIdentityKey] = raddress

			if err := a.lxcClient.UpdateNetwork(a.networkName, network.NetworkPut, etag); err != nil {
				rerr = fmt.Errorf("failed to allocate address %q on network %q: %w", raddress, a.networkName, err)
				raddress = ""
			} else if network, _, err := a.lxcClient.GetNetwork(a.networkName); err != nil {
				rerr = fmt.Errorf("failed to allocate address %q on network %q: failed to check network after update: %w", raddress, a.networkName, err)
				raddress = ""
			} else if network.Config[volatileAddrKey] != identity || network.Config[volatileIdentityKey] != raddress {
				rerr = fmt.Errorf("failed to allocate address %q on network %q: optimistic update failed", raddress, a.networkName)
				raddress = ""
			} else {
				log.FromContext(ctx).V(2).WithValues("address", raddress, "networkName", a.networkName).Info("Allocated new LoadBalancer address")
			}
		}
	}()

	// test if cluster already has an allocated IP address
	if addr, ok := network.Config[volatileIdentityKey]; ok { // already have address for cluster
		if v, ok := network.Config[a.volatileKey(addr)]; !ok || v == identity { // address is unallocated, or allocated to this cluster
			return addr, nil
		}
	}

	rangeString, ok := network.Config[a.rangesKey]
	if !ok {
		return "", fmt.Errorf("network %q does not have configuration %q", a.networkName, a.rangesKey)
	}

	iprange, err := utils.ParseIPRanges(rangeString)
	if err != nil {
		return "", fmt.Errorf("network %q has invalid %q configuration: %w", a.networkName, a.rangesKey, err)
	}

	// range over ip ranges to find a free IP
	for addr := range iprange.Iterate() {
		if v, ok := network.Config[a.volatileKey(addr)]; !ok || v == identity { // address is unallocated, or allocated to this cluster
			return addr, nil
		}
	}

	return "", fmt.Errorf("network %q range %q is fully allocated", a.networkName, rangeString)
}

// Release will free the address allocated for the cluster (if any).
func (a *ipam) Release(ctx context.Context, identity string) error {
	network, etag, err := a.lxcClient.GetNetwork(a.networkName)
	if err != nil {
		return fmt.Errorf("failed to retrieve network %q: %w", a.networkName, err)
	}

	volatileIdentityKey := a.volatileKey(identity)
	if address, ok := network.Config[volatileIdentityKey]; !ok {
		// cluster key does not match address, nothing to do
		return nil
	} else if volatileAddrKey := a.volatileKey(address); network.Config[volatileAddrKey] != identity {
		// address key does not match, nothing to do
		return nil
	} else {
		delete(network.Config, volatileAddrKey)
		delete(network.Config, volatileIdentityKey)

		if err := a.lxcClient.UpdateNetwork(a.networkName, network.NetworkPut, etag); err != nil {
			return fmt.Errorf("failed to release address %q from network %q: %w", address, a.networkName, err)
		} else if network, _, err := a.lxcClient.GetNetwork(a.networkName); err != nil {
			return fmt.Errorf("failed to release address %q from network %q: failed to check network after update: %w", address, a.networkName, err)
		} else if network.Config[volatileAddrKey] != "" || network.Config[volatileIdentityKey] != "" {
			return fmt.Errorf("failed to release address %q from network %q: optimistic update failed", address, a.networkName)
		} else {
			log.FromContext(ctx).V(2).WithValues("address", address, "networkName", a.networkName).Info("Released LoadBalancer address")
		}
	}

	return nil
}

func maybeAllocateAddressFromNetwork(ctx context.Context, address string, networkName string, lxcClient *lxc.Client, clusterName string, clusterNamespace string) (string, error) {
	if address != "" {
		return address, nil
	}

	if networkName == "" {
		return "", utils.TerminalError(fmt.Errorf("using external load balancer but none of .spec.controlPlaneEndpoint or .spec.loadBalancer.external.networkName are set"))
	}

	ipam := &ipam{
		lxcClient:   lxcClient,
		networkName: networkName,
		rangesKey:   "user.capn.vip.ranges",
		volatileKey: func(s string) string { return fmt.Sprintf("user.capn.vip.volatile.%s", s) },
	}

	address, err := ipam.Allocate(ctx, fmt.Sprintf("%s/%s", clusterName, clusterNamespace))
	if err != nil {
		return "", fmt.Errorf("allocator failed: %w", err)
	}

	return address, nil
}

func maybeReleaseAddressFromNetwork(ctx context.Context, networkName string, lxcClient *lxc.Client, clusterName string, clusterNamespace string) error {
	if networkName == "" {
		return nil
	}

	ipam := &ipam{
		lxcClient:   lxcClient,
		networkName: networkName,
		rangesKey:   "user.capn.vip.ranges",
		volatileKey: func(s string) string { return fmt.Sprintf("user.capn.vip.volatile.%s", s) },
	}

	if err := ipam.Release(ctx, fmt.Sprintf("%s/%s", clusterName, clusterNamespace)); err != nil {
		return fmt.Errorf("failed to release VIP: %w", err)
	}

	return nil
}
