package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/cloudprovider"
	"github.com/neoaggelos/cluster-api-provider-lxc/internal/incus"
)

func (r *LXCMachineReconciler) reconcileNormal(ctx context.Context, cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, machine *clusterv1.Machine, lxcMachine *infrav1.LXCMachine, lxcClient *incus.Client) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Check if the infrastructure is ready, otherwise return and wait for the cluster object to be updated
	if !cluster.Status.InfrastructureReady {
		log.Info("Waiting for LXCCluster Controller to create cluster infrastructure")
		conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, infrav1.WaitingForClusterInfrastructureReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	}

	// TODO(neoaggelos): enable this code from capd, and adjust for LXC
	/*
		// if the corresponding machine is deleted but the docker machine not yet, update load balancer configuration to divert all traffic from this instance
		if util.IsControlPlaneMachine(machine) && !machine.DeletionTimestamp.IsZero() && dockerMachine.DeletionTimestamp.IsZero() {
			if _, ok := dockerMachine.Annotations["dockermachine.infrastructure.cluster.x-k8s.io/weight"]; !ok {
				if err := r.reconcileLoadBalancerConfiguration(ctx, cluster, dockerCluster, externalLoadBalancer); err != nil {
					return ctrl.Result{}, err
				}
			}
			if dockerMachine.Annotations == nil {
				dockerMachine.Annotations = map[string]string{}
			}
			dockerMachine.Annotations["dockermachine.infrastructure.cluster.x-k8s.io/weight"] = "0"
		}
	*/

	// if the machine is already provisioned, return
	if lxcMachine.Spec.ProviderID != nil {
		lxcMachine.Status.Ready = true

		state, _, err := lxcClient.Client.GetInstanceState(lxcMachine.GetInstanceName())
		if err != nil && strings.Contains(err.Error(), "Instance not found") {
			conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, infrav1.InstanceDeletedReason, clusterv1.ConditionSeverityError, "Instance %s does not exist anymore", lxcMachine.GetInstanceName())
			return ctrl.Result{}, nil
		} else if err == nil {
			conditions.MarkTrue(lxcMachine, infrav1.InstanceProvisionedCondition)
			r.setLXCMachineAddresses(lxcMachine, lxcClient.ParseActiveMachineAddresses(state))

			return ctrl.Result{}, nil
		}
	}

	dataSecretName := machine.Spec.Bootstrap.DataSecretName
	version := machine.Spec.Version
	_ = version

	// Make sure bootstrap data is available and populated.
	if dataSecretName == nil {
		if !util.IsControlPlaneMachine(machine) && !conditions.IsTrue(cluster, clusterv1.ControlPlaneInitializedCondition) {
			log.Info("Waiting for the control plane to be initialized")
			conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, clusterv1.WaitingForControlPlaneAvailableReason, clusterv1.ConditionSeverityInfo, "")
			return ctrl.Result{}, nil
		}

		log.Info("Waiting for the Bootstrap provider controller to set bootstrap data")
		conditions.MarkFalse(lxcMachine, infrav1.InstanceProvisionedCondition, infrav1.WaitingForBootstrapDataReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	}

	// Create the lxc instance hosting the machine
	cloudInit, err := r.getBootstrapData(ctx, lxcMachine.Namespace, *dataSecretName)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to retrieve bootstrap data: %w", err)
	}
	address, err := lxcClient.CreateInstance(ctx, machine, lxcMachine, lxcCluster, cloudInit)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create instance: %w", err)
	}
	r.setLXCMachineAddresses(lxcMachine, address)
	conditions.MarkTrue(lxcMachine, infrav1.InstanceProvisionedCondition)

	// update load balancer
	if util.IsControlPlaneMachine(machine) && !lxcMachine.Status.LoadBalancerConfigured {
		if err := lxcClient.LoadBalancerManagerForCluster(lxcCluster).Reconfigure(ctx); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update loadbalancer configuration: %w", err)
		}
		lxcMachine.Status.LoadBalancerConfigured = true
	}

	// check cloud-init status on the node
	cloudInitStatus, err := lxcClient.CheckCloudInitStatus(ctx, lxcMachine.GetInstanceName())
	if err != nil || cloudInitStatus == incus.CloudInitStatusUnknown {
		log.Error(err, "Could not retrieve cloud-init status")
		conditions.MarkUnknown(lxcMachine, infrav1.BootstrapSucceededCondition, infrav1.BootstrappingUnknownStatusReason, "%s", err)
	}
	switch cloudInitStatus {
	case incus.CloudInitStatusRunning:
		log.Info("Waiting for bootstrap script to complete")
		conditions.MarkFalse(lxcMachine, infrav1.BootstrapSucceededCondition, infrav1.BootstrappingReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{RequeueAfter: 20 * time.Second}, nil
	case incus.CloudInitStatusError:
		err := fmt.Errorf("bootstrap failed")
		log.WithValues("FailureReason", infrav1.FailureReasonBootstrapFailed).Error(err, "Marking machine as failed")
		conditions.MarkFalse(lxcMachine, infrav1.BootstrapSucceededCondition, infrav1.BootstrapFailedReason, clusterv1.ConditionSeverityError, "%s", err)
		lxcMachine.Status.FailureReason = ptr.To(infrav1.FailureReasonBootstrapFailed)
		lxcMachine.Status.FailureMessage = ptr.To(infrav1.FailureMessageBootstrapFailed)
		return ctrl.Result{}, nil
	case incus.CloudInitStatusDone:
		log.Info("Bootstrap finished successfully")
		conditions.MarkTrue(lxcMachine, infrav1.BootstrapSucceededCondition)
	default:
		// This should never happen, but not adding a panic on purpose. If only Go had enums :)
	}

	// If the Cluster is using a control plane and the control plane is not yet initialized, there is no API server
	// to contact to get the ProviderID for the Node hosted on this machine, so return early.
	// NOTE: We are using RequeueAfter with a short interval in order to make test execution time more stable.
	// NOTE: If the Cluster doesn't use a control plane, the ControlPlaneInitialized condition is only
	// set to true after a control plane machine has a node ref. If we would requeue here in this case, the
	// Machine will never get a node ref as ProviderID is required to set the node ref, so we would get a deadlock.
	if cluster.Spec.ControlPlaneRef != nil && !conditions.IsTrue(cluster, clusterv1.ControlPlaneInitializedCondition) {
		log.Info("Waiting for initialized ControlPlane")
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	remoteClient, err := r.ClusterCache.GetClient(ctx, client.ObjectKeyFromObject(cluster))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to generate workload cluster client: %w", err)
	}

	if err := cloudprovider.PatchNode(ctx, remoteClient, lxcMachine); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to apply cloud-provider node patch: %w", err)
	}

	lxcMachine.Status.Ready = true
	lxcMachine.Spec.ProviderID = ptr.To(lxcMachine.GetExpectedProviderID())

	return ctrl.Result{}, nil
}
