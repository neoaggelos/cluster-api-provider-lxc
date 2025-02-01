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
	"fmt"
	"os"
	"time"

	"sigs.k8s.io/cluster-api/test/framework/clusterctl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Logf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(GinkgoWriter, "["+time.Now().Format(time.RFC3339)+"] "+format+"\n", a...)
}

// LoadE2EConfig loads the e2econfig from the specified path.
func LoadE2EConfig(configPath string) *clusterctl.E2EConfig {
	config := clusterctl.LoadE2EConfig(context.TODO(), clusterctl.LoadE2EConfigInput{ConfigPath: configPath})
	Expect(config).ToNot(BeNil(), "Failed to load E2E config from %s", configPath)
	return config
}

// SetEnvVar sets an environment variable in the process. If marked private,
// the value is not printed.
func SetEnvVar(key, value string, private bool) {
	printableValue := "*******"
	if !private {
		printableValue = value
	}

	Logf("Setting environment variable: key=%s, value=%s", key, printableValue)
	_ = os.Setenv(key, value)
}
