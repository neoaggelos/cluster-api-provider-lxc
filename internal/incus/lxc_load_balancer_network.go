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

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("networkName", l.networkName, "listenAddress", l.listenAddress))

	if l.networkName == "" {
		return nil, terminalError{fmt.Errorf("network load balancer cannot be provisioned as .spec.loadBalancer.ovnNetworkName is not specified")}
	}

	if unsupported, err := l.lxcClient.serverSupportsExtensions("network_load_balancer", "network_load_balancer_health_check"); err != nil {
		return nil, fmt.Errorf("failed to check if server supports network load balancer extensions: %w", err)
	} else if len(unsupported) > 0 {
		return nil, terminalError{fmt.Errorf("server cannot create network load balancers, required extensions are missing: %v", unsupported)}
	}

	if _, _, err := l.lxcClient.Client.GetNetwork(l.networkName); err != nil {
		return nil, terminalError{fmt.Errorf("failed to check network %q: %w", l.networkName, err)}
	}
	if lb, _, err := l.lxcClient.Client.GetNetworkLoadBalancer(l.networkName, l.listenAddress); err != nil && !strings.Contains(err.Error(), "Network load balancer not found") {
		return nil, fmt.Errorf("failed to GetNetworkLoadBalancer: %w", err)
	} else if err == nil {
		if lb.Config[configClusterNameKey] != l.clusterName || lb.Config[configClusterNamespaceKey] != l.clusterNamespace {
			return nil, terminalError{fmt.Errorf("conflict: a LoadBalancer with IP %s already exists without the required keys %s=%s and %s=%s", l.listenAddress, configClusterNameKey, l.clusterName, configClusterNamespaceKey, l.clusterNamespace)}
		}
		return []string{l.listenAddress}, nil
	}

	log.FromContext(ctx).V(2).Info("Creating network load balancer")
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

	return []string{l.listenAddress}, nil
}

// Delete implements loadBalancerManager.
func (l *loadBalancerNetwork) Delete(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerDeleteTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("networkName", l.networkName, "listenAddress", l.listenAddress))

	log.FromContext(ctx).V(2).Info("Deleting network load balancer")
	if err := l.lxcClient.Client.DeleteNetworkLoadBalancer(l.networkName, l.listenAddress); err != nil && !strings.Contains(err.Error(), "not found") {
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
