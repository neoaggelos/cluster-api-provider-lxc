package controllers

import (
	"context"

	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

func patchLXCCluster(ctx context.Context, patchHelper *patch.Helper, lxcCluster *infrav1.LXCCluster) error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding it during the deletion process).
	conditions.SetSummary(lxcCluster,
		conditions.WithConditions(
			infrav1.KubeadmProfileAvailableCondition,
			infrav1.LoadBalancerAvailableCondition,
		),
		conditions.WithStepCounterIf(lxcCluster.ObjectMeta.DeletionTimestamp.IsZero() && lxcCluster.Status.FailureReason == nil),
	)

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	return patchHelper.Patch(
		ctx,
		lxcCluster,
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			clusterv1.ReadyCondition,
			infrav1.KubeadmProfileAvailableCondition,
			infrav1.LoadBalancerAvailableCondition,
		}},
	)
}
