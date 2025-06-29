package loadbalancer

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

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

func defaultHaproxyOCIImage() api.InstanceSource {
	return api.InstanceSource{
		Type:     "image",
		Protocol: "oci",
		Server:   "https://ghcr.io",
		Alias:    "lxc/cluster-api-provider-incus/haproxy:v20230606-42a2262b",
	}
}

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

	// Use default haproxy image if not set
	var image api.InstanceSource
	if l.spec.Image.IsZero() {
		image = defaultHaproxyOCIImage()
	} else {
		image = api.InstanceSource{
			Type:        "image",
			Protocol:    l.spec.Image.Protocol,
			Server:      l.spec.Image.Server,
			Alias:       l.spec.Image.Name,
			Fingerprint: l.spec.Image.Fingerprint,
		}
	}

	log.FromContext(ctx).V(1).Info("Launching load balancer instance")
	addrs, err := l.lxcClient.WithTarget(l.spec.Target).WaitForLaunchInstance(ctx, api.InstancesPost{
		Name:         l.name,
		Type:         api.InstanceTypeContainer,
		Source:       image,
		InstanceType: l.spec.Flavor,
		InstancePut: api.InstancePut{
			Profiles: l.spec.Profiles,
			Config: map[string]string{
				"user.cluster-name":      l.clusterName,
				"user.cluster-namespace": l.clusterNamespace,
				"user.cluster-role":      "loadbalancer",
			},
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create load balancer instance: %w", err)
	}

	return addrs, nil
}

// Delete implements loadBalancerManager.
func (l *managerOCI) Delete(ctx context.Context) error {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("loadbalancer.instance", l.name))

	log.FromContext(ctx).V(1).Info("Deleting load balancer instance")
	if err := l.lxcClient.WaitForDeleteInstance(ctx, l.name); err != nil {
		return fmt.Errorf("failed to delete load balancer instance: %w", err)
	}

	return nil
}

// Reconfigure implements loadBalancerManager.
func (l *managerOCI) Reconfigure(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerReconfigureTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("loadbalancer.instance", l.name))

	config, err := getLoadBalancerConfiguration(ctx, l.lxcClient, l.clusterName, l.clusterNamespace)
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
	var haproxyPids []string
	if _, response, err := l.lxcClient.GetInstanceFile(l.name, "/proc"); err != nil {
		return fmt.Errorf("failed to list running processes in load balancer instance: %w", err)
	} else {
		for _, entry := range response.Entries {
			if _, err := strconv.ParseUint(entry, 10, 64); err != nil {
				continue
			}
			haproxyPids = append(haproxyPids, entry)
		}
	}

	if err := l.lxcClient.RunCommand(ctx, l.name, append([]string{"kill", "--signal", "SIGUSR2"}, haproxyPids...), nil, nil, nil); err != nil {
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

var _ Manager = &managerOCI{}
