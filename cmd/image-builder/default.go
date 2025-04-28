package main

var (
	defaultUbuntuVersion = "24.04"

	defaultInstanceName     = "capn-builder"
	defaultInstanceType     = "container"
	defaultInstanceProfiles = []string{"default"}

	defaultPullExtraImages = []string{
		"docker.io/flannel/flannel-cni-plugin:v1.6.0-flannel1",
		"docker.io/flannel/flannel:v0.26.3",
	}
)
