package loadbalancer

import (
	"bytes"
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

// managerKubeVIP is a Manager that injects KubeVIP configuration into control plane instances.
type managerKubeVIP struct {
	lxcClient *lxc.Client

	clusterName      string
	clusterNamespace string

	address     string
	networkName string

	interfaceName  string
	kubeconfigPath string
	manifestPath   string
	image          string
}

// Create implements Manager.
func (l *managerKubeVIP) Create(ctx context.Context) ([]string, error) {
	var err error
	if l.address, err = maybeAllocateAddressFromNetwork(ctx, l.address, l.networkName, l.lxcClient, l.clusterName, l.clusterNamespace); err != nil {
		return nil, fmt.Errorf("failed to allocate load balancer address: %w", err)
	}

	log.FromContext(ctx).V(1).Info("Using KubeVIP load balancer, nothing to create")
	return []string{l.address}, nil
}

// Delete implements Manager.
func (l *managerKubeVIP) Delete(ctx context.Context) error {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("address", l.address))

	if err := maybeReleaseAddressFromNetwork(ctx, l.networkName, l.lxcClient, l.clusterName, l.clusterNamespace); err != nil {
		return fmt.Errorf("failed to release load balancer address: %w", err)
	}

	log.FromContext(ctx).V(1).Info("Using KubeVIP load balancer, nothing to delete")
	return nil
}

// Reconfigure implements Manager.
func (l *managerKubeVIP) Reconfigure(ctx context.Context) error {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("address", l.address))
	log.FromContext(ctx).V(1).Info("Using KubeVIP load balancer, nothing to reconfigure")

	return nil
}

// Inspect implements Manager.
func (l *managerKubeVIP) Inspect(ctx context.Context) map[string]string {
	result := map[string]string{}

	instances, err := l.lxcClient.ListInstances(ctx, lxc.WithConfig(map[string]string{
		"user.cluster-namespace": l.clusterNamespace,
		"user.cluster-name":      l.clusterName,
		"user.cluster-role":      "control-plane",
	}))
	if err != nil {
		return map[string]string{"instances.err": err.Error()}
	}

	type logItem struct {
		name    string
		command []string
	}

	for _, instance := range instances {
		for _, item := range []logItem{
			{name: fmt.Sprintf("%s/ip-a.txt", instance.Name), command: []string{"ip", "a"}},
			{name: fmt.Sprintf("%s/ip-r.txt", instance.Name), command: []string{"ip", "r"}},
			{name: fmt.Sprintf("%s/kube-vip.yaml", instance.Name), command: []string{"cat", l.getManifestPath()}},
		} {
			var stdout, stderr bytes.Buffer
			if err := l.lxcClient.RunCommand(ctx, instance.Name, item.command, nil, &stdout, &stderr); err != nil {
				result[fmt.Sprintf("%s.error", item.name)] = fmt.Errorf("failed to RunCommand %v on %s: %w", item.command, instance.Name, err).Error()
			}
			result[item.name] = fmt.Sprintf("%s\n%s\n", stdout.String(), stderr.String())
		}
	}

	return result
}

func (l *managerKubeVIP) getManifestPath() string {
	if len(l.manifestPath) > 0 {
		return l.manifestPath
	}
	return "/etc/kubernetes/manifests/kube-vip.yaml"
}

func (l *managerKubeVIP) getImage() string {
	if len(l.image) > 0 {
		return l.image
	}
	return "ghcr.io/kube-vip/kube-vip:v0.6.4"
}

func (l *managerKubeVIP) getKubeconfigPath(controlPlaneInitialized bool) string {
	if len(l.kubeconfigPath) != 0 {
		return l.kubeconfigPath
	}
	if !controlPlaneInitialized {
		return "/etc/kubernetes/super-admin.conf"
	}

	return "/etc/kubernetes/admin.conf"
}

func (l *managerKubeVIP) ControlPlaneInstanceTemplates(controlPlaneInitialized bool) (map[string]string, error) {
	if b, err := renderKubeVIPConfiguration(kubeVIPTemplateInput{
		Address:        l.address,
		Interface:      l.interfaceName,
		Image:          l.getImage(),
		KubeconfigPath: l.getKubeconfigPath(controlPlaneInitialized),
	}); err != nil {
		return nil, utils.TerminalError(fmt.Errorf("failed to generate KubeVIP config file: %w", err))
	} else {
		return map[string]string{l.getManifestPath(): string(b)}, nil
	}
}

var _ Manager = &managerKubeVIP{}
