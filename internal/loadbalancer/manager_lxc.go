package loadbalancer

import (
	"bytes"
	"context"
	"fmt"

	incus "github.com/lxc/incus/v6/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/instances"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// managerLXC is a Manager that spins up an Ubuntu LXC container and installs haproxy from apt.
type managerLXC struct {
	lxcClient *lxc.Client

	clusterName      string
	clusterNamespace string

	name string
	spec infrav1.LXCLoadBalancerMachineSpec
}

// Create implements Manager.
func (l *managerLXC) Create(ctx context.Context) ([]string, error) {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("loadbalancer.instance", l.name))

	launchOpts := instances.HaproxyLXCLaunchOptions().
		WithProfiles(l.spec.Profiles).
		WithFlavor(l.spec.Flavor).
		WithConfig(map[string]string{
			"user.cluster-name":      l.clusterName,
			"user.cluster-namespace": l.clusterNamespace,
			"user.cluster-role":      "loadbalancer",
		}).
		WithImage(lxc.Image{
			Protocol:    l.spec.Image.Protocol,
			Server:      l.spec.Image.Server,
			Alias:       l.spec.Image.Name,
			Fingerprint: l.spec.Image.Fingerprint,
		})

	log.FromContext(ctx).V(1).Info("Launching load balancer instance")
	addrs, err := l.lxcClient.WithTarget(l.spec.Target).WaitForLaunchInstance(ctx, l.name, launchOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create load balancer instance: %w", err)
	}

	return addrs, nil
}

// Delete implements Manager.
func (l *managerLXC) Delete(ctx context.Context) error {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("loadbalancer.instance", l.name))

	log.FromContext(ctx).V(1).Info("Deleting load balancer instance")
	if err := l.lxcClient.WaitForDeleteInstance(ctx, l.name); err != nil {
		return fmt.Errorf("failed to delete load balancer instance: %w", err)
	}

	return nil
}

// Reconfigure implements Manager.
func (l *managerLXC) Reconfigure(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerReconfigureTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("loadbalancer.instance", l.name))

	config, err := getLoadBalancerConfiguration(ctx, l.lxcClient, filterClusterControlPlaneInstances(l.clusterName, l.clusterNamespace))
	if err != nil {
		return fmt.Errorf("failed to build load balancer configuration: %w", err)
	}

	haproxyCfg, err := renderHaproxyConfiguration(config, DefaultHaproxyTemplate)
	if err != nil {
		return fmt.Errorf("failed to render load balancer config: %w", err)
	}
	log.FromContext(ctx).V(1).WithValues("path", "/etc/haproxy/haproxy.cfg", "servers", config.BackendServers).Info("Write haproxy config")
	if err := l.lxcClient.CreateInstanceFile(l.name, "/etc/haproxy/haproxy.cfg", incus.InstanceFileArgs{
		Content:   bytes.NewReader(haproxyCfg),
		WriteMode: "overwrite",
		Type:      "file",
		Mode:      0440,
		UID:       0,
		GID:       0,
	}); err != nil {
		return fmt.Errorf("failed to write haproxy config: %w", err)
	}

	log.FromContext(ctx).V(1).Info("Reloading haproxy service")
	if err := l.lxcClient.RunCommand(ctx, l.name, []string{"systemctl", "reload", "haproxy.service"}, nil, nil, nil); err != nil {
		return fmt.Errorf("failed to reload haproxy service: %w", err)
	}

	return nil
}

func (l *managerLXC) Inspect(ctx context.Context) map[string]string {
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
		instance, _, err := l.lxcClient.GetInstanceFull(l.name)
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
		if err := l.lxcClient.RunCommand(ctx, l.name, item.command, nil, &stdout, &stderr); err != nil {
			result[fmt.Sprintf("%s.error", item.name)] = fmt.Errorf("failed to RunCommand %v on %s: %w", item.command, l.name, err).Error()
		}
		result[item.name] = fmt.Sprintf("%s\n%s\n", stdout.String(), stderr.String())
	}

	return result
}

func (l *managerLXC) ControlPlaneInstanceTemplates(controlPlaneInitialized bool) (map[string]string, error) {
	return nil, nil
}

var _ Manager = &managerLXC{}
