package controllers

import (
	"context"
	"fmt"

	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
	"sigs.k8s.io/cluster-api/util/conditions"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

func (r *LXCClusterReconciler) reconcileNormal(ctx context.Context, lxcCluster *infrav1.LXCCluster, lxcClient *incus.Client) error {
	// Create the container hosting the load balancer.
	lbIPs, err := lxcClient.LoadBalancerManagerForCluster(lxcCluster).Create(ctx)
	if err != nil {
		err = fmt.Errorf("failed to create load balancer: %w", err)
		conditions.MarkFalse(lxcCluster, infrav1.LoadBalancerAvailableCondition, infrav1.LoadBalancerProvisioningFailedReason, clusterv1.ConditionSeverityWarning, "%s", err)
		return err
	}

	// Surface the control plane endpoint
	if lxcCluster.Spec.ControlPlaneEndpoint.Host == "" {
		// TODO(neoaggelos): care about IPv4 vs IPv6
		lxcCluster.Spec.ControlPlaneEndpoint.Host = lbIPs[0]
	}
	if lxcCluster.Spec.ControlPlaneEndpoint.Port == 0 {
		lxcCluster.Spec.ControlPlaneEndpoint.Port = 6443
	}

	// Mark the lxcCluster ready
	lxcCluster.Status.Ready = true
	conditions.MarkTrue(lxcCluster, infrav1.LoadBalancerAvailableCondition)

	return nil
}
