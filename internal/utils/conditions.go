package utils

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/paused"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EnsureConditionReasons is used to migrate from v1beta1 ClusterAPI contract.
//
// When applying the v1beta2 APIs
// and v1beta1 objects already exist on the cluster, their status.conditions changes from []clusterv1.Condition
// to []metav1.Condition. These two types differ like so:
//   - clusterv1.Condition has no Reason field
//   - metav1.Condition has a Reason field which is required (min_length >= 1)
//
// Because of this, any subsequent patch operations will fail unless a .Reason is set on these conditions.
func EnsureConditionReasons(ctx context.Context, client client.Client, obj paused.ConditionSetter) (bool, error) {
	patchHelper, err := patch.NewHelper(obj, client)
	if err != nil {
		return false, err
	}

	var changed bool
	conditions := obj.GetConditions()
	for i := range conditions {
		if conditions[i].Reason == "" {
			conditions[i].Reason = conditions[i].Type
			changed = true
		}
	}
	if !changed {
		return false, nil
	}

	if err := patchHelper.Patch(ctx, obj); err != nil {
		return false, fmt.Errorf("failed to patch v1beta1 conditions: %w", err)
	}
	return true, nil
}
