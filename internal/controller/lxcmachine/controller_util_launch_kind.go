package lxcmachine

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/lxc/incus/v6/shared/api"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/cloudinit"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/static"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

func launchKindInstance(
	ctx context.Context,
	lxcClient *lxc.Client,
	clusterName string,
	clusterNamespace string,
	machineName string,

	name string,
	isControlPlane bool,
	cloudInit string,

	spec infrav1.LXCMachineSpec,
	version *string,

	manualCloudInit bool,
) ([]string, error) {
	if err := lxcClient.SupportsInstanceOCI(); err != nil {
		return nil, utils.TerminalError(fmt.Errorf("cannot launch kind instance as OCI containers are not supported: %w", err))
	}

	role := "control-plane"
	if !isControlPlane {
		role = "worker"
	}
	instanceType := lxc.Container

	// Parse device configurations
	devices := map[string]map[string]string{}
	for _, deviceSpec := range spec.Devices {
		deviceName, deviceArgs, hasSeparator := strings.Cut(deviceSpec, ",")
		if !hasSeparator {
			return nil, utils.TerminalError(fmt.Errorf("device spec %q is not using the expected %q format", deviceSpec, "<device>,<key>=<value>,<key2>=<value2>"))
		}

		if _, ok := devices[deviceName]; !ok {
			devices[deviceName] = map[string]string{}
		}

		for _, deviceArg := range strings.Split(deviceArgs, ",") {
			key, value, hasEqual := strings.Cut(deviceArg, "=")
			if !hasEqual {
				return nil, utils.TerminalError(fmt.Errorf("device argument %q of device spec %q is not using the expected %q format", deviceArg, deviceSpec, "<key>=<value>"))
			}

			devices[deviceName][key] = value
		}
	}

	var image api.InstanceSource
	switch {
	case spec.Image.IsZero():
		if version == nil {
			return nil, utils.TerminalError(fmt.Errorf("no image source specified on LXCMachineTemplate and Machine %q does not have a Kubernetes version", machineName))
		}

		version := *version

		// test if kindest/node image for this version exists on DockerHub, fail otherwise.
		if _, err := crane.Head(fmt.Sprintf("docker.io/kindest/node:%s", version)); err != nil {
			// example errors:
			// HEAD https://index.docker.io/v2/kindest/node/manifests/v1.34.0-not-exist: unexpected status code 404 Not Found (HEAD responses have no body, use GET for details)
			// HEAD https://index.docker.io/v2/kindest/node13131/manifests/v1.33.0: unexpected status code 401 Unauthorized (HEAD responses have no body, use GET for details)
			// HEAD http://w00:5050/v2/kindest/node13131/manifests/v1.33.0: unexpected status code 404 Not Found (HEAD responses have no body, use GET for details)
			if strings.Contains(err.Error(), "unexpected status code 4") {
				return nil, utils.TerminalError(fmt.Errorf("no image source specified and could not find kindest/node:%s image on DockerHub: %w. Please consider using a different Kubernetes version, or build your own base image and set the image source on the LXCMachineTemplate resource", version, err))
			} else {
				return nil, fmt.Errorf("no image source specified and failed to connect to DockerHub: %w", err)
			}
		}

		image = api.InstanceSource{
			Type:     "image",
			Protocol: "oci",
			Server:   "https://docker.io",
			Alias:    fmt.Sprintf("kindest/node:%s", version),
		}
	default:
		image = api.InstanceSource{
			Type:        "image",
			Protocol:    spec.Image.Protocol,
			Server:      spec.Image.Server,
			Alias:       spec.Image.Name,
			Fingerprint: spec.Image.Fingerprint,
		}
	}

	instance := api.InstancesPost{
		Name:         name,
		Type:         api.InstanceType(instanceType),
		Source:       image,
		InstanceType: spec.Flavor,
		InstancePut: api.InstancePut{
			Config:   map[string]string{},
			Profiles: spec.Profiles,
			Devices:  devices,
		},
	}

	// apply custom config
	maps.Copy(instance.Config, spec.Config)
	maps.Copy(instance.Config, map[string]string{
		"user.cluster-name":      clusterName,
		"user.cluster-namespace": clusterNamespace,
		"user.machine-name":      name,
		"user.cluster-role":      role,
		"cloud-init.user-data":   cloudInit,
	})

	// apply profile for Kubernetes to run in LXC containers
	profile := static.DefaultKindProfile()
	maps.Copy(instance.Devices, profile.Devices)
	maps.Copy(instance.Config, profile.Config)

	seedFiles := maps.Clone(defaultKindSeedFiles)

	// configure cloud-init
	if manualCloudInit {
		// manual cloud-init mode:
		// - parse YAML (ensure no unknown fields are present), and replace {{ v1.local_hostname }} with hostname
		// - marshal to JSON
		// - embed to instance at /hack/cloud-init.json
		// - instance will run using the kind-cloud-init.py script (see internal/embed/kind-cloud-init.py)
		cloudConfig, err := cloudinit.Parse(cloudInit, strings.NewReplacer(
			"{{ v1.local_hostname }}", name,
		))
		if err != nil {
			return nil, utils.TerminalError(fmt.Errorf("failed to parse instance cloud-config, please report this bug to https://github.com/lxc/cluster-api-provider-incus/issues: %w", err))
		}

		b, err := json.Marshal(cloudConfig)
		if err != nil {
			return nil, utils.TerminalError(fmt.Errorf("failed to generate JSON cloud-config for instance, please report this bug to github.com/lxc/cluster-api-provider-incus/issues: %w", err))
		}

		seedFiles["/hack/cloud-init.json"] = string(b)
	}

	addrs, err := lxcClient.WithTarget(spec.Target).WaitForLaunchInstance(ctx, instance, &lxc.LaunchOptions{SeedFiles: seedFiles, Symlinks: defaultKindSymlinks})
	if err != nil {
		return nil, fmt.Errorf("failed to launch instance: %w", err)
	}

	return addrs, nil
}

// defaultKindSeedFiles that are injected to LXCMachine kind instances.
var defaultKindSeedFiles = map[string]string{
	"/kind/product_name":                            "kind",
	"/kind/product_uuid":                            "kind",
	"/var/lib/cloud/seed/nocloud-net/meta-data":     static.CloudInitMetaDataTemplate(),
	"/var/lib/cloud/seed/nocloud-net/user-data":     static.CloudInitUserDataTemplate(),
	"/hack/cloud-init.py":                           static.KindCloudInitScript(),
	"/etc/systemd/system/cloud-init-launch.service": static.CloudInitLaunchSystemdServiceTemplate(),
}

// defaultKindSymlinks that are injected to LXCMachine kind instances.
var defaultKindSymlinks = map[string]string{
	"/etc/systemd/system/multi-user.target.wants/cloud-init-launch.service": "/etc/systemd/system/cloud-init-launch.service",
}
