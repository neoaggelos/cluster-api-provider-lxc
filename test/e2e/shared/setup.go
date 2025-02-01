//go:build e2e

package shared

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	"sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/gomega"
)

// FixupNamespace installs the LXC credentials secret and CNI resources configmap on the target namespace.
func FixupNamespace(e2eCtx *E2EContext, namespace string, enableCredentials bool, enableCNIResources bool) {
	clusterClient := e2eCtx.Environment.BootstrapClusterProxy.GetClient()

	for _, item := range []struct {
		description string
		object      client.Object
		enabled     bool
	}{
		{
			enabled:     enableCredentials,
			description: "secret/lxc-credentials",
			object:      e2eCtx.Settings.LXCClientOptions.ToSecret(e2eCtx.E2EConfig.GetVariable(LXCSecretName), namespace),
		},
		{
			enabled:     enableCNIResources,
			description: "configmap/cni-resources",
			object: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cni-resources",
					Namespace: namespace,
				},
				Data: map[string]string{
					"cni.yaml": e2eCtx.Settings.CNIManifest,
				},
			},
		},
		{
			enabled:     enableCNIResources,
			description: "clusterresourceset/cni-resource-set",
			object: &addonsv1.ClusterResourceSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cni-resource-set",
					Namespace: namespace,
				},
				Spec: addonsv1.ClusterResourceSetSpec{
					Strategy: "ApplyOnce",
					Resources: []addonsv1.ResourceRef{
						{Kind: "ConfigMap", Name: "cni-resources"},
					},
					ClusterSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"cni": "cni-resources",
						},
					},
				},
			},
		},
	} {
		if item.enabled {
			e2e.Byf("Creating resource %s on namespace %s", item.description, namespace)

			Expect(clusterClient.Create(context.TODO(), item.object)).To(Or(Succeed(), MatchError(apierrors.IsAlreadyExists, "apierrors.IsAlreadyExists")), "Failed to deploy %s", item.description)
		} else {
			e2e.Byf("Skipping resource %s on namespace %s", item.description, namespace)
		}
	}
}

// FixupWorkloadCluster patches the workload cluster object to install CNI.
func FixupWorkloadCluster(e2eCtx *E2EContext, name string, namespace string, cni bool) {
	clusterClient := e2eCtx.Environment.BootstrapClusterProxy.GetClient()

	clusterName := types.NamespacedName{Name: name, Namespace: namespace}
	cluster := &clusterv1.Cluster{}

	e2e.Byf("Fetch workload cluster %v", clusterName)
	Expect(clusterClient.Get(context.TODO(), clusterName, cluster)).To(Succeed(), "Failed to retrieve workload cluster")

	// Label cluster to match ClusterResourceSet that deploys the CNI
	if cni {
		e2e.Byf("Label workload cluster %v with cni=cni-resources", clusterName)

		if cluster.Labels == nil {
			cluster.Labels = make(map[string]string, 1)
		}
		cluster.Labels["cni"] = "cni-resources"
	}

	e2e.Byf("Patch workload cluster %v", clusterName)
	Expect(clusterClient.Update(context.TODO(), cluster)).To(Succeed(), "Failed to patch workload cluster with necessary labels and configs")
}
