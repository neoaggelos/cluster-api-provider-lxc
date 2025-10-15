package instances

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lxc/incus/v6/shared/api"

	"github.com/lxc/cluster-api-provider-incus/internal/cloudinit"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/static"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

type KindLaunchOptionsInput struct {
	KubernetesVersion string

	Privileged  bool
	SkipProfile bool

	PodNetworkCIDR string

	CloudInit           string
	CloudInitAptInstall bool
}

// KindLaunchOptions launches kindest/node nodes.
func KindLaunchOptions(in KindLaunchOptionsInput) (*lxc.LaunchOptions, error) {
	opts := (&lxc.LaunchOptions{}).
		WithInstanceType(api.InstanceTypeContainer).
		WithImage(lxc.KindestNodeImage(in.KubernetesVersion)).
		WithReplacements(map[string]map[string]string{
			"/usr/local/bin/entrypoint": {
				// Incus unprivileged containers cannot edit /etc/resolv.conf, so do not let the entrypoint attempt it.
				">/etc/resolv.conf": ">/etc/local-resolv.conf",
			},
		}).
		WithSymlinks(map[string]string{
			// Incus will inject its own PID 1 init process unless the entrypoint is one of "/init", "/sbin/init", "/s6-init".
			"/init": "/usr/local/bin/entrypoint",
		})

	// add cloud-init configuration as nocloud-net datasource in the instance
	if len(in.CloudInit) > 0 {
		opts = opts.
			WithConfig(map[string]string{
				"cloud-init.user-data": in.CloudInit,
			}).
			WithInstanceTemplates(map[string]string{
				// inject cloud-init into instance.
				"/var/lib/cloud/seed/nocloud-net/meta-data": static.CloudInitMetaDataTemplate(),
				"/var/lib/cloud/seed/nocloud-net/user-data": static.CloudInitUserDataTemplate(),
				// cloud-init-launch.service is used to start the cloud-init scripts.
				"/etc/systemd/system/cloud-init-launch.service": static.CloudInitLaunchSystemdServiceTemplate(),
				"/hack/cloud-init.py":                           static.KindCloudInitScript(),
			}).
			WithSymlinks(map[string]string{
				// enable the cloud-init-launch service.
				"/etc/systemd/system/multi-user.target.wants/cloud-init-launch.service": "/etc/systemd/system/cloud-init-launch.service",
			})

		if !in.CloudInitAptInstall {
			// manual cloud-init mode:
			// - parse YAML (ensure no unknown fields are present), and replace "{{ v1.local_hostname }}" with "{{ container.name }}", which is a pango template that will resolve to the instance hostname upon launch
			// - marshal to JSON
			// - embed to instance at /hack/cloud-init.json
			// - instance will run using the kind-cloud-init.py script (see internal/embed/kind-cloud-init.py)
			cloudConfig, err := cloudinit.Parse(in.CloudInit, strings.NewReplacer(
				"{{ v1.local_hostname }}", "{{ container.name }}",
			))
			if err != nil {
				return nil, utils.TerminalError(fmt.Errorf("failed to parse instance cloud-config, please report this bug to https://github.com/lxc/cluster-api-provider-incus/issues: %w", err))
			}

			b, err := json.Marshal(cloudConfig)
			if err != nil {
				return nil, utils.TerminalError(fmt.Errorf("failed to generate JSON cloud-config for instance, please report this bug to github.com/lxc/cluster-api-provider-incus/issues: %w", err))
			}

			opts = opts.WithInstanceTemplates(map[string]string{
				"/hack/cloud-init.json": string(b),
			})
		}
	}

	// pod network CIDR
	if len(in.PodNetworkCIDR) > 0 {
		opts = opts.WithReplacements(map[string]map[string]string{
			"/kind/manifests/default-cni.yaml": {
				"{{ .PodSubnet }}": in.PodNetworkCIDR,
			},
		})
	}

	// apply profile for Kubernetes to run in LXC containers
	if !in.SkipProfile {
		profile := static.DefaultKindProfile(in.Privileged)
		opts = opts.
			WithConfig(profile.Config).
			WithDevices(profile.Devices)
	}

	return opts, nil
}
