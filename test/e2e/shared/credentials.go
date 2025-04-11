//go:build e2e

package shared

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lxc/cluster-api-provider-incus/internal/incus"

	. "github.com/onsi/gomega"
)

func ensureLXCClientOptions(e2eCtx *E2EContext) {
	clusterClient := e2eCtx.Environment.BootstrapClusterProxy.GetClient()

	secretName := types.NamespacedName{
		Name:      e2eCtx.E2EConfig.GetVariable(LXCSecretName),
		Namespace: "default",
	}

	secret := &corev1.Secret{}
	if err := clusterClient.Get(context.TODO(), secretName, secret); err != nil {
		Expect(apierrors.IsNotFound(err)).To(BeTrue(), "Failed to retrieve the secret with infrastructure credentials")
	} else {
		Logf("Found existing secret in the cluster %s", secretName)
		e2eCtx.Settings.LXCClientOptions = incus.NewOptionsFromSecret(secret)
		return
	}

	configFile := e2eCtx.E2EConfig.GetVariable(LXCLoadConfigFile)
	remoteName := e2eCtx.E2EConfig.GetVariable(LXCLoadRemoteName)
	Logf("Looking for infrastructure credentials in local node (configFile: %q, remoteName: %q)", configFile, remoteName)
	options, err := incus.NewOptionsFromConfigFile(configFile, remoteName, true)
	Expect(err).ToNot(HaveOccurred(), "Failed to find infrastructure credentials in local node")

	e2eCtx.Settings.LXCClientOptions = options

	// validate client options
	Expect(e2eCtx.Settings.LXCClientOptions).ToNot(BeZero(), "Could not detect infrastructure credentials from local node or existing secret")
	client, err := incus.New(context.TODO(), e2eCtx.Settings.LXCClientOptions)
	Expect(err).ToNot(HaveOccurred(), "Failed to initialize client")

	_, err = client.Client.GetProfileNames()
	Expect(err).ToNot(HaveOccurred())
}
