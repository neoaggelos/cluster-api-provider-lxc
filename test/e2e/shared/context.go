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
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/test/framework"
	"sigs.k8s.io/cluster-api/test/framework/bootstrap"
	"sigs.k8s.io/cluster-api/test/framework/clusterctl"

	"github.com/lxc/cluster-api-provider-incus/internal/incus"
)

// Option represents an option to use when creating a e2e context.
type Option func(*E2EContext)

func NewE2EContext(options ...Option) *E2EContext {
	ctx := &E2EContext{}
	ctx.Environment.Scheme = DefaultScheme()

	for _, opt := range options {
		opt(ctx)
	}

	return ctx
}

// E2EContext represents the context of the e2e test.
type E2EContext struct {
	// Settings is the settings used for the test
	Settings Settings
	// E2EConfig to be used for this test, read from configPath.
	E2EConfig *clusterctl.E2EConfig
	// Environment represents the runtime environment
	Environment RuntimeEnvironment
}

// Settings represents the test settings.
type Settings struct {
	// ConfigPath is the path to the e2e config file.
	ConfigPath string
	// useExistingCluster instructs the test to use the current cluster instead of creating a new one (default discovery rules apply).
	UseExistingCluster bool
	// ArtifactFolder is the folder to store e2e test artifacts.
	ArtifactFolder string
	// DataFolder is the root folder for the data required by the tests
	DataFolder string
	// SkipCleanup prevents cleanup of test resources e.g. for debug purposes.
	SkipCleanup bool
	// LXCClientOptions is infrastructure credentials
	LXCClientOptions incus.Options
	// CNIManifest is the raw CNI manifest that will be deployed in workload clusters
	CNIManifest string
}

// RuntimeEnvironment represents the runtime environment of the test.
type RuntimeEnvironment struct {
	// BootstrapClusterProvider manages provisioning of the bootstrap cluster to be used for the e2e tests.
	// Please note that provisioning will be skipped if use-existing-cluster is provided.
	BootstrapClusterProvider bootstrap.ClusterProvider
	// BootstrapClusterProxy allows to interact with the bootstrap cluster to be used for the e2e tests.
	BootstrapClusterProxy framework.ClusterProxy
	// ClusterctlConfigPath to be used for this test, created by generating a clusterctl local repository
	// with the providers specified in the configPath.
	ClusterctlConfigPath string
	// Scheme is the GVK scheme to use for the tests
	Scheme *runtime.Scheme
}

// DefaultPostNamespaceCreated deploys the LXC credentials secret, as well as the CNI resource set on the namespace.
func (c *E2EContext) DefaultPostNamespaceCreated() func(framework.ClusterProxy, string) {
	return func(managementClusterProxy framework.ClusterProxy, workloadClusterNamespace string) {
		FixupNamespace(c, workloadClusterNamespace, true, true)
	}
}
