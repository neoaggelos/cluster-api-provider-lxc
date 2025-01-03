package controllers

import (
	"context"
	"fmt"

	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/profile"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/ptr"
)

func (r *LXCClusterReconciler) reconcileNormal(ctx context.Context, lxcCluster *infrav1.LXCCluster, lxcClient *incus.Client) error {
	// Detect if we are using LXD or Incus.
	if lxcCluster.Spec.ServerType == "" {
		if server, _, err := lxcClient.Client.GetServer(); err != nil {
			log.FromContext(ctx).Error(fmt.Errorf("failed to GetServer: %w", err), "Failed to retrieve server information")
		} else {
			switch server.Environment.Server {
			case "incus":
				lxcCluster.Spec.ServerType = "incus"
			case "lxd":
				lxcCluster.Spec.ServerType = "lxd"
			default:
				lxcCluster.Spec.ServerType = "unknown"
				log.FromContext(ctx).Error(fmt.Errorf("unknown server name %q", server.Environment.Server), "Failed to identify remote server type")
			}
		}
	}
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("serverType", lxcCluster.Spec.ServerType))

	// Create the default kubeadm profile for LXC containers
	profileName := lxcCluster.GetProfileName()
	if lxcCluster.Spec.SkipDefaultKubeadmProfile {
		conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, infrav1.KubeadmProfileDisabledReason, clusterv1.ConditionSeverityInfo, "Will not create default kubeadm profile %s", profileName)
	} else {

		ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("profileName", profileName))
		log.FromContext(ctx).Info("Creating default kubeadm profile")
		if err := lxcClient.InitProfile(ctx, api.ProfilesPost{Name: profileName, ProfilePut: profile.DefaultKubeadm}); err != nil {
			err = fmt.Errorf("failed to create default kubeadm profile %q: %w", profileName, err)
			log.FromContext(ctx).Error(err, "Failed to create default kubeadm profile")

			if incus.IsTerminalError(err) {
				conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, infrav1.KubeadmProfileCreationAbortedReason, clusterv1.ConditionSeverityError, "%s", err)
				lxcCluster.Status.FailureReason = ptr.To(infrav1.FailureReasonKubeadmProfileCreationFailed)
				lxcCluster.Status.FailureMessage = ptr.To(infrav1.FailureMessageKubeadmProfileCreationFailed)
				return nil
			}

			conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, infrav1.KubeadmProfileCreationFailedReason, clusterv1.ConditionSeverityWarning, "%s", err)
			return err
		}

		conditions.MarkTrue(lxcCluster, infrav1.KubeadmProfileAvailableCondition)
	}

	// Create the container hosting the load balancer.
	log.FromContext(ctx).Info("Creating load balancer")
	lbIPs, err := lxcClient.LoadBalancerManagerForCluster(lxcCluster).Create(ctx)
	if err != nil {
		err = fmt.Errorf("failed to create load balancer: %w", err)
		log.FromContext(ctx).Error(err, "Failed to provision cluster infrastructure")
		if incus.IsTerminalError(err) {
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
