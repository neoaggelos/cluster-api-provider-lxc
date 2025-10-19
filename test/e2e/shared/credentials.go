//go:build e2e

package shared

import (
	"context"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"

	. "github.com/onsi/gomega"
)

func ensureLXCClientOptions(e2eCtx *E2EContext) {
	configFile := e2eCtx.E2EConfig.GetVariableOrEmpty(LXCLoadConfigFile)
	remoteName := e2eCtx.E2EConfig.GetVariableOrEmpty(LXCLoadRemoteName)
	Logf("Looking for infrastructure credentials in local node (configFile: %q, remoteName: %q)", configFile, remoteName)
	options, _, err := lxc.ConfigurationFromLocal(configFile, remoteName, false)
	Expect(err).ToNot(HaveOccurred(), "Failed to find infrastructure credentials in local node")

	e2eCtx.Settings.LXCClientOptions = options

	// validate client options
	Expect(e2eCtx.Settings.LXCClientOptions).ToNot(BeZero(), "Could not detect infrastructure credentials from local node")
	lxcClient, err := lxc.New(context.TODO(), e2eCtx.Settings.LXCClientOptions)
	Expect(err).ToNot(HaveOccurred(), "Failed to initialize client")

	_, err = lxcClient.GetProfileNames()
	Expect(err).ToNot(HaveOccurred())
}
