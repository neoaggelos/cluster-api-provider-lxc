package incus

import (
	"bytes"
	"context"
	"fmt"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/loadbalancer"
)

// ReconfigureLoadBalancer updates the haproxy configuration based on the list of running control plane instances
// and reloads haproxy.
func (c *Client) ReconfigureLoadBalancer(ctx context.Context, lxcCluster *infrav1.LXCCluster) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerReconfigureTimeout)
	defer cancel()

	name := lxcCluster.GetLoadBalancerInstanceName()
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", name))

	instances, err := c.getInstancesWithFilter(ctx, api.InstanceTypeAny, map[string]string{
		configClusterNameKey:      lxcCluster.Name,
		configClusterNamespaceKey: lxcCluster.Namespace,
		configInstanceRoleKey:     "control-plane",
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster control plane instances: %w", err)
	}

	config := &loadbalancer.ConfigData{
		FrontendControlPlanePort: "6443",
		BackendControlPlanePort:  "80",
		BackendServers:           make(map[string]loadbalancer.BackendServer, len(instances)),
	}
	for _, instance := range instances {
		if address := c.GetAddressIfExists(instance.State); address != "" {
			// TODO(neoaggelos): care about the instance weight (e.g. for deleted machines)
			config.BackendServers[instance.Name] = loadbalancer.BackendServer{Address: address, Weight: 100}
		}
	}

	haproxyCfg, err := loadbalancer.Config(config, loadbalancer.DefaultTemplate)
	if err != nil {
		return fmt.Errorf("failed to render load balancer config: %w", err)
	}
	log.FromContext(ctx).V(4).WithValues("path", loadBalancerDefaultHaproxyConfigPath, "servers", config.BackendServers).Info("Write haproxy config")
	if err := c.Client.CreateInstanceFile(name, loadBalancerDefaultHaproxyConfigPath, incus.InstanceFileArgs{
		Content:   bytes.NewReader(haproxyCfg),
		WriteMode: "overwrite",
		Type:      "file",
		Mode:      0440,
		UID:       0,
		GID:       0,
	}); err != nil {
		return fmt.Errorf("failed to write load balancer config to container: %w", err)
	}

	if err := c.killInstance(ctx, name, "SIGUSR2"); err != nil {
		return fmt.Errorf("failed to send SIGUSR2 signal to reload configuration: %w", err)
	}

	return nil
}
