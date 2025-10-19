package instances

import (
	"fmt"

	"github.com/lxc/incus/v6/shared/api"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/static"
)

type KubeadmLaunchOptionsInput struct {
	KubernetesVersion string

	InstanceType             api.InstanceType
	Privileged               bool
	SkipProfile              bool
	SkipInstallKubeadmScript bool

	ServerName string

	CloudInit string
}

// KubeadmLaunchOptions launches kubeadm nodes.
func KubeadmLaunchOptions(in KubeadmLaunchOptionsInput) *lxc.LaunchOptions {
	opts := (&lxc.LaunchOptions{}).
		WithInstanceType(in.InstanceType).
		WithImage(lxc.CapnImage(fmt.Sprintf("kubeadm/%s", in.KubernetesVersion)))

	if !in.SkipInstallKubeadmScript {
		opts = opts.WithInstanceTemplates(map[string]string{
			"/opt/cluster-api/install-kubeadm.sh": static.InstallKubeadmScript(),
		})
	}

	// add cloud-init
	if len(in.CloudInit) > 0 {
		opts = opts.WithConfig(map[string]string{
			"cloud-init.user-data": in.CloudInit,
		})
	}

	// apply profile for Kubernetes to run in LXC containers
	if in.InstanceType == api.InstanceTypeContainer && !in.SkipProfile {
		profile := static.DefaultKubeadmProfile(in.Privileged, in.ServerName)
		opts = opts.
			WithConfig(profile.Config).
			WithDevices(profile.Devices)
	}

	return opts
}
