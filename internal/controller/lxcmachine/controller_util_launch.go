package lxcmachine

import (
	"context"
	"fmt"
	"strings"

	"github.com/lxc/incus/v6/shared/api"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util"

	infrav1 "github.com/lxc/cluster-api-provider-incus/api/v1alpha2"
	"github.com/lxc/cluster-api-provider-incus/internal/instances"
	"github.com/lxc/cluster-api-provider-incus/internal/loadbalancer"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/ptr"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

func launchInstance(ctx context.Context, cluster *clusterv1.Cluster, lxcCluster *infrav1.LXCCluster, machine *clusterv1.Machine, lxcMachine *infrav1.LXCMachine, lxcClient *lxc.Client, cloudInit string) ([]string, error) {
	// TODO: merge the two code paths as much as possible
	if lxcMachine.Spec.InstanceType == "kind" {
		return launchKindInstance(ctx, cluster, lxcCluster, machine, lxcMachine, lxcClient, cloudInit)
	}

	role := "control-plane"
	if !util.IsControlPlaneMachine(machine) {
		role = "worker"
	}
	instanceType := api.InstanceTypeContainer
	if lxcMachine.Spec.InstanceType != "" {
		instanceType = api.InstanceType(lxcMachine.Spec.InstanceType)
	}

	// Parse device configurations
	devices, err := lxcMachine.Spec.Devices.ToMap()
	if err != nil {
		return nil, utils.TerminalError(fmt.Errorf("invalid .spec.devices on LXCMachine: %w", err))
	}

	machineVersion := machine.Spec.Version

	imageSpec := lxcMachine.Spec.Image.DeepCopy()
	if strings.Contains(imageSpec.Name, "VERSION") {
		if machineVersion == "" {
			return nil, utils.TerminalError(fmt.Errorf("image name %q contains VERSION but Machine %q does not have a Kubernetes version", imageSpec.Name, machine.Name))
		}
		imageSpec.Name = strings.ReplaceAll(imageSpec.Name, "VERSION", machineVersion)
	}

	var image lxc.ImageFamily = lxc.Image{
		Protocol:    imageSpec.Protocol,
		Server:      imageSpec.Server,
		Fingerprint: imageSpec.Fingerprint,
		Alias:       imageSpec.Name,
	}
	if imageSpec.Name != "" {
		parsed, isParsed, err := lxc.ParseImage(imageSpec.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to parse image %q: %w", imageSpec.Name, err)
		} else if isParsed {
			image = parsed
		}
	} else if imageSpec.IsZero() {
		if machineVersion == "" {
			return nil, utils.TerminalError(fmt.Errorf("no image source specified on LXCMachineTemplate and Machine %q does not have a Kubernetes version", machine.Name))
		}
		kubeadmImage := lxc.CapnImage(fmt.Sprintf("kubeadm/%s", machineVersion))
		if err := kubeadmImage.Check(instanceType); err != nil {
			if utils.IsTerminalError(err) {
				err = fmt.Errorf("image not specified and default simplestreams server does not provide images for Kubernetes version %q. The error was: %w. Please consider using a different Kubernetes version, or build your own base image and set the image source on the LXCMachineTemplate resource", machineVersion, err)
			}
			return nil, err
		}
		image = kubeadmImage
	}

	launchOpts := instances.KubeadmLaunchOptions(instances.KubeadmLaunchOptionsInput{
		InstanceType:      instanceType,
		KubernetesVersion: machineVersion,
		Privileged:        !lxcCluster.Spec.Unprivileged,
		SkipProfile:       lxcCluster.Spec.SkipDefaultKubeadmProfile,
		ServerName:        lxcClient.GetServerName(),

		CloudInit: cloudInit,
	}).
		WithFlavor(lxcMachine.Spec.Flavor).
		WithProfiles(lxcMachine.Spec.Profiles).
		WithDevices(devices).
		WithConfig(lxcMachine.Spec.Config).
		WithConfig(map[string]string{
			"user.cluster-name":      cluster.Name,
			"user.cluster-namespace": cluster.Namespace,
			"user.machine-name":      machine.Name,
			"user.cluster-role":      role,
		}).
		WithImage(image)

	// apply instance templates from load balancer manager
	if util.IsControlPlaneMachine(machine) {
		if files, err := loadbalancer.ManagerForCluster(cluster, lxcCluster, lxcClient).ControlPlaneInstanceTemplates(ptr.Deref(cluster.Status.Initialization.ControlPlaneInitialized, false)); err != nil {
			return nil, fmt.Errorf("failed to generate load balancer configuration files: %w", err)
		} else {
			launchOpts = launchOpts.WithInstanceTemplates(files)
		}
	}

	return lxcClient.WithTarget(lxcMachine.Spec.Target).WaitForLaunchInstance(ctx, lxcMachine.GetInstanceName(), launchOpts)
}
