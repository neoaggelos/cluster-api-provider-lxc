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

package lxcmachine

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/clustercache"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/finalizers"
	utillog "sigs.k8s.io/cluster-api/util/log"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/paused"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/incus"
)

// LXCMachineReconciler reconciles a LXCMachine object
type LXCMachineReconciler struct {
	client.Client
	ClusterCache clustercache.ClusterCache

	// CachingClient is a client that can cache responses, will be used for retrieving secrets.
	CachingClient client.Client

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=lxcmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=lxcmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=lxcmachines/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;machinesets;machines,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets;configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LXCMachine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *LXCMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, rerr error) {
	_ = log.FromContext(ctx)

	// Fetch the LXCMachine instance.
	lxcMachine := &infrav1.LXCMachine{}
	if err := r.Client.Get(ctx, req.NamespacedName, lxcMachine); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// AddOwners adds the owners of LXCMachine as k/v pairs to the logger.
	// Specifically, it will add KubeadmControlPlane, MachineSet and MachineDeployment.
	ctx, log, err := utillog.AddOwners(ctx, r.Client, lxcMachine)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Fetch the Machine.
	machine, err := util.GetOwnerMachine(ctx, r.Client, lxcMachine.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if machine == nil {
		log.Info("Waiting for Machine Controller to set OwnerRef on LXCMachine")
		return ctrl.Result{}, nil
	}

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		log.Info("LXCMachine owner Machine is missing cluster label or cluster does not exist")
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info(fmt.Sprintf("Please associate this machine with a cluster using the label %s: <name of cluster>", clusterv1.ClusterNameLabel))
		return ctrl.Result{}, nil
	}

	ctx = ctrl.LoggerInto(ctx, log.WithValues("Cluster", klog.KObj(cluster)))

	if isPaused, conditionChanged, err := paused.EnsurePausedCondition(ctx, r.Client, cluster, lxcMachine); err != nil || isPaused || conditionChanged {
		return ctrl.Result{}, err
	}

	if cluster.Spec.InfrastructureRef == nil {
		log.Info("Cluster infrastructureRef is not available yet")
		return ctrl.Result{}, nil
	}

	// Fetch the LXC Cluster.
	lxcCluster := &infrav1.LXCCluster{}
	lxcClusterName := client.ObjectKey{
		Namespace: lxcMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, lxcClusterName, lxcCluster); err != nil {
		log.Info("LXCCluster is not available yet")
		return ctrl.Result{}, nil
	}

	// Fetch the lxcSecret before adding any finalizers, so that clusters without a valid secretRef do not get stuck
	lxcSecret := &corev1.Secret{}
	if err := r.Client.Get(ctx, lxcCluster.GetLXCSecretNamespacedName(), lxcSecret); err != nil {
		log.WithValues("secret", lxcCluster.GetLXCSecretNamespacedName()).Error(err, "Failed to fetch LXC credentials secret")
		return ctrl.Result{}, fmt.Errorf("failed to fetch LXC credentials: %w", err)
	}
	lxcClient, err := incus.New(ctx, incus.NewOptionsFromSecret(lxcSecret))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create incus client: %w", err)
	}

	// Add finalizer first if not set to avoid the race condition between init and delete.
	if finalizerAdded, err := finalizers.EnsureFinalizer(ctx, r.Client, lxcMachine, infrav1.MachineFinalizer); err != nil || finalizerAdded {
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(lxcMachine, r)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the LXCMachine object and status after each reconciliation.
	defer func() {
		if err := patchLXCMachine(ctx, patchHelper, lxcMachine); err != nil {
			log.Error(err, "Failed to patch LXCMachine")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted machines
	if !lxcMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.reconcileDelete(ctx, cluster, lxcCluster, machine, lxcMachine, lxcClient)
	}

	result, err := r.reconcileNormal(ctx, cluster, lxcCluster, machine, lxcMachine, lxcClient)
	// Requeue if the reconcile failed because the ClusterCacheTracker was locked for the
	// current cluster because of concurrent access.
	if errors.Is(err, clustercache.ErrClusterNotConnected) {
		log.V(5).Info("Requeuing because connection to the workload cluster is down")
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
	return result, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *LXCMachineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	switch {
	case r.Client == nil:
		return fmt.Errorf("required field Client must not be nil")
	case r.ClusterCache == nil:
		return fmt.Errorf("required field ClusterCache must not be nil")
	case r.CachingClient == nil:
		return fmt.Errorf("required field CachingClient must not be nil")
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
