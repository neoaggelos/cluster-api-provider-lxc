//go:build e2e

package shared

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// OverrideVariables creates a clusterctl config with variable overrides and updates the e2eCtx. It will DeferCleanup() to restore the previous configs after the spec is finished.
//
// TODO: this is working around external specs that do not expose the ClusterctlVariables input.
// Can remove after https://github.com/kubernetes-sigs/cluster-api/pull/11780
func (e2eCtx *E2EContext) OverrideVariables(variables map[string]string) {
	Expect(e2eCtx.Environment.ClusterctlConfigPath).To(BeAnExistingFile(), "clusterctlConfigPath should be an existing file and point to the clusterctl config file in use by the E2E tests.")

	oldPath := e2eCtx.Environment.ClusterctlConfigPath
	oldVariables := e2eCtx.E2EConfig.DeepCopy().Variables
	DeferCleanup(func() {
		Logf("Restoring config file %s", oldPath)
		e2eCtx.Environment.ClusterctlConfigPath = oldPath
		e2eCtx.E2EConfig.Variables = oldVariables
	})

	oldConfig, err := os.ReadFile(oldPath)
	Expect(err).ToNot(HaveOccurred())

	var config map[string]any
	Expect(yaml.Unmarshal(oldConfig, &config)).To(Succeed())
	for k, v := range variables {
		config[k] = v
		e2eCtx.E2EConfig.Variables[k] = v
	}

	b, err := yaml.Marshal(config)
	Expect(err).ToNot(HaveOccurred())

	configPath := fmt.Sprintf("%s.%d.yaml", e2eCtx.Environment.ClusterctlConfigPath, GinkgoParallelProcess())
	Expect(os.WriteFile(configPath, b, 0o600)).To(Succeed())

	Logf("Using config file %s with overrides %v", configPath, variables)
	e2eCtx.Environment.ClusterctlConfigPath = configPath
}
