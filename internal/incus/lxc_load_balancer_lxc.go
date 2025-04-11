package incus

import (
	"bytes"
	"context"
	"fmt"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/loadbalancer"
)

// loadBalancerLXC is a LoadBalancerManager that spins up an Ubuntu LXC container and installs haproxy from apt.
type loadBalancerLXC struct {
	lxcClient *Client

	clusterName      string
	clusterNamespace string

	name string
	spec infrav1.LXCLoadBalancerMachineSpec
}

// Create implements loadBalancerManager.
func (l *loadBalancerLXC) Create(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerCreateTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", l.name))

	// If image is not set, use the default image (depending on the remote server type)
	image := l.spec.Image
	if image.IsZero() {
		image = infrav1.LXCMachineImageSource{
			Name:     "haproxy",
			Server:   defaultSimplestreamsServer,
			Protocol: "simplestreams",
		}
	}

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("image", image))

	if err := l.lxcClient.createInstanceIfNotExists(ctx, api.InstancesPost{
		Name:         l.name,
		Type:         api.InstanceTypeContainer,
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
func (l *loadBalancerLXC) Delete(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerDeleteTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", l.name))

	return l.lxcClient.forceRemoveInstanceIfExists(ctx, l.name)
}

// Reconfigure implements loadBalancerManager.
func (l *loadBalancerLXC) Reconfigure(ctx context.Context) error {
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
	log.FromContext(ctx).V(2).WithValues("path", "/etc/haproxy/haproxy.cfg", "servers", config.BackendServers).Info("Write haproxy config")
	if err := l.lxcClient.Client.CreateInstanceFile(l.name, "/etc/haproxy/haproxy.cfg", incus.InstanceFileArgs{
		Content:   bytes.NewReader(haproxyCfg),
		WriteMode: "overwrite",
		Type:      "file",
		Mode:      0440,
		UID:       0,
		GID:       0,
	}); err != nil {
		return fmt.Errorf("failed to write load balancer config to container: %w", err)
	}

	log.FromContext(ctx).V(2).Info("Reloading haproxy service")
	if err := l.lxcClient.wait(ctx, "ExecInstance", func() (incus.Operation, error) {
		return l.lxcClient.Client.ExecInstance(l.name, api.InstanceExecPost{
			Command: []string{"systemctl", "reload", "haproxy.service"},
		}, nil)
	}); err != nil {
		return fmt.Errorf("failed to reload haproxy service: %w", err)
	}

	return nil
}

func (l *loadBalancerLXC) Inspect(ctx context.Context) map[string]string {
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

	type logItem struct {
		name    string
		command []string
	}

	for _, item := range []logItem{
		{name: "ip-a.txt", command: []string{"ip", "a"}},
		{name: "ip-r.txt", command: []string{"ip", "r"}},
		{name: "ss-plnt.txt", command: []string{"ss", "-plnt"}},
		{name: "haproxy.service", command: []string{"systemctl", "status", "--no-pager", "-l", "haproxy.service"}},
		{name: "haproxy.log", command: []string{"journalctl", "--no-pager", "-u", "haproxy.service"}},
		{name: "haproxy.cfg", command: []string{"cat", "/etc/haproxy/haproxy.cfg"}},
	} {
		var stdout, stderr bytes.Buffer
		if err := l.lxcClient.RunCommand(ctx, l.name, item.command, &stdout, &stderr); err != nil {
			result[fmt.Sprintf("%s.error", item.name)] = fmt.Errorf("failed to RunCommand %v on %s: %w", item.command, l.name, err).Error()
		}
		result[item.name] = fmt.Sprintf("%s\n%s\n", stdout.String(), stderr.String())
	}

	return result
}

var _ LoadBalancerManager = &loadBalancerLXC{}
