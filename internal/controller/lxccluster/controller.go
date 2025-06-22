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

package lxccluster

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/finalizers"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/paused"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// LXCClusterReconciler reconciles a LXCCluster object
type LXCClusterReconciler struct {
	client.Client

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=lxcclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=lxcclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=lxcclusters/finalizers,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *LXCClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, rerr error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the lxcCluster instance
	lxcCluster := &infrav1.LXCCluster{}
	if err := r.Get(ctx, req.NamespacedName, lxcCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, lxcCluster.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info("Waiting for Cluster Controller to set OwnerRef on LXCCluster")
		return ctrl.Result{}, nil
	}

	// Fetch the lxcSecret before adding any finalizers, so that clusters without a valid secretRef do not get stuck
	lxcSecret := &corev1.Secret{}
	if err := r.Get(ctx, lxcCluster.GetLXCSecretNamespacedName(), lxcSecret); err != nil {
		log.WithValues("secret", lxcCluster.GetLXCSecretNamespacedName()).Error(err, "Failed to fetch LXC credentials secret")
		return ctrl.Result{}, fmt.Errorf("failed to fetch LXC credentials: %w", err)
	}
	lxcClient, err := lxc.New(ctx, lxc.ConfigurationFromKubernetesSecret(lxcSecret))
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create incus client: %w", err)
	}

	// Add finalizer first if not set to avoid the race condition between init and delete.
	if finalizerAdded, err := finalizers.EnsureFinalizer(ctx, r.Client, lxcCluster, infrav1.ClusterFinalizer); err != nil || finalizerAdded {
		return ctrl.Result{}, err
	}

	if isPaused, conditionChanged, err := paused.EnsurePausedCondition(ctx, r.Client, cluster, lxcCluster); err != nil || isPaused || conditionChanged {
		return ctrl.Result{}, err
	}

	log = log.WithValues("Cluster", klog.KObj(cluster))
	ctx = ctrl.LoggerInto(ctx, log)

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(lxcCluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the LXCCluster object and status after each reconciliation.
	defer func() {
		if err := patchLXCCluster(ctx, patchHelper, lxcCluster); err != nil {
			log.Error(err, "Failed to patch LXCCluster")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted clusters
	if !lxcCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, cluster, lxcCluster, lxcClient)
	}

	// Handle non-deleted clusters
	return ctrl.Result{}, r.reconcileNormal(ctx, cluster, lxcCluster, lxcClient)
}

// SetupWithManager sets up the controller with the Manager.
func (r *LXCClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	if r.Client == nil {
		return fmt.Errorf("required field Client must not be nil")
	}

	predicateLog := ctrl.LoggerFrom(ctx).WithValues("controller", "lxccluster")

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.LXCCluster{}).
		WithOptions(options).
		WithEventFilter(predicates.ResourceHasFilterLabel(mgr.GetScheme(), predicateLog, r.WatchFilterValue)).
		Watches(
			&clusterv1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(util.ClusterToInfrastructureMapFunc(ctx, infrav1.GroupVersion.WithKind("LXCCluster"), mgr.GetClient(), &infrav1.LXCCluster{})),
			builder.WithPredicates(
				predicates.ClusterPausedTransitions(mgr.GetScheme(), predicateLog),
			),
		).Complete(r); err != nil {
		return fmt.Errorf("failed setting up with a controller manager: %w", err)
	}

	return nil
}
