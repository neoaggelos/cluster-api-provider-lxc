package loadbalancer

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// managerExternal is a no-op Manager when using an external LoadBalancer mechanism for the cluster (e.g. kube-vip).
type managerExternal struct {
	lxcClient *lxc.Client

	clusterName      string
	clusterNamespace string

	address     string
	networkName string
}

// Create implements Manager.
func (l *managerExternal) Create(ctx context.Context) ([]string, error) {
	var err error
	if l.address, err = maybeAllocateAddressFromNetwork(ctx, l.address, l.networkName, l.lxcClient, l.clusterName, l.clusterNamespace); err != nil {
		return nil, fmt.Errorf("failed to allocate load balancer address: %w", err)
	}

	log.FromContext(ctx).V(1).WithValues("address", l.address).Info("Using external load balancer")
	return []string{l.address}, nil
}

// Delete implements Manager.
func (l *managerExternal) Delete(ctx context.Context) error {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("address", l.address))

	if err := maybeReleaseAddressFromNetwork(ctx, l.networkName, l.lxcClient, l.clusterName, l.clusterNamespace); err != nil {
		return fmt.Errorf("failed to release load balancer address: %w", err)
	}

	log.FromContext(ctx).V(1).Info("Using external load balancer, nothing to delete")
	return nil
}

// Reconfigure implements Manager.
func (l *managerExternal) Reconfigure(ctx context.Context) error {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("address", l.address))
	log.FromContext(ctx).V(1).Info("Using external load balancer, nothing to reconfigure")

	return nil
}

// Inspect implements Manager.
func (l *managerExternal) Inspect(ctx context.Context) map[string]string {
	return map[string]string{"address": l.address}
}

// ControlPlaneInstanceTemplates implements Manager.
func (l *managerExternal) ControlPlaneInstanceTemplates(controlPlaneInitialized bool) (map[string]string, error) {
	return nil, nil
}

var _ Manager = &managerExternal{}
