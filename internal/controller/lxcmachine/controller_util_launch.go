package lxcmachine

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"time"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/static"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

func launchInstance(ctx context.Context, cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, machine *clusterv1.Machine, lxcMachine *infrav1.LXCMachine, lxcClient *lxc.Client, cloudInit string) ([]string, error) {
	// TODO: merge the two code paths as much as possible
	if lxcMachine.Spec.InstanceType == "kind" {
		return launchKindInstance(ctx, lxcClient, cluster.Name, cluster.Namespace, machine.Name, lxcMachine.GetInstanceName(), util.IsControlPlaneMachine(machine), cloudInit, lxcMachine.Spec, machine.Spec.Version, true)
	}

	name := lxcMachine.GetInstanceName()

	role := "control-plane"
	if !util.IsControlPlaneMachine(machine) {
		role = "worker"
	}
	instanceType := lxc.Container
	if lxcMachine.Spec.InstanceType != "" {
		instanceType = lxcMachine.Spec.InstanceType
	}

	// Parse device configurations
	devices := map[string]map[string]string{}
	for _, deviceSpec := range lxcMachine.Spec.Devices {
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
	case strings.HasPrefix(lxcMachine.Spec.Image.Name, "ubuntu:"):
		ubuntuImage, isUbuntuImage, err := lxcClient.GetDefaultUbuntuImage(ctx, lxcMachine.Spec.Image.Name)
		if err != nil {
			return nil, err
		} else if isUbuntuImage {
			image = ubuntuImage
		}
	case lxcMachine.Spec.Image.IsZero():
		if machine.Spec.Version == nil {
			return nil, utils.TerminalError(fmt.Errorf("no image source specified on LXCMachineTemplate and Machine %q does not have a Kubernetes version", machine.Name))
		}

		version := *machine.Spec.Version

		// test if image for version exists on the default simplestreams server, fail otherwise.
		if ssClient, err := incus.ConnectSimpleStreams(lxc.DefaultSimplestreamsServer, &incus.ConnectionArgs{HTTPClient: &http.Client{Timeout: 10 * time.Second}}); err != nil {
			return nil, fmt.Errorf("no image source specified and failed to connect to simplestreams server %q: %w", lxc.DefaultSimplestreamsServer, err)
		} else if _, _, err := ssClient.GetImageAliasType(instanceType, fmt.Sprintf("kubeadm/%s", version)); err != nil {
			return nil, utils.TerminalError(fmt.Errorf("no image source specified and simplestreams server %q does not provide images for Kubernetes version %q: %w. Please consider using a different Kubernetes version, or build your own base image and set the image source on the LXCMachineTemplate resource", lxc.DefaultSimplestreamsServer, version, err))
		}

		image = api.InstanceSource{
			Type:     "image",
			Protocol: "simplestreams",
			Server:   lxc.DefaultSimplestreamsServer,
			Alias:    fmt.Sprintf("kubeadm/%s", version),
		}
	default:
		image = api.InstanceSource{
			Type:        "image",
			Protocol:    lxcMachine.Spec.Image.Protocol,
			Server:      lxcMachine.Spec.Image.Server,
			Alias:       lxcMachine.Spec.Image.Name,
			Fingerprint: lxcMachine.Spec.Image.Fingerprint,
		}
	}

	instance := api.InstancesPost{
		Name:         name,
		Type:         api.InstanceType(instanceType),
		Source:       image,
		InstanceType: lxcMachine.Spec.Flavor,
		InstancePut: api.InstancePut{
			Config:   map[string]string{},
			Profiles: lxcMachine.Spec.Profiles,
			Devices:  devices,
		},
	}

	// apply instance config
	maps.Copy(instance.Config, lxcMachine.Spec.Config)
	maps.Copy(instance.Config, map[string]string{
		"user.cluster-name":      cluster.Name,
		"user.cluster-namespace": cluster.Namespace,
		"user.machine-name":      machine.Name,
		"user.cluster-role":      role,
		"cloud-init.user-data":   cloudInit,
	})

	// apply profile for Kubernetes to run in LXC containers
	if instanceType == lxc.Container && !lxcCluster.Spec.SkipDefaultKubeadmProfile {
		profile := static.DefaultKubeadmProfile(!lxcCluster.Spec.Unprivileged, lxcClient.GetServerName(ctx))

		maps.Copy(instance.Devices, profile.Devices)
		maps.Copy(instance.Config, profile.Config)
	}

	return lxcClient.WithTarget(lxcMachine.Spec.Target).WaitForLaunchInstance(ctx, instance, &lxc.LaunchOptions{SeedFiles: defaultSeedFiles})
}

// defaultSeedFiles that are injected to LXCMachine instances.
var defaultSeedFiles = map[string]string{
	"/opt/cluster-api/install-kubeadm.sh": static.InstallKubeadmScript(),
}
