package lxccluster

import (
	"context"
	"fmt"
	"time"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/incus"
	"github.com/lxc/cluster-api-provider-incus/internal/util"
)

func (r *LXCClusterReconciler) reconcileDelete(ctx context.Context, cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, lxcClient *incus.Client) (ctrl.Result, error) {
	// Set the LoadBalancerAvailableCondition reporting delete is started, and issue a patch in order to make
	// this visible to the users.
	// NB. The operation in LXC is fast, so there is the chance the user will not notice the status change;
	// nevertheless we are issuing a patch so we can test a pattern that will be used by other providers as well
	patchHelper, err := patch.NewHelper(lxcCluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	conditions.MarkFalse(lxcCluster, infrav1.LoadBalancerAvailableCondition, clusterv1.DeletingReason, clusterv1.ConditionSeverityInfo, "")
	conditions.MarkFalse(lxcCluster, infrav1.KubeadmProfileAvailableCondition, clusterv1.DeletingReason, clusterv1.ConditionSeverityInfo, "")
	if err := patchLXCCluster(ctx, patchHelper, lxcCluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to patch LXCCluster: %w", err)
	}

	// Delete the container hosting the load balancer
	log.FromContext(ctx).Info("Deleting load balancer")
	if err := lxcClient.LoadBalancerManagerForCluster(cluster, lxcCluster).Delete(ctx); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete the load balancer instance: %w", err)
	}

	machines, err := util.GetMachinesForCluster(ctx, r.Client, client.ObjectKeyFromObject(cluster))
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to get list of Machines for Cluster")
	} else if len(machines) > 0 {
		log.FromContext(ctx).WithValues("machines", len(machines)).Info("Waiting for all Machines to be deleted")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	log.FromContext(ctx).Info("Deleting default kubeadm profile")
	if err := lxcClient.DeleteProfile(ctx, lxcCluster.GetProfileName()); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete the default kubeadm profile: %w", err)
	}

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(lxcCluster, infrav1.ClusterFinalizer)

	return ctrl.Result{}, nil
}
