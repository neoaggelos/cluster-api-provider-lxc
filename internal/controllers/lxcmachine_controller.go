/*
Copyright 2024 Angelos Kolaitis.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/cluster-api/controllers/clustercache"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

// LXCMachineReconciler reconciles a LXCMachine object
type LXCMachineReconciler struct {
	client.Client
	ClusterCache clustercache.ClusterCache

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=lxcmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=lxcmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=lxcmachines/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;machinesets;machines,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets;,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LXCMachine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *LXCMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LXCMachineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	switch {
	case r.Client == nil:
		return fmt.Errorf("required field Client must not be nil")
	case r.ClusterCache == nil:
		return fmt.Errorf("required field ClusterCache must not be nil")
	}

	predicateLog := ctrl.LoggerFrom(ctx).WithValues("controller", "lxcmachine")
	clusterToLXCMachines, err := util.ClusterToTypedObjectsMapper(mgr.GetClient(), &infrav1.LXCMachineList{}, mgr.GetScheme())
	if err != nil {
		return err
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.LXCMachine{}).
		WithOptions(options).
		WithEventFilter(predicates.ResourceHasFilterLabel(mgr.GetScheme(), predicateLog, r.WatchFilterValue)).
		Watches(
			&clusterv1.Machine{},
			handler.EnqueueRequestsFromMapFunc(util.MachineToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("LXCMachine"))),
		).
		Watches(
			&infrav1.LXCCluster{},
			handler.EnqueueRequestsFromMapFunc(r.LXCClusterToLXCMachines),
		).
		Watches(
			&clusterv1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(clusterToLXCMachines),
			builder.WithPredicates(
				predicates.ClusterPausedTransitionsOrInfrastructureReady(mgr.GetScheme(), predicateLog),
			),
		).
		WatchesRawSource(r.ClusterCache.GetClusterSource("lxcmachine", clusterToLXCMachines)).
		Complete(r); err != nil {
		return fmt.Errorf("failed setting up with a controller manager: %w", err)
	}

	return nil
}

// LXCClusterToLXCMachines is a handler.ToRequestsFunc to be used to enqueue
// requests for reconciliation of LXCMachines.
func (r *LXCMachineReconciler) LXCClusterToLXCMachines(ctx context.Context, o client.Object) []ctrl.Request {
	c, ok := o.(*infrav1.LXCCluster)
	if !ok {
		panic(fmt.Sprintf("Expected a LXCCluster but got a %T", o))
	}

	cluster, err := util.GetOwnerCluster(ctx, r.Client, c.ObjectMeta)
	switch {
	case apierrors.IsNotFound(err) || cluster == nil:
		return nil
	case err != nil:
		return nil
	}

	labels := map[string]string{clusterv1.ClusterNameLabel: cluster.Name}
	machineList := &clusterv1.MachineList{}
	if err := r.Client.List(ctx, machineList, client.InNamespace(c.Namespace), client.MatchingLabels(labels)); err != nil {
		return nil
	}
	result := make([]ctrl.Request, 0, len(machineList.Items))
	for _, m := range machineList.Items {
		if m.Spec.InfrastructureRef.Name == "" {
			continue
		}
		result = append(result, ctrl.Request{NamespacedName: client.ObjectKey{
			Namespace: m.Namespace,
			Name:      m.Spec.InfrastructureRef.Name,
		}})
	}

	return result
}
