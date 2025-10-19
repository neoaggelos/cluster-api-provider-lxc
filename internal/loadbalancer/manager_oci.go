package loadbalancer

import (
	"bytes"
	"context"
	"fmt"
	"io"

	incus "github.com/lxc/incus/v6/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/instances"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// managerOCI is a Manager that spins up a kindest/haproxy OCI container.
type managerOCI struct {
	lxcClient *lxc.Client

	clusterName      string
	clusterNamespace string

	name string
	spec infrav1.LXCLoadBalancerMachineSpec
}

// Create implements Manager.
func (l *managerOCI) Create(ctx context.Context) ([]string, error) {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("loadbalancer.instance", l.name))

	if err := l.lxcClient.SupportsInstanceOCI(); err != nil {
		return nil, fmt.Errorf("server does not support OCI containers: %w", err)
	}

	launchOpts := instances.HaproxyOCILaunchOptions().
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
func (l *managerOCI) Delete(ctx context.Context) error {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("loadbalancer.instance", l.name))

	log.FromContext(ctx).V(1).Info("Deleting load balancer instance")
	if err := l.lxcClient.WaitForDeleteInstance(ctx, l.name); err != nil {
		return fmt.Errorf("failed to delete load balancer instance: %w", err)
	}

	return nil
}

// Reconfigure implements Manager.
func (l *managerOCI) Reconfigure(ctx context.Context) error {
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
	log.FromContext(ctx).V(1).WithValues("path", "/usr/local/etc/haproxy/haproxy.cfg", "servers", config.BackendServers).Info("Write haproxy config")
	if err := l.lxcClient.CreateInstanceFile(l.name, "/usr/local/etc/haproxy/haproxy.cfg", incus.InstanceFileArgs{
		Content:   bytes.NewReader(haproxyCfg),
		WriteMode: "overwrite",
		Type:      "file",
		Mode:      0440,
		UID:       0,
		GID:       0,
	}); err != nil {
		return fmt.Errorf("failed to write haproxy config: %w", err)
	}

	log.FromContext(ctx).V(1).Info("Reloading haproxy configuration")
	if err := l.lxcClient.RunCommand(ctx, l.name, append([]string{"kill", "--signal", "SIGUSR2"}, "1"), nil, nil, nil); err != nil {
		return fmt.Errorf("failed to send SIGUSR2 to haproxy pids: %w", err)
	}

	return nil
}

func (l *managerOCI) Inspect(ctx context.Context) map[string]string {
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

	reader, _, err := l.lxcClient.GetInstanceFile(l.name, "/usr/local/etc/haproxy/haproxy.cfg")
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

func (l *managerOCI) ControlPlaneInstanceTemplates(controlPlaneInitialized bool) (map[string]string, error) {
	return nil, nil
}

var _ Manager = &managerOCI{}
