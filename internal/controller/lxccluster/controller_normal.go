package lxccluster

import (
	"context"
	"fmt"

	"github.com/lxc/incus/v6/shared/api"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha2"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/profile"
)

func (r *LXCClusterReconciler) reconcileNormal(ctx context.Context, cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, lxcClient *incus.Client) error {
	// Create the default kubeadm profile for LXC containers
	profileName := lxcCluster.GetProfileName()
	if lxcCluster.Spec.SkipDefaultKubeadmProfile {
		conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, infrav1.KubeadmProfileDisabledReason, clusterv1.ConditionSeverityInfo, "Will not create default kubeadm profile %s", profileName)
	} else {
		ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("profileName", profileName, "privileged", !lxcCluster.Spec.Unprivileged))
		log.FromContext(ctx).Info("Creating default kubeadm profile")
		if err := lxcClient.InitProfile(ctx, api.ProfilesPost{Name: profileName, ProfilePut: profile.DefaultKubeadm(!lxcCluster.Spec.Unprivileged)}); err != nil {
			err = fmt.Errorf("failed to create default kubeadm profile %q: %w", profileName, err)
			log.FromContext(ctx).Error(err, "Failed to create default kubeadm profile")

			if incus.IsTerminalError(err) {
				conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, infrav1.KubeadmProfileCreationAbortedReason, clusterv1.ConditionSeverityError, "The default kubeadm LXC profile could not be created, most likely because of a permissions issue. Either enable privileged containers on the project, or specify .spec.skipDefaultKubeadmProfile=true on the LXCCluster object. The error was: %s", err)
				return nil
			}

			conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, infrav1.KubeadmProfileCreationFailedReason, clusterv1.ConditionSeverityWarning, "%s", err)
			return err
		}

		conditions.MarkTrue(lxcCluster, infrav1.KubeadmProfileAvailableCondition)
	}

	// Create the container hosting the load balancer.
	log.FromContext(ctx).Info("Creating load balancer")
	lbIPs, err := lxcClient.LoadBalancerManagerForCluster(cluster, lxcCluster).Create(ctx)
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to provision load balancer")
		if incus.IsTerminalError(err) {
			conditions.MarkFalse(lxcCluster, infrav1.LoadBalancerAvailableCondition, infrav1.LoadBalancerProvisioningAbortedReason, clusterv1.ConditionSeverityError, "The cluster load balancer could not be provisioned. The error was: %s", err)
			return nil
		}
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
