package controllers

import (
	"context"
	"fmt"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
)

func (r *LXCClusterReconciler) reconcileNormal(ctx context.Context, lxcCluster *infrav1.LXCCluster, lxcClient *incus.Client) error {
	// Create the container hosting the load balancer.
	lbIPs, err := lxcClient.LoadBalancerManagerForCluster(lxcCluster).Create(ctx)
	if err != nil {
		err = fmt.Errorf("failed to create load balancer: %w", err)
		if incus.IsTerminalError(err) {
			log.FromContext(ctx).Error(err, "Failed to provision cluster infrastructure")
			conditions.MarkFalse(lxcCluster, infrav1.LoadBalancerAvailableCondition, infrav1.LoadBalancerProvisioningAbortedReason, clusterv1.ConditionSeverityError, "%s", err)
			lxcCluster.Status.FailureReason = ptr.To(infrav1.FailureReasonLoadBalancerProvisionFailed)
			lxcCluster.Status.FailureMessage = ptr.To(infrav1.FailureMessageLoadBalancerProvisionFailed)
			return nil
		}
		conditions.MarkFalse(lxcCluster, infrav1.LoadBalancerAvailableCondition, infrav1.LoadBalancerProvisioningFailedReason, clusterv1.ConditionSeverityWarning, "%s", err)
		return err
	}

	lxcCluster.Status.FailureReason = nil
	lxcCluster.Status.FailureMessage = nil

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
