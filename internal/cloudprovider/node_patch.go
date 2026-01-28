package cloudprovider

import (
	"context"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
)

var (
	cloudProviderTaint = corev1.Taint{Key: "node.cloudprovider.kubernetes.io/uninitialized", Effect: corev1.TaintEffectNoSchedule}
)

func matchesCloudProviderTaint(taint corev1.Taint) bool {
	return taint.MatchTaint(&cloudProviderTaint)
}

// PatchNode implements the responsibilities of the external cloud-provider for node objects in the workload cluster.
// It is required because there is no external cloud-provider integration between Kubernetes and Incus.
func PatchNode(ctx context.Context, remoteClient client.Client, lxcMachine *infrav1.LXCMachine) error {
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

	// slices.DeleteFunc() preserves the length of the original slice and zeroes out last elements. If length is different, it means node originally had the taint.
	if taints := slices.DeleteFunc(remoteNode.Spec.Taints, matchesCloudProviderTaint); len(taints) != len(remoteNode.Spec.Taints) {
		// 2. set addresses in the remoteNode `.status.addresses`
		remoteNodeAddressSet := sets.New(remoteNode.Status.Addresses...)
		for _, address := range lxcMachine.Status.Addresses {
			// Only consider "InternalIP" addresses
			if address.Type != clusterv1.MachineInternalIP {
				continue
			}
			// Append address if not already set in the remoteNode
			nodeAddress := corev1.NodeAddress{Type: corev1.NodeAddressType(address.Type), Address: address.Address}
			if !remoteNodeAddressSet.Has(nodeAddress) {
				log.FromContext(ctx).WithValues("address", nodeAddress).Info("Adding missing machine address to remote node")
				remoteNode.Status.Addresses = append(remoteNode.Status.Addresses, nodeAddress)
			}
		}

		// 3. remove cloud provider taint to initialize node
		log.FromContext(ctx).WithValues("taint", cloudProviderTaint).Info("Removing cloud provider taint to initialize remote node")
		remoteNode.Spec.Taints = taints
	}

	if err := patchHelper.Patch(ctx, remoteNode); err != nil {
		return fmt.Errorf("failed to patch remote node: %w", err)
	}
	return nil
}
