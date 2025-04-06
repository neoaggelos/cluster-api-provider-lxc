package incus

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha2"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/loadbalancer"
)

// loadBalancerOCI is a LoadBalancerManager that spins up a haproxy OCI container.
type loadBalancerOCI struct {
	lxcClient *Client

	clusterName      string
	clusterNamespace string

	name string
	spec infrav1.LXCLoadBalancerMachineSpec
}

// Create implements loadBalancerManager.
func (l *loadBalancerOCI) Create(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerCreateTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", l.name))

	if err := l.lxcClient.SupportsInstanceOCI(); err != nil {
		return nil, err
	}

	image := l.spec.Image
	if image.IsZero() {
		image = infrav1.LXCMachineImageSource{
			Name:     "neoaggelos/cluster-api-provider-lxc/haproxy:v0.0.1",
			Server:   "https://ghcr.io",
			Protocol: "oci",
		}
	}

	if err := l.lxcClient.createInstanceIfNotExists(ctx, api.InstancesPost{
		Name:         l.name,
		Type:         api.InstanceTypeContainer, // instance type must be Container for OCI containers.
		Source:       l.lxcClient.instanceSourceFromAPI(image),
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
	log.FromContext(ctx).V(2).WithValues("path", "/usr/local/etc/haproxy/haproxy.cfg", "servers", config.BackendServers).Info("Write haproxy config")
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

	if err := l.lxcClient.ensureInstanceRunning(ctx, l.name); err != nil {
		return fmt.Errorf("failed to ensure load balancer is running: %w", err)
	}

	// NOTE(neoaggelos): lxc will silence signals to the init process from the same namespace
	// https://github.com/lxc/lxc/pull/4503/files#diff-bf8397458e8edecf47bdd0021167704c859cdd088fd7afdf0b56d289cf87a54fR430-R434
	//
	// Instead, scan `/proc` for pids and send SIGUSR2 to all of them
	log.FromContext(ctx).V(2).Info("Reloading haproxy configuration")
	if err := l.lxcClient.wait(ctx, "ExecInstance", func() (incus.Operation, error) {

		command := []string{"kill", "--signal", "SIGUSR2"}
		if _, response, err := l.lxcClient.Client.GetInstanceFile(l.name, "/proc"); err != nil {
			return nil, fmt.Errorf("failed to list running processes in load balancer instance: %w", err)
		} else {
			// find numeric pids under /proc
			for _, entry := range response.Entries {
				if _, err := strconv.ParseUint(entry, 10, 64); err != nil {
					continue
				}
				command = append(command, entry)
			}
		}

		log.FromContext(ctx).V(4).WithValues("command", command).Info("Kill haproxy processes")
		return l.lxcClient.Client.ExecInstance(l.name, api.InstanceExecPost{Command: command}, nil)
	}); err != nil {
		return fmt.Errorf("failed to send SIGUSR2 signal to reload configuration: %w", err)
	}

	return nil
}

// Inspect implements loadBalancerManager.
func (l *loadBalancerOCI) Inspect(ctx context.Context) map[string]string {
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

	addInfoFor("Instance", func() (any, error) {
		instance, _, err := l.lxcClient.Client.GetInstanceFull(l.name)
		return instance, err
	})

	reader, _, err := l.lxcClient.Client.GetInstanceFile(l.name, "/usr/local/etc/haproxy/haproxy.cfg")
	if err != nil || reader == nil {
		result["haproxy.cfg"] = fmt.Errorf("failed to GetInstanceFile: %w", err).Error()
	} else {
		defer func() { _ = reader.Close() }()
		if b, err := io.ReadAll(reader); err != nil {
			result["haproxy.cfg"] = fmt.Errorf("failed to read haproxy.cfg: %w", err).Error()
		} else {
			result["haproxy.cfg"] = string(b)
		}
	}

	return result
}

var _ LoadBalancerManager = &loadBalancerOCI{}
