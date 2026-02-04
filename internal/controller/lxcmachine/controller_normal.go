package lxcmachine

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/cloudprovider"
	"github.com/lxc/cluster-api-provider-incus/internal/loadbalancer"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

func (r *LXCMachineReconciler) reconcileNormal(ctx context.Context, cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, machine *clusterv1.Machine, lxcMachine *infrav1.LXCMachine, lxcClient *lxc.Client) (ctrl.Result, error) {
	// Check if the infrastructure is ready, otherwise return and wait for the cluster object to be updated
	if !cluster.Status.InfrastructureReady {
		log.FromContext(ctx).Info("Waiting for LXCCluster Controller to create cluster infrastructure")
		conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, infrav1.WaitingForClusterInfrastructureReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	}

	// if the machine is already provisioned, return
	if lxcMachine.Spec.ProviderID != nil {
		state, _, err := lxcClient.GetInstanceState(lxcMachine.GetInstanceName())
		if err != nil {
			if strings.Contains(err.Error(), "Instance not found") {
				lxcMachine.Status.Ready = false
				conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, infrav1.InstanceDeletedReason, clusterv1.ConditionSeverityError, "Instance %s does not exist anymore", lxcMachine.GetInstanceName())
				return ctrl.Result{}, nil
			}

			log.FromContext(ctx).Error(err, "Failed to check instance state")
			return ctrl.Result{}, err
		}

		lxcMachine.Status.Ready = true
		conditions.MarkTrue(lxcMachine, infrav1.InstanceProvisionedCondition)
		r.setLXCMachineAddresses(lxcMachine, lxc.ParseHostAddresses(state))

		// Handle cloud provider node patch
		if lxcCluster.Spec.CloudProviderNodePatch && !lxcMachine.Status.CloudProviderNodePatchConfigured {
			// If the Cluster is using a control plane and the control plane is not yet initialized, there is no API server
			// to contact to get the ProviderID for the Node hosted on this machine, so return early.
			// NOTE: We are using RequeueAfter with a short interval in order to make test execution time more stable.
			// NOTE: If the Cluster doesn't use a control plane, the ControlPlaneInitialized condition is only
			// set to true after a control plane machine has a node ref. If we would requeue here in this case, the
			// Machine will never get a node ref as ProviderID is required to set the node ref, so we would get a deadlock.
			if cluster.Spec.ControlPlaneRef != nil && !conditions.IsTrue(cluster, clusterv1.ControlPlaneInitializedCondition) {
				log.FromContext(ctx).Info("Waiting for initialized ControlPlane")
				return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
			}

			remoteClient, err := r.ClusterCache.GetClient(ctx, client.ObjectKeyFromObject(cluster))
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to generate workload cluster client: %w", err)
			}

			remoteNode := &corev1.Node{}
			if err := remoteClient.Get(ctx, types.NamespacedName{Name: lxcMachine.GetInstanceName()}, remoteNode); err != nil {
				// NOTE(neoaggelos): we assume the node will appear with a name that matches the lxcMachine instance name.
				// This might not be true in case of a non-Ubuntu OS (e.g. hostname vs fqdn), or in case a custom node name is set.
				//
				// However: this is what capd does, and the situations described above should be infrequent to not worry about right now.
				if apierrors.IsNotFound(err) {
					log.FromContext(ctx).Info("Waiting for node to appear in workload cluster")
					return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
				}
				return ctrl.Result{}, fmt.Errorf("failed to retrieve node with name %q from workload cluster: %w", lxcMachine.GetInstanceName(), err)
			}

			if err := cloudprovider.PatchNode(ctx, lxcMachine, remoteClient, remoteNode); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to apply cloud-provider node patch: %w", err)
			}

			lxcMachine.Status.CloudProviderNodePatchConfigured = true
		}

		return ctrl.Result{}, nil
	}

	dataSecretName := machine.Spec.Bootstrap.DataSecretName

	// Make sure bootstrap data is available and populated.
	if dataSecretName == nil {
		if !util.IsControlPlaneMachine(machine) && !conditions.IsTrue(cluster, clusterv1.ControlPlaneInitializedCondition) {
			log.FromContext(ctx).Info("Waiting for the control plane to be initialized")
			conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, clusterv1.WaitingForControlPlaneAvailableReason, clusterv1.ConditionSeverityInfo, "")
			return ctrl.Result{}, nil
		}

		log.FromContext(ctx).Info("Waiting for the Bootstrap provider controller to set bootstrap data")
		conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, infrav1.WaitingForBootstrapDataReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	}

	// Create the lxc instance hosting the machine
	cloudInit, err := r.getBootstrapData(ctx, lxcMachine.Namespace, *dataSecretName)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to retrieve bootstrap data: %w", err)
	}

	log.FromContext(ctx).Info("Launching instance")
	addresses, err := launchInstance(ctx, cluster, lxcCluster, machine, lxcMachine, lxcClient, cloudInit)
	if err != nil {
		if utils.IsTerminalError(err) {
			log.FromContext(ctx).Error(err, "Fatal error while creating instance spec")
			conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, infrav1.InstanceProvisioningAbortedReason, clusterv1.ConditionSeverityError, "Failed to create instance spec: %s", err.Error())
			return ctrl.Result{}, nil
		}
		if strings.HasSuffix(err.Error(), "context deadline exceeded") {
			log.FromContext(ctx).Error(err, "Instance creation timed out, retrying in 10 seconds")
			conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, infrav1.CreatingInstanceReason, clusterv1.ConditionSeverityWarning, "Instance creation still in progress: %s", err.Error())
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
		conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, infrav1.InstanceProvisioningFailedReason, clusterv1.ConditionSeverityWarning, "Failed to create instance: %s", err.Error())
		return ctrl.Result{}, fmt.Errorf("failed to create instance: %w", err)
	}
	r.setLXCMachineAddresses(lxcMachine, addresses)
	conditions.MarkTrue(lxcMachine, infrav1.InstanceProvisionedCondition)

	// update load balancer
	if util.IsControlPlaneMachine(machine) && !lxcMachine.Status.LoadBalancerConfigured {
		log.FromContext(ctx).Info("Updating control plane load balancer")

		if err := loadbalancer.ManagerForCluster(cluster, lxcCluster, lxcClient).Reconfigure(ctx); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update loadbalancer configuration: %w", err)
		}
		lxcMachine.Status.LoadBalancerConfigured = true
	}

	lxcMachine.Spec.ProviderID = ptr.To(lxcMachine.GetExpectedProviderID())
	lxcMachine.Status.Ready = true

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}
