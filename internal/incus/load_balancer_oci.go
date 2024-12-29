package incus

import (
	"bytes"
	"context"
	"fmt"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/loadbalancer"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// loadBalancerOCI is a LoadBalancerManager that spins up a haproxy OCI container.
type loadBalancerOCI struct {
	lxcClient *Client

	clusterName      string
	clusterNamespace string

	instanceName string
}

// Create implements loadBalancerManager.
func (l *loadBalancerOCI) Create(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerCreateTimeout)
	defer cancel()

	if !l.lxcClient.Client.HasExtension("instance_oci") {
		return nil, &terminalError{fmt.Errorf("server missing required 'instance_oci' extension, cannot create OCI container instance")}
	}

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", l.instanceName))
	if err := l.lxcClient.createInstanceIfNotExists(ctx, api.InstancesPost{
		Name: l.instanceName,
		Type: api.InstanceTypeContainer,
		Source: api.InstanceSource{
			Type:     "image",
			Protocol: "oci",
			Server:   loadBalancerDefaultHaproxyImageRegistry,
			Mode:     "pull",
			Alias:    loadBalancerDefaultHaproxyImage,
		},
		InstancePut: api.InstancePut{
			Config: map[string]string{
				configClusterNameKey:      l.clusterName,
				configClusterNamespaceKey: l.clusterNamespace,
				configInstanceRoleKey:     "loadbalancer",
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to ensure loadbalancer instance exists: %w", err)
	}

	if err := l.lxcClient.ensureInstanceRunning(ctx, l.instanceName); err != nil {
		return nil, fmt.Errorf("failed to ensure loadbalancer instance is running: %w", err)
	}

	addrs, err := l.lxcClient.waitForInstanceAddress(ctx, l.instanceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get loadbalancer instance address: %w", err)
	}
	return addrs, nil
}

// Delete implements loadBalancerManager.
func (l *loadBalancerOCI) Delete(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerDeleteTimeout)
	defer cancel()

	// name := lxcCluster.GetLoadBalancerInstanceName()
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", l.instanceName))

	return l.lxcClient.forceRemoveInstanceIfExists(ctx, l.instanceName)
}

// Reconfigure implements loadBalancerManager.
func (l *loadBalancerOCI) Reconfigure(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerReconfigureTimeout)
	defer cancel()

	// name := lxcCluster.GetLoadBalancerInstanceName()
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", l.instanceName))

	instances, err := l.lxcClient.getInstancesWithFilter(ctx, api.InstanceTypeAny, map[string]string{
		configClusterNameKey:      l.clusterName,
		configClusterNamespaceKey: l.clusterNamespace,
		configInstanceRoleKey:     "control-plane",
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster control plane instances: %w", err)
	}

	config := &loadbalancer.ConfigData{
		FrontendControlPlanePort: "6443",
		BackendControlPlanePort:  "6443",
		BackendServers:           make(map[string]loadbalancer.BackendServer, len(instances)),
	}
	for _, instance := range instances {
		if addresses := l.lxcClient.ParseActiveMachineAddresses(instance.State); len(addresses) > 0 {
			// TODO(neoaggelos): care about the instance weight (e.g. for deleted machines)
			// TODO(neoaggelos): care about ipv4 vs ipv6 addresses
			config.BackendServers[instance.Name] = loadbalancer.BackendServer{Address: addresses[0], Weight: 100}
		}
	}

	haproxyCfg, err := loadbalancer.Config(config, loadbalancer.DefaultTemplate)
	if err != nil {
		return fmt.Errorf("failed to render load balancer config: %w", err)
	}
	log.FromContext(ctx).V(4).WithValues("path", loadBalancerDefaultHaproxyConfigPath, "servers", config.BackendServers).Info("Write haproxy config")
	if err := l.lxcClient.Client.CreateInstanceFile(l.instanceName, loadBalancerDefaultHaproxyConfigPath, incus.InstanceFileArgs{
		Content:   bytes.NewReader(haproxyCfg),
		WriteMode: "overwrite",
		Type:      "file",
		Mode:      0440,
		UID:       0,
		GID:       0,
	}); err != nil {
		return fmt.Errorf("failed to write load balancer config to container: %w", err)
	}

	if err := l.lxcClient.killInstance(ctx, l.instanceName, "SIGUSR2"); err != nil {
		return fmt.Errorf("failed to send SIGUSR2 signal to reload configuration: %w", err)
	}

	return nil
}

var _ LoadBalancerManager = &loadBalancerOCI{}
