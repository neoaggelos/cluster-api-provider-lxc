//go:build e2e

package shared

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

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
			object:      e2eCtx.Settings.LXCClientOptions.ToSecret(e2eCtx.E2EConfig.MustGetVariable(LXCSecretName), namespace),
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
