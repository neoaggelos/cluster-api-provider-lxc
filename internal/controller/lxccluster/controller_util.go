package lxccluster

import (
	"context"
	"fmt"

	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
)

func patchLXCCluster(ctx context.Context, patchHelper *patch.Helper, lxcCluster *infrav1.LXCCluster) error {
	infraConditions := []string{ //nolint:prealloc
		infrav1.LoadBalancerAvailableCondition,
	}

	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding it during the deletion process).
	if err := conditions.SetSummaryCondition(lxcCluster, lxcCluster, clusterv1.ReadyCondition, conditions.ForConditionTypes(infraConditions)); err != nil {
		return fmt.Errorf("failed to set summary condition: %w", err)
	}

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	return patchHelper.Patch(
		ctx,
		lxcCluster,
		patch.WithOwnedConditions{Conditions: append(infraConditions, clusterv1.ReadyCondition)},
		patch.WithStatusObservedGeneration{},
	)
}
