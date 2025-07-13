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
	"errors"
	"flag"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/cluster-api/test/framework"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
)

const (
	KubeContext       = "KUBE_CONTEXT"
	KubernetesVersion = "KUBERNETES_VERSION"

	// Load LXC server credentials from local config file
	LXCLoadConfigFile = "LXC_LOAD_CONFIG_FILE"
	LXCLoadRemoteName = "LXC_LOAD_REMOTE_NAME"

	// Name of secret for LXC credentials
	LXCSecretName = "LXC_SECRET_NAME"

	// KubeVIP address to use for kube-vip tests
	KubeVIPAddress = "KUBE_VIP_ADDRESS"

	// Tarball with provider images to load into instances for SelfHosted tests
	PreloadImages = "PRELOAD_IMAGES"

	FlavorDefault     = ""
	FlavorDevelopment = "development"
	FlavorAutoscaler  = "autoscaler"
)

// DefaultScheme returns the default scheme to use for testing.
func DefaultScheme() *runtime.Scheme {
	sc := runtime.NewScheme()
	framework.TryAddDefaultSchemes(sc)

	err := errors.Join(
		infrav1.AddToScheme(sc),
		clientgoscheme.AddToScheme(sc),
	)
	if err != nil {
		panic("error adding types to scheme: " + err.Error())
	}
	return sc
}

// CreateDefaultFlags will create the default flags used for the tests and binds them to the e2e context.
func CreateDefaultFlags(ctx *E2EContext) {
	flag.StringVar(&ctx.Settings.ConfigPath, "config-path", "", "path to the e2e config file")
	flag.StringVar(&ctx.Settings.ArtifactFolder, "artifacts-folder", "", "folder where e2e test artifact should be stored")
	flag.BoolVar(&ctx.Settings.UseExistingCluster, "use-existing-cluster", false, "if true, the test will try to use an existing cluster and fallback to create a new one if it couldn't be found")
	flag.BoolVar(&ctx.Settings.SkipCleanup, "skip-cleanup", false, "if true, the resource cleanup after tests will be skipped")
	flag.StringVar(&ctx.Settings.DataFolder, "data-folder", "", "path to the data folder")
}
