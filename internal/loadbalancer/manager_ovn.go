package loadbalancer

import (
	"context"
	"fmt"
	"strings"

	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

// managerOVN is a Manager that spins up a network load-balancer.
// managerOVN requires an OVN network.
type managerOVN struct {
	lxcClient *lxc.Client

	clusterName      string
	clusterNamespace string

	networkName   string
	listenAddress string
}

// Create implements Manager.
func (l *managerOVN) Create(ctx context.Context) ([]string, error) {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("networkName", l.networkName, "listenAddress", l.listenAddress))

	if l.networkName == "" {
		return nil, utils.TerminalError(fmt.Errorf("network load balancer cannot be provisioned as .spec.loadBalancer.ovn.networkName is not specified"))
	}

	if err := l.lxcClient.SupportsNetworkLoadBalancers(); err != nil {
		return nil, fmt.Errorf("server does not support network load balancers: %w", err)
	}

	if _, _, err := l.lxcClient.GetNetwork(l.networkName); err != nil {
		return nil, utils.TerminalError(fmt.Errorf("failed to check network %q: %w", l.networkName, err))
	}
	if lb, _, err := l.lxcClient.GetNetworkLoadBalancer(l.networkName, l.listenAddress); err != nil && !strings.Contains(err.Error(), "Network load balancer not found") {
		return nil, fmt.Errorf("failed to GetNetworkLoadBalancer: %w", err)
	} else if err == nil {
		if lb.Config["user.cluster-name"] != l.clusterName || lb.Config["user.cluster-namespace"] != l.clusterNamespace {
			return nil, utils.TerminalError(fmt.Errorf("conflict: a LoadBalancer with IP %s already exists without the required keys %s=%s and %s=%s", l.listenAddress, "user.cluster-name", l.clusterName, "user.cluster-namespace", l.clusterNamespace))
		}
		log.FromContext(ctx).V(1).Info("Network load balancer already exists")
		return []string{l.listenAddress}, nil
	}

	log.FromContext(ctx).V(1).Info("Creating network load balancer")
	if err := l.lxcClient.CreateNetworkLoadBalancer(l.networkName, api.NetworkLoadBalancersPost{
		ListenAddress: l.listenAddress,
		NetworkLoadBalancerPut: api.NetworkLoadBalancerPut{
			Config: map[string]string{
				"user.cluster-name":      l.clusterName,
				"user.cluster-namespace": l.clusterNamespace,
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to CreateNetworkLoadBalancer: %w", err)
	}

	return []string{l.listenAddress}, nil
}

// Delete implements Manager.
func (l *managerOVN) Delete(ctx context.Context) error {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("networkName", l.networkName, "listenAddress", l.listenAddress))

	log.FromContext(ctx).V(1).Info("Deleting network load balancer")
	if err := l.lxcClient.DeleteNetworkLoadBalancer(l.networkName, l.listenAddress); err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to DeleteNetworkLoadBalancer: %w", err)
	}
	return nil
}

// Reconfigure implements Manager.
func (l *managerOVN) Reconfigure(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerReconfigureTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("networkName", l.networkName, "listenAddress", l.listenAddress))

	config, err := getLoadBalancerConfiguration(ctx, l.lxcClient, filterClusterControlPlaneInstances(l.clusterName, l.clusterNamespace))
	if err != nil {
		return fmt.Errorf("failed to build load balancer configuration: %w", err)
	}

	log.FromContext(ctx).V(1).WithValues("servers", config.BackendServers).Info("Updating network load balancer")

	lbConfig := api.NetworkLoadBalancerPut{
		Config: map[string]string{
			"user.cluster-name":      l.clusterName,
			"user.cluster-namespace": l.clusterNamespace,

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

	if err := l.lxcClient.UpdateNetworkLoadBalancer(l.networkName, l.listenAddress, lbConfig, ""); err != nil {
		return fmt.Errorf("failed to UpdateNetworkLoadBalancer: %w", err)
	}

	return nil
}

// Inspect implements Manager.
func (l *managerOVN) Inspect(ctx context.Context) map[string]string {
	result := map[string]string{}

	addInfoFor := func(name string, getter func() (any, error)) {
		if obj, err := getter(); err != nil {
			result[fmt.Sprintf("%s.err", name)] = fmt.Errorf("failed to get %s: %w", name, err).Error()
		} else {
			result[fmt.Sprintf("%s.txt", name)] = fmt.Sprintf("%#v\n", obj)
			b, err := yaml.Marshal(obj)
			if err != nil {
				result[fmt.Sprintf("%s.err", name)] = fmt.Errorf("failed to marshal yaml: %w", err).Error()
			} else {
				result[fmt.Sprintf("%s.yaml", name)] = string(b)
			}
		}
	}

	var uplinkNetwork string
	addInfoFor("Network", func() (any, error) {
		network, _, err := l.lxcClient.GetNetwork(l.networkName)
		uplinkNetwork = network.Config["network"]
		return network, err
	})
	addInfoFor("UplinkNetwork", func() (any, error) {
		network, _, err := l.lxcClient.GetNetwork(uplinkNetwork)
		return network, err
	})
	addInfoFor("NetworkLoadBalancer", func() (any, error) {
		lb, _, err := l.lxcClient.GetNetworkLoadBalancer(l.networkName, l.listenAddress)
		return lb, err
	})
	addInfoFor("NetworkLoadBalancerState", func() (any, error) {
		return l.lxcClient.GetNetworkLoadBalancerState(l.networkName, l.listenAddress)
	})

	return result
}

func (l *managerOVN) ControlPlaneInstanceTemplates(controlPlaneInitialized bool) (map[string]string, error) {
	return nil, nil
}

var _ Manager = &managerOVN{}
