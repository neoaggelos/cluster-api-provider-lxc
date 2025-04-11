//go:build e2e

/*
Copyright 2021 The Kubernetes Authors.

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

package shared

import (
	"context"
	"flag"
	"os"
	"path"
	"path/filepath"

	"sigs.k8s.io/cluster-api/test/framework"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"
	"sigs.k8s.io/yaml"

	"github.com/lxc/cluster-api-provider-incus/internal/incus"

	. "github.com/onsi/gomega"
)

type synchronizedBeforeTestSuiteConfig struct {
	ArtifactFolder       string               `json:"artifactFolder,omitempty"`
	ConfigPath           string               `json:"configPath,omitempty"`
	ClusterctlConfigPath string               `json:"clusterctlConfigPath,omitempty"`
	KubeconfigPath       string               `json:"kubeconfigPath,omitempty"`
	E2EConfig            clusterctl.E2EConfig `json:"e2eConfig,omitempty"`

	LXCClientOptions incus.Options `json:"lxcClientOptions,omitempty"`
	CNIManifest      string        `json:"cni,omitempty"`
}

// Node1BeforeSuite is the common setup down on the first ginkgo node before the test suite runs.
func Node1BeforeSuite(ctx context.Context, e2eCtx *E2EContext) []byte {
	flag.Parse()
	Expect(e2eCtx.Settings.ConfigPath).To(BeAnExistingFile(), "Invalid test suite argument. configPath should be an existing file.")
	Expect(os.MkdirAll(e2eCtx.Settings.ArtifactFolder, 0o750)).To(Succeed(), "Invalid test suite argument. Can't create artifacts-folder %q", e2eCtx.Settings.ArtifactFolder)
	templatesDir := path.Join(e2eCtx.Settings.ArtifactFolder, "templates")
	Expect(os.MkdirAll(templatesDir, 0o750)).To(Succeed(), "Can't create templates folder %q", templatesDir)
	Logf("Loading the e2e test configuration from %q", e2eCtx.Settings.ConfigPath)
	e2eCtx.E2EConfig = LoadE2EConfig(e2eCtx.Settings.ConfigPath)

	Expect(e2eCtx.E2EConfig.GetVariable(LXCSecretName)).ToNot(BeEmpty(), "Invalid test suite argument. Value of environment variable LXC_SECRET_NAME should be set")

	Logf("Creating a clusterctl local repository into %q", e2eCtx.Settings.ArtifactFolder)
	e2eCtx.Environment.ClusterctlConfigPath = createClusterctlLocalRepository(e2eCtx.E2EConfig, filepath.Join(e2eCtx.Settings.ArtifactFolder, "repository"))

	// Ensuring a CNI file is defined in the config and update settings
	Logf("Loading CNI manifest")
	e2eCtx.Settings.CNIManifest = loadCNIManifest(e2eCtx.E2EConfig)

	Logf("Setting up the bootstrap cluster")
	e2eCtx.Environment.BootstrapClusterProvider, e2eCtx.Environment.BootstrapClusterProxy = setupBootstrapCluster(e2eCtx.E2EConfig, e2eCtx.Environment.Scheme, e2eCtx.Settings.UseExistingCluster)

	Logf("Initializing the bootstrap cluster")
	initBootstrapCluster(e2eCtx)

	Logf("Ensuring infrastructure credentials can be used")
	ensureLXCClientOptions(e2eCtx)

	conf := synchronizedBeforeTestSuiteConfig{
		E2EConfig:            *e2eCtx.E2EConfig,
		ArtifactFolder:       e2eCtx.Settings.ArtifactFolder,
		ConfigPath:           e2eCtx.Settings.ConfigPath,
		ClusterctlConfigPath: e2eCtx.Environment.ClusterctlConfigPath,
		KubeconfigPath:       e2eCtx.Environment.BootstrapClusterProxy.GetKubeconfigPath(),

		LXCClientOptions: e2eCtx.Settings.LXCClientOptions,
		CNIManifest:      e2eCtx.Settings.CNIManifest,
	}

	data, err := yaml.Marshal(conf)
	Expect(err).NotTo(HaveOccurred())
	return data
}

// AllNodesBeforeSuite is the common setup down on each ginkgo parallel node before the test suite runs.
func AllNodesBeforeSuite(e2eCtx *E2EContext, data []byte) {
	conf := &synchronizedBeforeTestSuiteConfig{}
	Expect(yaml.UnmarshalStrict(data, conf)).To(Succeed())

	e2eCtx.Settings.ArtifactFolder = conf.ArtifactFolder
	e2eCtx.Settings.ConfigPath = conf.ConfigPath
	e2eCtx.Environment.ClusterctlConfigPath = conf.ClusterctlConfigPath
	withLogCollector := framework.WithMachineLogCollector(IncusLogCollector{E2EContext: e2eCtx})
	e2eCtx.Environment.BootstrapClusterProxy = framework.NewClusterProxy("bootstrap", conf.KubeconfigPath, e2eCtx.Environment.Scheme, withLogCollector)
	e2eCtx.E2EConfig = &conf.E2EConfig
	e2eCtx.Settings.LXCClientOptions = conf.LXCClientOptions
	e2eCtx.Settings.CNIManifest = conf.CNIManifest
}

// AllNodesAfterSuite is cleanup that runs on all ginkgo parallel nodes after the test suite finishes.
func AllNodesAfterSuite(e2eCtx *E2EContext) {
}

// Node1AfterSuite is cleanup that runs on the first ginkgo node after the test suite finishes.
func Node1AfterSuite(ctx context.Context, e2eCtx *E2EContext) {
	Logf("Tearing down the management cluster")
	if !e2eCtx.Settings.SkipCleanup {
		tearDown(e2eCtx.Environment.BootstrapClusterProvider, e2eCtx.Environment.BootstrapClusterProxy)
	}
}
