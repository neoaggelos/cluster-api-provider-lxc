package profile

import (
	_ "embed"

	"github.com/lxc/incus/v6/shared/api"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed embed/kubeadm.yaml
	defaultKubeadmYAML []byte

	// DefaultKubeadm is the default kubeadm profile to use with LXC nodes.
	DefaultKubeadm api.ProfilePut
)

func init() {
	DefaultKubeadm = mustParseProfile(defaultKubeadmYAML)
}

func mustParseProfile(b []byte) api.ProfilePut {
	var profile api.ProfilePut
	if err := yaml.Unmarshal(b, &profile); err != nil {
		panic(err)
	}
	return profile
}
