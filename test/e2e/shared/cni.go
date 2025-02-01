//go:build e2e

package shared

import (
	"os"

	"sigs.k8s.io/cluster-api/test/e2e"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"

	. "github.com/onsi/gomega"
)

func loadCNIManifest(config *clusterctl.E2EConfig) string {
	Expect(config.Variables).To(HaveKey(e2e.CNIPath), "Missing %s variable in the config", e2e.CNIPath)
	cniPath := config.GetVariable(e2e.CNIPath)
	Expect(cniPath).To(BeAnExistingFile(), "The %s variable should resolve to an existing file", e2e.CNIPath)

	b, err := os.ReadFile(cniPath)
	Expect(err).ToNot(HaveOccurred(), "Failed to read CNI manifest from %q", cniPath)

	return string(b)
}
