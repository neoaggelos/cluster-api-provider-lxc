package lxccluster

import (
	"context"
	"slices"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
)

func patchLXCCluster(ctx context.Context, patchHelper *patch.Helper, lxcCluster *infrav1.LXCCluster) error {
	infraConditions := []clusterv1.ConditionType{
		infrav1.LoadBalancerAvailableCondition,
	}
	hasInfraConditionError := false
	for _, condition := range lxcCluster.GetConditions() {
		// slices.Contains is fast enough as we only have < 5 conditions
		if slices.Contains(infraConditions, condition.Type) && condition.Severity == clusterv1.ConditionSeverityError {
			hasInfraConditionError = true
			break
		}
	}

	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding it during the deletion process).
	conditions.SetSummary(lxcCluster,
		conditions.WithConditions(infraConditions...),
		conditions.WithStepCounterIf(lxcCluster.DeletionTimestamp.IsZero() && !hasInfraConditionError),
	)

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	return patchHelper.Patch(
		ctx,
		lxcCluster,
		patch.WithOwnedConditions{Conditions: append(infraConditions, clusterv1.ReadyCondition)},
	)
}
