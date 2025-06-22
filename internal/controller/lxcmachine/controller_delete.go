package lxcmachine

import (
	"context"
	"fmt"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/loadbalancer"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

func (r *LXCMachineReconciler) reconcileDelete(ctx context.Context, cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, machine *clusterv1.Machine, lxcMachine *infrav1.LXCMachine, lxcClient *lxc.Client) error {
	// Set the InstanceProvisionedCondition reporting delete is started, and issue a patch in order to make
	// this visible to the users.
	// NB. The operation in LXC is fast, so there is the chance the user will not notice the status change;
	// nevertheless we are issuing a patch so we can test a pattern that will be used by other providers as well
	patchHelper, err := patch.NewHelper(lxcMachine, r.Client)
	if err != nil {
		return err
	}
	conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, clusterv1.DeletingReason, clusterv1.ConditionSeverityInfo, "")
	if err := patchLXCMachine(ctx, patchHelper, lxcMachine); err != nil {
		return fmt.Errorf("failed to patch LXCMachine: %w", err)
	}

	// Delete the machine
	log.FromContext(ctx).Info("Deleting instance")
	if err := lxcClient.WaitForDeleteInstance(ctx, lxcMachine.Name); err != nil {
		return fmt.Errorf("failed to delete the instance: %w", err)
	}

	// If the deleted machine is a control-plane node, remove it from the load balancer configuration (unless the cluster is getting deleted)
	if util.IsControlPlaneMachine(machine) && cluster.DeletionTimestamp.IsZero() {
		log.FromContext(ctx).Info("Reconfigure load balancer after removing control plane machine")
		if err := loadbalancer.ManagerForCluster(cluster, lxcCluster, lxcClient).Reconfigure(ctx); err != nil {
			return fmt.Errorf("failed to reconfigure load balancer after removing control plane node: %w", err)
		}
	}

	// Machine is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(lxcMachine, infrav1.MachineFinalizer)

	return nil
}
