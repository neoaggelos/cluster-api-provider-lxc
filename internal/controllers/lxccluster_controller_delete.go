package controllers

import (
	"context"
	"fmt"

	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

func (r *LXCClusterReconciler) reconcileDelete(ctx context.Context, lxcCluster *infrav1.LXCCluster, lxcClient *incus.Client) error {
	// Set the LoadBalancerAvailableCondition reporting delete is started, and issue a patch in order to make
	// this visible to the users.
	// NB. The operation in LXC is fast, so there is the chance the user will not notice the status change;
	// nevertheless we are issuing a patch so we can test a pattern that will be used by other providers as well
	patchHelper, err := patch.NewHelper(lxcCluster, r.Client)
	if err != nil {
		return err
	}
	conditions.MarkFalse(lxcCluster, infrav1.LoadBalancerAvailableCondition, clusterv1.DeletingReason, clusterv1.ConditionSeverityInfo, "")
	if err := patchLXCCluster(ctx, patchHelper, lxcCluster); err != nil {
		return fmt.Errorf("failed to patch LXCCluster: %w", err)
	}

	// Delete the container hosting the load balancer
	if err := lxcClient.LoadBalancerManagerForCluster(lxcCluster).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete the load balancer instance: %w", err)
	}

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(lxcCluster, infrav1.ClusterFinalizer)

	return nil
}
