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

// loadBalancerOCI is a LoadBalancerManager that spins up a haproxy OCI container.
type loadBalancerOCI struct {
	lxcClient *Client

	clusterName      string
	clusterNamespace string

	name string
	spec infrav1.LXCMachineSpec
}

// Create implements loadBalancerManager.
func (l *loadBalancerOCI) Create(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerCreateTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", l.name))

	if supports, err := l.lxcClient.serverSupportsExtension("instance_oci"); err != nil {
		return nil, fmt.Errorf("failed to check if server supports 'instance_oci' extension: %w", err)
	} else if !supports {
		return nil, terminalError{fmt.Errorf("server missing required 'instance_oci' extension, cannot create OCI container instances")}
	}

	source := api.InstanceSource{
		Type:     "image",
		Protocol: "oci",
		Server:   "https://docker.io",
		Mode:     "pull",
		Alias:    "kindest/haproxy:v20230606-42a2262b",
	}

	empty := infrav1.LXCMachineImageSource{}
	if l.spec.Image != empty {
		source = l.lxcClient.instanceSourceFromAPI(l.spec.Image)
	}

	if err := l.lxcClient.createInstanceIfNotExists(ctx, api.InstancesPost{
		Name:         l.name,
		Type:         api.InstanceTypeContainer, // instance type must be Container for OCI containers.
		Source:       source,
		InstanceType: l.spec.Flavor,
		InstancePut: api.InstancePut{
			Profiles: l.spec.Profiles,
			Config: map[string]string{
				configClusterNameKey:      l.clusterName,
				configClusterNamespaceKey: l.clusterNamespace,
				configInstanceRoleKey:     "loadbalancer",
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to ensure loadbalancer instance exists: %w", err)
	}

	if err := l.lxcClient.ensureInstanceRunning(ctx, l.name); err != nil {
		return nil, fmt.Errorf("failed to ensure loadbalancer instance is running: %w", err)
	}

	addrs, err := l.lxcClient.waitForInstanceAddress(ctx, l.name)
	if err != nil {
		return nil, fmt.Errorf("failed to get loadbalancer instance address: %w", err)
	}
	return addrs, nil
}

// Delete implements loadBalancerManager.
func (l *loadBalancerOCI) Delete(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerDeleteTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", l.name))

	return l.lxcClient.forceRemoveInstanceIfExists(ctx, l.name)
}

// Reconfigure implements loadBalancerManager.
func (l *loadBalancerOCI) Reconfigure(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerReconfigureTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", l.name))

	config, err := l.lxcClient.getLoadBalancerConfiguration(ctx, l.clusterName, l.clusterNamespace)
	if err != nil {
		return fmt.Errorf("failed to build load balancer configuration: %w", err)
	}

	haproxyCfg, err := loadbalancer.Config(config, loadbalancer.DefaultTemplate)
	if err != nil {
		return fmt.Errorf("failed to render load balancer config: %w", err)
	}
	log.FromContext(ctx).V(4).WithValues("path", "/usr/local/etc/haproxy/haproxy.cfg", "servers", config.BackendServers).Info("Write haproxy config")
	if err := l.lxcClient.Client.CreateInstanceFile(l.name, "/usr/local/etc/haproxy/haproxy.cfg", incus.InstanceFileArgs{
		Content:   bytes.NewReader(haproxyCfg),
		WriteMode: "overwrite",
		Type:      "file",
		Mode:      0440,
		UID:       0,
		GID:       0,
	}); err != nil {
		return fmt.Errorf("failed to write load balancer config to container: %w", err)
	}

	log.FromContext(ctx).V(4).WithValues("signal", "SIGUSR2").Info("Reloading haproxy configuration")
	if err := l.lxcClient.wait(ctx, "ExecInstance", func() (incus.Operation, error) {
		return l.lxcClient.Client.ExecInstance(l.name, api.InstanceExecPost{Command: []string{"kill", "1", "--signal", "SIGUSR2"}}, nil)
	}); err != nil {
		return fmt.Errorf("failed to send SIGUSR2 signal to reload configuration: %w", err)
	}

	return nil
}

var _ LoadBalancerManager = &loadBalancerOCI{}
