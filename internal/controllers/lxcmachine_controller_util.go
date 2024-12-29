package controllers

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

func patchLXCMachine(ctx context.Context, patchHelper *patch.Helper, lxcMachine *infrav1.LXCMachine) error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding the step counter during the deletion process).
	conditions.SetSummary(lxcMachine,
		conditions.WithConditions(
			infrav1.InstanceProvisionedCondition,
			infrav1.BootstrapSucceededCondition,
		),
		conditions.WithStepCounterIf(lxcMachine.ObjectMeta.DeletionTimestamp.IsZero() && lxcMachine.Spec.ProviderID == nil && lxcMachine.Status.FailureReason == nil),
	)

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	return patchHelper.Patch(
		ctx,
		lxcMachine,
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			clusterv1.ReadyCondition,
			infrav1.InstanceProvisionedCondition,
			infrav1.BootstrapSucceededCondition,
		}},
	)
}

func (r *LXCMachineReconciler) getBootstrapData(ctx context.Context, namespace string, dataSecretName string) (string, error) {
	s := &corev1.Secret{}
	key := client.ObjectKey{Namespace: namespace, Name: dataSecretName}
	if err := r.Client.Get(ctx, key, s); err != nil {
		return "", fmt.Errorf("failed to retrieve bootstrap data secret %q: %w", dataSecretName, err)
	}

	value, ok := s.Data["value"]
	if !ok {
		return "", fmt.Errorf("secret %q is missing value key", dataSecretName)
	}

	return string(value), nil
}

func (r *LXCMachineReconciler) setLXCMachineAddresses(lxcMachine *infrav1.LXCMachine, addrs []string) {
	lxcMachine.Status.Addresses = make([]clusterv1.MachineAddress, 0, 1+2*len(addrs))
	lxcMachine.Status.Addresses = append(lxcMachine.Status.Addresses, clusterv1.MachineAddress{
		Type:    clusterv1.MachineHostName,
		Address: lxcMachine.GetInstanceName(),
	})
	for _, address := range addrs {
		lxcMachine.Status.Addresses = append(lxcMachine.Status.Addresses,
			clusterv1.MachineAddress{
				Type:    clusterv1.MachineInternalIP,
				Address: address,
			},
			clusterv1.MachineAddress{
				Type:    clusterv1.MachineExternalIP,
				Address: address,
			},
		)
	}
}
