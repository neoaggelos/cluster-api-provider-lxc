package lxccluster

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/loadbalancer"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

func (r *LXCClusterReconciler) reconcileDelete(ctx context.Context, cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, lxcClient *lxc.Client) (ctrl.Result, error) {
	// Set the LoadBalancerAvailableCondition reporting delete is started, and issue a patch in order to make
	// this visible to the users.
	// NB. The operation in LXC is fast, so there is the chance the user will not notice the status change;
	// nevertheless we are issuing a patch so we can test a pattern that will be used by other providers as well
	patchHelper, err := patch.NewHelper(lxcCluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	conditions.Set(lxcCluster, metav1.Condition{Type: infrav1.LoadBalancerAvailableCondition, Status: metav1.ConditionFalse, Reason: clusterv1.DeletingReason})
	if err := patchLXCCluster(ctx, patchHelper, lxcCluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to patch LXCCluster: %w", err)
	}

	// Delete the container hosting the load balancer
	log.FromContext(ctx).Info("Deleting load balancer")
	if err := loadbalancer.ManagerForCluster(cluster, lxcCluster, lxcClient).Delete(ctx); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete the load balancer instance: %w", err)
	}

	machines, err := utils.GetMachinesForCluster(ctx, r.Client, client.ObjectKeyFromObject(cluster))
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to get list of Machines for Cluster")
	} else if len(machines) > 0 {
		log.FromContext(ctx).WithValues("machines", len(machines)).Info("Waiting for all Machines to be deleted")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(lxcCluster, infrav1.ClusterFinalizer)

	return ctrl.Result{}, nil
}
