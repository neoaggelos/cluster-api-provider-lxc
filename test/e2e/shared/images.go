//go:build e2e

package shared

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/test/e2e"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"

	. "github.com/onsi/gomega"
)

// defaultImages to download before running the e2e tests.
func defaultImages(e2eCtx *E2EContext, serverName string) []string {
	images := []string{
		UbuntuImage,
		DebianImage,
		fmt.Sprintf("capi:kubeadm/%s", e2eCtx.E2EConfig.MustGetVariable(KubernetesVersion)),
		fmt.Sprintf("capi:kubeadm/%s", e2eCtx.E2EConfig.MustGetVariable(KubernetesVersionUpgradeFrom)),
		fmt.Sprintf("capi:kubeadm/%s", e2eCtx.E2EConfig.MustGetVariable(KubernetesVersionUpgradeTo)),
	}
	if serverName == lxc.Incus {
		images = append(images, fmt.Sprintf("kind:%s", e2eCtx.E2EConfig.MustGetVariable(KubernetesVersion)))
	}
	return images
}

func ensureLXCSystemImages(e2eCtx *E2EContext) {
	lxcClient, err := lxc.New(context.TODO(), e2eCtx.Settings.LXCClientOptions)
	Expect(err).ToNot(HaveOccurred(), "Failed to initialize client")

	for _, imageName := range defaultImages(e2eCtx, lxcClient.GetServerName()) {
		e2e.Byf("Fetching image %s", imageName)

		image, parsed, err := lxc.ParseImage(imageName)
		Expect(err).ToNot(HaveOccurred(), "Image %s not recognized", imageName)
		Expect(parsed).To(BeTrue(), "Image prefix not recognized")

		Expect(lxcClient.PullImage(context.TODO(), image)).To(Succeed(), "Failed to pull image %s", imageName)
	}
}
