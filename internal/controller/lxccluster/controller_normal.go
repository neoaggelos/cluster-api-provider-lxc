package lxccluster

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/loadbalancer"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

func (r *LXCClusterReconciler) reconcileNormal(ctx context.Context, cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, lxcClient *lxc.Client) error {
	// Create the container hosting the load balancer.
	log.FromContext(ctx).Info("Creating load balancer")
	lbIPs, err := loadbalancer.ManagerForCluster(cluster, lxcCluster, lxcClient).Create(ctx)
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to provision load balancer")
		if utils.IsTerminalError(err) {
			conditions.Set(lxcCluster, metav1.Condition{Type: infrav1.LoadBalancerAvailableCondition, Status: metav1.ConditionFalse, Reason: infrav1.LoadBalancerProvisioningAbortedReason, Message: fmt.Sprintf("Fatal error provisioning the cluster load balancer: %v", err)})
			return nil
		}
		conditions.Set(lxcCluster, metav1.Condition{Type: infrav1.LoadBalancerAvailableCondition, Status: metav1.ConditionFalse, Reason: infrav1.LoadBalancerProvisioningFailedReason, Message: err.Error()})
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
	lxcCluster.Status.Initialization.Provisioned = ptr.To(true)
	conditions.Set(lxcCluster, metav1.Condition{Type: infrav1.LoadBalancerAvailableCondition, Status: metav1.ConditionTrue, Reason: infrav1.LoadBalancerProvisionedReason})

	return nil
}
