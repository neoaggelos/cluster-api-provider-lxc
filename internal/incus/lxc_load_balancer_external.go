package incus

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// loadBalancerExternal is a no-op LoadBalancerManager when using an external LoadBalancer mechanism for the cluster (e.g. kube-vip).
type loadBalancerExternal struct {
	lxcClient *Client

	clusterName      string
	clusterNamespace string

	address string
}

// Create implements loadBalancerManager.
func (l *loadBalancerExternal) Create(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerCreateTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("address", l.address))

	if l.address == "" {
		return nil, fmt.Errorf("using external load balancer but no address is configured")
	}

	log.FromContext(ctx).V(4).Info("Using external load balancer")
	return []string{l.address}, nil
}

// Delete implements loadBalancerManager.
func (l *loadBalancerExternal) Delete(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerDeleteTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("address", l.address))

	log.FromContext(ctx).V(4).Info("Using external load balancer, nothing to delete")
	return nil
}

// Reconfigure implements loadBalancerManager.
func (l *loadBalancerExternal) Reconfigure(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, loadBalancerReconfigureTimeout)
	defer cancel()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("address", l.address))
	log.FromContext(ctx).V(4).Info("Using external load balancer, nothing to reconfigure")

	return nil
}

var _ LoadBalancerManager = &loadBalancerExternal{}
