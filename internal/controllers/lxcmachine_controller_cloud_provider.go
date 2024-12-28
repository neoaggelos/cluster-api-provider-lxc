package controllers

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"

	corev1 "k8s.io/api/core/v1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

var (
	cloudProviderTaint = corev1.Taint{Key: "node.cloudprovider.kubernetes.io/uninitialized", Effect: corev1.TaintEffectNoSchedule}
)

// cloudProviderNodePatch implements the responsibilities of the external cloud-provider for node objects in the workload cluster.
// it is required because there is no external cloud-provider integration between Kubernetes and Incus.
func (r *LXCMachineReconciler) cloudProviderNodePatch(ctx context.Context, remoteClient client.Client, lxcMachine *infrav1.LXCMachine) error {
	expectedProviderID := lxcMachine.GetExpectedProviderID()
	expectedRemoteNodeName := lxcMachine.GetInstanceName()

	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("providerID", expectedProviderID, "nodeName", expectedRemoteNodeName))

	remoteNode := &corev1.Node{}
	if err := remoteClient.Get(ctx, types.NamespacedName{Name: lxcMachine.GetInstanceName()}, remoteNode); err != nil {
		// NOTE(neoaggelos): we assume the node will appear with a name that matches the lxcMachine instance name.
		// This might not be true in case of a non-Ubuntu OS (e.g. hostname vs fqdn), or in case a custom node name is set.
		//
		// However: this is what capd does, and the situations described above should be infrequent to not worry about right now.
		return fmt.Errorf("failed to retrieve node with name %q from workload cluster: %w", lxcMachine.GetInstanceName(), err)
	}

	patchHelper, err := patch.NewHelper(remoteNode, remoteClient)
	if err != nil {
		return err
	}

	// 1. set providerID on the remote node
	log.FromContext(ctx).Info("Setting providerID on remote node")
	remoteNode.Spec.ProviderID = expectedProviderID

	// (TODO) 2. if the node has the external cloud provider taint, add any missing node addresses

	// 3. if the node has the external cloud provider taint, remove it
	// TODO: request sigs.k8s.io/cluster-api/internal/taints to be made public to reuse
	newTaints := make([]corev1.Taint, 0, len(remoteNode.Spec.Taints))
	for _, taint := range remoteNode.Spec.Taints {
		if taint.MatchTaint(&cloudProviderTaint) {
			log.FromContext(ctx).WithValues("taint", taint).Info("Removing cloud provider taint to initialize remote node")
			continue
		}

		newTaints = append(newTaints, taint)
	}
	remoteNode.Spec.Taints = newTaints

	if err := patchHelper.Patch(ctx, remoteNode); err != nil {
		return fmt.Errorf("failed to patch remote node: %w", err)
	}
	return nil
}
