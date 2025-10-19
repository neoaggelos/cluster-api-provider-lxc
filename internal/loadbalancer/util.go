package loadbalancer

import (
	"context"
	"fmt"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

func filterClusterControlPlaneInstances(clusterName string, clusterNamespace string) lxc.ListInstanceFilter {
	return lxc.WithConfig(map[string]string{
		"user.cluster-name":      clusterName,
		"user.cluster-namespace": clusterNamespace,
		"user.cluster-role":      "control-plane",
	})
}

func getLoadBalancerConfiguration(ctx context.Context, lxcClient *lxc.Client, filters ...lxc.ListInstanceFilter) (*configData, error) {
	instances, err := lxcClient.ListInstances(ctx, filters...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cluster control plane instances: %w", err)
	}

	config := &configData{
		FrontendControlPlanePort: "6443",
		BackendControlPlanePort:  "6443",
		BackendServers:           make(map[string]backendServer, len(instances)),
	}
	for _, instance := range instances {
		if addresses := lxc.ParseHostAddresses(instance.State); len(addresses) > 0 {
			// TODO(neoaggelos): care about the instance weight (e.g. for deleted machines)
			// TODO(neoaggelos): care about ipv4 vs ipv6 addresses
			config.BackendServers[instance.Name] = backendServer{Address: addresses[0], Weight: 100}
		}
	}

	return config, nil
}

func GenerateHaproxyLoadBalancerConfiguration(ctx context.Context, lxcClient *lxc.Client, filters ...lxc.ListInstanceFilter) ([]byte, error) {
	config, err := getLoadBalancerConfiguration(ctx, lxcClient, filters...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve load balancer config: %w", err)
	}

	return renderHaproxyConfiguration(config, DefaultHaproxyTemplate)
}
