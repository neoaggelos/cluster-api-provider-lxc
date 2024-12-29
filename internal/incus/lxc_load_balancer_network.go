package incus

import (
	"context"
	"fmt"
	"strings"

	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// loadBalancerNetwork is a LoadBalancerManager that spins up a network load-balancer.
// loadBalancerNetwork requires an OVN network.
type loadBalancerNetwork struct {
	lxcClient *Client

	clusterName      string
	clusterNamespace string

	networkName   string
	listenAddress string
}

// Create implements loadBalancerManager.
func (l *loadBalancerNetwork) Create(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerCreateTimeout)
	defer cancel()

	if !l.lxcClient.Client.HasExtension("network_load_balancer") {
		return nil, &terminalError{fmt.Errorf("server missing required 'network_load_balancer' extension, cannot create network load balancers")}
	}

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("networkName", l.networkName, "listenAddress", l.listenAddress))
	if _, _, err := l.lxcClient.Client.GetNetworkLoadBalancer(l.networkName, l.listenAddress); err != nil {
		if !strings.Contains(err.Error(), "Network Load Balancer not found") {
			return nil, fmt.Errorf("failed to GetNetworkLoadBalancer: %w", err)
		}
		return []string{l.listenAddress}, nil
	}

	log.FromContext(ctx).V(4).Info("Creating network load balancer")
	if err := l.lxcClient.Client.CreateNetworkLoadBalancer(l.networkName, api.NetworkLoadBalancersPost{
		ListenAddress: l.listenAddress,
		NetworkLoadBalancerPut: api.NetworkLoadBalancerPut{
			Config: map[string]string{
				configClusterNameKey:      l.clusterName,
				configClusterNamespaceKey: l.clusterNamespace,
				configInstanceRoleKey:     "loadbalancer",
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to CreateNetworkLoadBalancer: %w", err)
	}

	return []string{l.networkName}, nil
}

// Delete implements loadBalancerManager.
func (l *loadBalancerNetwork) Delete(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerDeleteTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("networkName", l.networkName, "listenAddress", l.listenAddress))

	log.FromContext(ctx).V(4).Info("Deleting network load balancer")
	if err := l.lxcClient.Client.DeleteNetworkLoadBalancer(l.networkName, l.listenAddress); err != nil && !strings.Contains(err.Error(), "Network load balancer not found") {
		return fmt.Errorf("failed to DeleteNetworkLoadBalancer: %w", err)
	}
	return nil
}

// Reconfigure implements loadBalancerManager.
func (l *loadBalancerNetwork) Reconfigure(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerReconfigureTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("networkName", l.networkName, "listenAddress", l.listenAddress))

	config, err := l.lxcClient.getLoadBalancerConfiguration(ctx, l.clusterName, l.clusterNamespace)
	if err != nil {
		return fmt.Errorf("failed to build load balancer configuration: %w", err)
	}

	log.FromContext(ctx).WithValues("servers", config.BackendServers).Info("Updating network load balancers")

	lbConfig := api.NetworkLoadBalancerPut{
		Config: map[string]string{
			configClusterNameKey:      l.clusterName,
			configClusterNamespaceKey: l.clusterNamespace,
			configInstanceRoleKey:     "loadbalancer",

			"healthcheck":               "true",
			"healthcheck.interval":      "5",
			"healthcheck.timeout":       "5",
			"healthcheck.failure_count": "3",
			"healthcheck.success_count": "2",
		},
		Backends: make([]api.NetworkLoadBalancerBackend, 0, len(config.BackendServers)),
		Ports: []api.NetworkLoadBalancerPort{{
			ListenPort:    config.FrontendControlPlanePort,
			Protocol:      "tcp",
			TargetBackend: make([]string, 0, len(config.BackendServers)),
		}},
	}
	for name, backend := range config.BackendServers {
		lbConfig.Backends = append(lbConfig.Backends, api.NetworkLoadBalancerBackend{
			Name:          name,
			TargetPort:    config.BackendControlPlanePort,
			TargetAddress: backend.Address,
		})

		lbConfig.Ports[0].TargetBackend = append(lbConfig.Ports[0].TargetBackend, name)
	}

	if err := l.lxcClient.Client.UpdateNetworkLoadBalancer(l.networkName, l.listenAddress, lbConfig, ""); err != nil {
		return fmt.Errorf("failed to UpdateNetworkLoadBalancer: %w", err)
	}

	return nil
}

var _ LoadBalancerManager = &loadBalancerNetwork{}
