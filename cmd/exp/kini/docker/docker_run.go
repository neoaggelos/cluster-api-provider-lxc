package docker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/lxc/incus/v6/shared/api"
	"github.com/spf13/cobra"

	"github.com/lxc/cluster-api-provider-incus/internal/instances"
	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/cluster-api-provider-incus/internal/static"
	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

// launchOptionsForImage initializes LaunchOptions for node or haproxy instances.
func launchOptionsForImage(ctx context.Context, rawImage string, env Environment, serverName string) (*lxc.LaunchOptions, error) {
	image, err := utils.ParseOCIImage(rawImage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q: %w", rawImage, err)
	}

	// handle haproxy instances
	if strings.Contains(image.Alias(), "kindest/haproxy") {
		if !env.KindInstances(ctx) {
			return nil, fmt.Errorf("haproxy instances (%q) not supported in LXC mode", rawImage)
		}
		log.V(3).Info("Launching haproxy instance", "image", rawImage)
		return instances.HaproxyOCILaunchOptions().WithImage(lxc.Image{
			Protocol: lxc.OCI,
			Server:   image.Server(),
			Alias:    image.Alias(),
		}), nil
	}

	// handle kindest/base images (kind build node-image)
	if strings.Contains(image.Alias(), "kindest/base") {
		if !env.KindInstances(ctx) {
			return nil, fmt.Errorf("kindest/base instances (%q) not supported in LXC mode", rawImage)
		}

		log.V(3).Info("Launching base instance", "image", rawImage)
		return (&lxc.LaunchOptions{}).
			WithConfig(map[string]string{
				"oci.entrypoint": "sleep infinity",
			}).
			WithImage(lxc.Image{
				Protocol: lxc.OCI,
				Server:   image.Server(),
				Alias:    image.Alias(),
			}), nil
	}

	// handle node instances (kind instances)
	if env.KindInstances(ctx) {
		log.V(3).Info("Launching node instance", "image", rawImage, "type", "kind")
		opts, err := instances.KindLaunchOptions(instances.KindLaunchOptionsInput{
			Privileged: env.Privileged(),
		})
		if err != nil {
			return nil, err
		}

		return opts.WithImage(lxc.Image{
			Protocol: lxc.OCI,
			Server:   image.Server(),
			Alias:    image.Alias(),
		}).WithUnixSocket(env.WithUnixSocket()), nil
	}

	// handle node instances (lxc instances)
	log.V(3).Info("Launching node instance", "image", rawImage, "type", "lxc")

	if digest := image.Digest(); digest != "" {
		log.Info("WARNING: Running in LXC mode, ignoring image digest", "digest", digest)
	}

	return instances.KubeadmLaunchOptions(instances.KubeadmLaunchOptionsInput{
		KubernetesVersion:        image.Tag(),
		InstanceType:             api.InstanceTypeContainer,
		Privileged:               env.Privileged(),
		ServerName:               serverName,
		SkipInstallKubeadmScript: true,
	}).WithInstanceTemplates(map[string]string{
		"/etc/sysconfig/kubelet":               "KUBELET_EXTRA_ARGS='--cgroup-root='",
		"/kind/product_name":                   "kind",
		"/kind/product_uuid":                   "kind",
		"/kind/version":                        image.Tag(),
		"/kind/manifests/default-storage.yaml": static.KindDefaultStorageManifestYAML(),
		"/kind/manifests/default-cni.yaml":     static.KubeFlannelManifestYAML(),
	}).WithUnixSocket(env.WithUnixSocket()), nil
}

// docker run --name c1-control-plane --hostname c1-control-plane --label io.x-k8s.kind.role=control-plane --privileged --security-opt seccomp=unconfined --security-opt apparmor=unconfined --tmpfs /tmp --tmpfs /run --volume /var --volume /lib/modules:/lib/modules:ro -e KIND_EXPERIMENTAL_CONTAINERD_SNAPSHOTTER --detach --tty --label io.x-k8s.kind.cluster=c1 --net kind --restart=on-failure:1 --init=false --cgroupns=private --publish=127.0.0.1:41435:6443/TCP -e KUBECONFIG=/etc/kubernetes/admin.conf kindest/node:v1.31.2@sha256:18fbefc20a7113353c7b75b5c869d7145a6abd6269154825872dc59c1329912e
// docker run --name t1-control-plane --hostname t1-control-plane --label io.x-k8s.kind.role=control-plane --privileged --security-opt seccomp=unconfined --security-opt apparmor=unconfined --tmpfs /tmp --tmpfs /run --volume /var --volume /lib/modules:/lib/modules:ro -e KIND_EXPERIMENTAL_CONTAINERD_SNAPSHOTTER --detach --tty --label io.x-k8s.kind.cluster=t1 --net kind --restart=on-failure:1 --init=false --cgroupns=private --userns=host --device /dev/fuse --publish=127.0.0.1:45295:6443/TCP -e KUBECONFIG=/etc/kubernetes/admin.conf kindest/node:v1.33.0@sha256:18fbefc20a7113353c7b75b5c869d7145a6abd6269154825872dc59c1329912e
// docker run --name test-external-load-balancer --hostname test-external-load-balancer --label io.x-k8s.kind.role=external-load-balancer --detach --tty --label io.x-k8s.kind.cluster=test --net kind --restart=on-failure:1 --init=false --cgroupns=private --publish=127.0.0.1:37715:6443/TCP docker.io/kindest/haproxy:v20230606-42a2262b
// docker run -d --entrypoint=sleep --name=kind-build-1760303841-1682602141 --platform=linux/amd64 --security-opt seccomp=unconfined docker.io/kindest/base:v20250214-acbabc1a infinity
func newDockerRunCmd(env Environment) *cobra.Command {
	var flags struct {
		// passed in command line, but will be ignored
		Init         bool
		TTY          bool
		Privileged   bool
		Detach       bool
		CgroupNS     string
		UserNS       string
		Network      string
		Restart      string
		SecurityOpts map[string]string
		Platform     string
		Entrypoint   string

		// configuration we care about
		Name         string
		Hostname     string
		Environment  []string
		Labels       map[string]string
		PublishPorts []string
		Volumes      []string
		Devices      []string
		Tmpfs        []string
		Sysctl       map[string]string
	}

	cmd := &cobra.Command{
		Use:           "run IMAGE",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.V(5).Info("docker run", "flags", flags, "args", args)

			lxcClient, err := env.Client(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			// environment
			var environment string
			for _, v := range flags.Environment {
				if !strings.Contains(v, "=") {
					v = fmt.Sprintf("%s=%s", v, env.Getenv(v))
				}
				environment += v + "\n"
			}

			// labels
			labels := make(map[string]string, len(flags.Labels))
			for key, value := range flags.Labels {
				labels[fmt.Sprintf("user.%s", key)] = value
			}

			// publish ports
			proxyDevices := make(map[string]map[string]string, len(flags.PublishPorts))
			for idx, publishPort := range flags.PublishPorts {
				publishPort, protocol, ok := strings.Cut(strings.ToLower(publishPort), "/")
				if !ok {
					return fmt.Errorf("publish port %q does not specify protocol", publishPort)
				}

				var connect, listen string
				parts := strings.Split(publishPort, ":")
				switch len(parts) {
				case 2: // "16443:6443" -> listen="tcp::16443", connect="tcp::6443"
					listen = fmt.Sprintf("%s::%s", protocol, parts[0])
					connect = fmt.Sprintf("%s::%s", protocol, parts[1])
				case 3: // "127.0.0.1:16443:6443" -> listen="tcp:127.0.0.1:16443", connect="tcp::6443"
					listen = fmt.Sprintf("%s:%s:%s", protocol, parts[0], parts[1])
					connect = fmt.Sprintf("%s::%s", protocol, parts[2])
				default:
					return fmt.Errorf("publish port %q does not specify listen and connect", publishPort)
				}

				proxyDevices[fmt.Sprintf("docker-proxy-%d", idx)] = map[string]string{
					"type":    "proxy",
					"bind":    "host",
					"listen":  listen,
					"connect": connect,
				}
			}

			// tmpfs mounts
			var tmpfsDevices map[string]map[string]string
			if lxcClient.SupportsContainerDiskTmpfs() == nil {
				tmpfsDevices = make(map[string]map[string]string, len(flags.Tmpfs))
				for idx, path := range flags.Tmpfs {
					tmpfsDevices[fmt.Sprintf("docker-tmpfs-%d", idx)] = map[string]string{
						"type":   "disk",
						"path":   path,
						"source": "tmpfs:",
					}
				}
			}

			// unix devices
			unixDevices := make(map[string]map[string]string, len(flags.Devices))
			for idx, device := range flags.Devices {
				unixDevices[fmt.Sprintf("docker-device-%d", idx)] = map[string]string{
					"type":   "unix-char",
					"source": device,
					"path":   device,
				}
			}

			// volumes
			volumeDevices := make(map[string]map[string]string, len(flags.Volumes))
			for idx, volume := range flags.Volumes {
				if volume == "/var" || volume == "/lib/modules:/lib/modules:ro" {
					// these are handled out of band
					continue
				}

				var (
					hostPath      string
					containerPath string
					readOnly      bool
					propagation   string
				)
				parts := strings.Split(volume, ":")
				switch len(parts) {
				case 1: // "/test"
					hostPath = volume
					containerPath = volume
				case 2: // "/test:/test"
					hostPath = parts[0]
					containerPath = parts[1]
				case 3: // "/test:/test:{ro,rshared,rprivate}"
					hostPath = parts[0]
					containerPath = parts[1]
					readOnly = strings.Contains(parts[2], "ro")
					if strings.Contains(parts[2], "rslave") {
						propagation = "rslave"
					} else if strings.Contains(parts[2], "rshared") {
						propagation = "rshared"
					}
				}

				volumeDevices[fmt.Sprintf("docker-volume-%d", idx)] = map[string]string{
					"type":        "disk",
					"source":      hostPath,
					"path":        containerPath,
					"readonly":    strconv.FormatBool(readOnly),
					"propagation": propagation,
				}
			}

			// sysctl configs
			sysctl := make(map[string]string, len(flags.Sysctl))
			for key, value := range flags.Sysctl {
				sysctl[fmt.Sprintf("linux.sysctl.%s", key)] = value
			}

			launchOpts, err := launchOptionsForImage(cmd.Context(), args[0], env, lxcClient.GetServerName())
			if err != nil {
				return fmt.Errorf("failed to generate launch options: %w", err)
			}

			launchOpts = launchOpts.
				WithConfig(labels).
				WithConfig(sysctl).
				WithDevices(proxyDevices).
				WithDevices(volumeDevices).
				WithDevices(tmpfsDevices).
				WithDevices(unixDevices).
				WithAppendToFiles(map[string]string{
					"/etc/environment": "\n# added by kini\n" + environment,
				})

			log.V(4).Info("Launching instance", "opts", strings.ReplaceAll(fmt.Sprintf("%#v", launchOpts), "\"", "'"))
			_, err = lxcClient.WaitForLaunchInstance(cmd.Context(), flags.Name, launchOpts)
			return err
		},
	}

	cmd.Flags().BoolVar(&flags.Init, "init", false, "use entrypoint")
	cmd.Flags().BoolVar(&flags.TTY, "tty", true, "tty")
	cmd.Flags().BoolVar(&flags.Privileged, "privileged", true, "privileged")
	cmd.Flags().BoolVarP(&flags.Detach, "detach", "d", true, "detach")
	cmd.Flags().StringVar(&flags.CgroupNS, "cgroupns", "private", "cgroup namespace")
	cmd.Flags().StringVar(&flags.UserNS, "userns", "", "user namespace")
	cmd.Flags().StringVar(&flags.Network, "net", "kind", "network")
	cmd.Flags().StringVar(&flags.Restart, "restart", "on-failure:1", "restart")
	cmd.Flags().StringToStringVar(&flags.SecurityOpts, "security-opt", nil, "security opt")
	cmd.Flags().StringVar(&flags.Platform, "platform", "", "platform")
	cmd.Flags().StringVar(&flags.Entrypoint, "entrypoint", "", "entrypoint")

	cmd.Flags().StringVar(&flags.Name, "name", "", "container name")
	cmd.Flags().StringVar(&flags.Hostname, "hostname", "", "container host name")
	cmd.Flags().StringArrayVarP(&flags.Environment, "environment", "e", nil, "environment")
	cmd.Flags().StringToStringVar(&flags.Labels, "label", nil, "labels")
	cmd.Flags().StringArrayVar(&flags.PublishPorts, "publish", nil, "publish ports")
	cmd.Flags().StringArrayVar(&flags.Volumes, "volume", nil, "volumes")
	cmd.Flags().StringArrayVar(&flags.Devices, "device", nil, "devices")
	cmd.Flags().StringToStringVar(&flags.Sysctl, "sysctl", nil, "sysctl")
	cmd.Flags().StringArrayVar(&flags.Tmpfs, "tmpfs", nil, "tmpfs")

	return cmd
}
