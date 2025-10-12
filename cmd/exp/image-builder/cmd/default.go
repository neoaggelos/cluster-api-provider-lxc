package cmd

import (
	"time"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

var (
	defaultBaseImage = "ubuntu:24.04"

	defaultInstanceName           = "capn-builder"
	defaultInstanceType           = lxc.Container
	defaultInstanceProfiles       = []string{"default"}
	defaultValidationInstanceName = "capn-validator"

	defaultInstanceStopGracePeriod = 2 * time.Minute

	defaultPullExtraImages = []string{
		// images for default flannel CNI
		// NOTE(neoaggelos): keep up to date with flannel CNI manifest in ./templates/cluster-template.yaml
		"ghcr.io/flannel-io/flannel-cni-plugin:v1.7.1-flannel1",
		"ghcr.io/flannel-io/flannel:v0.27.3",
	}
)
