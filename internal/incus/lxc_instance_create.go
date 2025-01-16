package incus

import (
	"context"
	"fmt"
	"slices"
	"strings"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/neoaggelos/cluster-api-provider-lxc/api/v1alpha1"
)

// CreateInstance creates the LXC instance based on configuration from the machine.
func (c *Client) CreateInstance(ctx context.Context, machine *clusterv1.Machine, lxcMachine *infrav1.LXCMachine, lxcCluster *infrav1.LXCCluster, cloudInit string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, instanceCreateTimeout)
	defer cancel()

	name := lxcMachine.GetInstanceName()

	role := "control-plane"
	if !util.IsControlPlaneMachine(machine) {
		role = "worker"
	}
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", name, "role", role))

	instanceType := c.instanceTypeFromAPI(lxcMachine.Spec.InstanceType)

	profiles := lxcMachine.Spec.Profiles
	if instanceType == api.InstanceTypeContainer && !lxcCluster.Spec.SkipDefaultKubeadmProfile && !slices.Contains(lxcMachine.Spec.Profiles, lxcCluster.GetProfileName()) {
		// for containers, include the default kubeadm profile
		profiles = append(lxcMachine.Spec.Profiles, lxcCluster.GetProfileName())
	}
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("profiles", profiles))

	image := lxcMachine.Spec.Image

	// Incus and LXD have diverged image servers for Ubuntu images, making it easy to confuse users.
	// To address the issue, we allow a special prefix `ubuntu:VERSION` for image names:
	if strings.HasPrefix(image.Name, "ubuntu:") {
		server, _, err := c.Client.GetServer()
		if err != nil {
			log.FromContext(ctx).Error(err, "Failed to GetServer")
		} else {
			switch server.Environment.Server {
			case "incus":
				image = infrav1.LXCMachineImageSource{
					Name:     fmt.Sprintf("ubuntu/%s/cloud", strings.TrimPrefix(image.Name, "ubuntu:")),
					Server:   "https://images.linuxcontainers.org",
					Protocol: "simplestreams",
				}
				log.FromContext(ctx).V(2).WithValues("image", image).Info("Using Ubuntu image from https://images.linuxcontainers.org")
			case "lxd":
				image = infrav1.LXCMachineImageSource{
					Name:     strings.TrimPrefix(image.Name, "ubuntu:"),
					Server:   "https://cloud-images.ubuntu.com/releases/",
					Protocol: "simplestreams",
				}
				log.FromContext(ctx).V(2).WithValues("image", image).Info("Using Ubuntu image from https://cloud-images.ubuntu.com/releases/")
			default:
				return nil, terminalError{fmt.Errorf("image name is %q, but server is %q. Images with 'ubuntu:' prefix are only allowed for Incus and LXD", image.Name, server.Environment.Server)}
			}
		}
	}
	if image.IsZero() {
		if machine.Spec.Version == nil {
			return nil, terminalError{fmt.Errorf("no image source specified on LXCMachineTemplate and Machine %q does not have a Kubernetes version", machine.Name)}
		}

		version := *machine.Spec.Version

		// test if image for version exists on the default simplestreams server, fail otherwise.
		if ssClient, err := incus.ConnectSimpleStreams(defaultSimplestreamsServer, &incus.ConnectionArgs{}); err != nil {
			return nil, fmt.Errorf("no image source specified and failed to connect to simplestreams server %q: %w", defaultSimplestreamsServer, err)
		} else if _, _, err := ssClient.GetImageAliasType(string(instanceType), fmt.Sprintf("kubeadm/%s", version)); err != nil {
			return nil, terminalError{fmt.Errorf("no image source specified and simplestreams server %q does not provide images for Kubernetes version %q: %w. Please consider using a different Kubernetes version, or build your own base image and set the image source on the LXCMachineTemplate resource", defaultSimplestreamsServer, version, err)}
		}

		image = infrav1.LXCMachineImageSource{
			Name:     fmt.Sprintf("kubeadm/%s", version),
			Server:   defaultSimplestreamsServer,
			Protocol: "simplestreams",
		}
	}
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("image", image))

	if err := c.createInstanceIfNotExists(ctx, api.InstancesPost{
		Name:         name,
		Type:         c.instanceTypeFromAPI(lxcMachine.Spec.InstanceType),
		Source:       c.instanceSourceFromAPI(image),
		InstanceType: lxcMachine.Spec.Flavor,
		InstancePut: api.InstancePut{
			Profiles: profiles,
			Config: map[string]string{
				configClusterNameKey:      lxcCluster.Name,
				configClusterNamespaceKey: lxcCluster.Namespace,
				configInstanceRoleKey:     role,
				configCloudInitKey:        cloudInit,
			},
		},
	}); err != nil {

		// TODO: Handle the below situation as a terminalError.
		//
		// E1230 21:42:45.170291 1388422 controller.go:316] "Reconciler error" err="failed to create instance: failed to ensure instance exists: failed to wait for CreateInstance operation: Requested image's type \"container\" doesn't match instance type \"virtual-machine\"" controller="lxcmachine" controllerGroup="infrastructure.cluster.x-k8s.io" controllerKind="LXCMachine" LXCMachine="default/c1-control-plane-kprl9" namespace="default" name="c1-control-plane-kprl9" reconcileID="d40dfec7-ce45-4585-9a1e-5974efbeb925"
		return nil, fmt.Errorf("failed to ensure instance exists: %w", err)
	}

	if err := c.ensureInstanceRunning(ctx, name); err != nil {
		return nil, fmt.Errorf("failed to ensure loadbalancer instance is running: %w", err)
	}

	addrs, err := c.waitForInstanceAddress(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get loadbalancer instance address: %w", err)
	}
	return addrs, nil
}
