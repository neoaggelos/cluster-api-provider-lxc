package instances

import (
	"github.com/lxc/incus/v7/shared/api"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// HaproxyLXCLaunchOptions launches LXC haproxy load balancer containers.
func HaproxyLXCLaunchOptions() *lxc.LaunchOptions {
	return (&lxc.LaunchOptions{}).
		WithInstanceType(api.InstanceTypeContainer).
		WithImage(lxc.Image{
			Protocol: lxc.Simplestreams,
			Server:   lxc.DefaultSimplestreamsServer,
			Alias:    "haproxy",
		})
}

// HaproxyOCILaunchOptions launches OCI haproxy load balancer containers.
func HaproxyOCILaunchOptions() *lxc.LaunchOptions {
	return (&lxc.LaunchOptions{}).
		WithInstanceType(api.InstanceTypeContainer).
		WithImage(lxc.Image{
			Protocol: lxc.OCI,
			Server:   "https://ghcr.io",
			Alias:    "lxc/cluster-api-provider-incus/haproxy:v20230606-42a2262b",
		}).
		WithSymlinks(map[string]string{
			// Incus will inject its own PID 1 init process unless the entrypoint is one of "/init", "/sbin/init", "/s6-init".
			"/init": "/usr/sbin/haproxy",
		}).
		WithConfig(map[string]string{
			// Use the /init symlink to avoid the Incus entrypoint from preventing SIGUSR2 propagating to child processes.
			"oci.entrypoint": "/init -W -db -f /usr/local/etc/haproxy/haproxy.cfg",
		}).
		WithCreateFiles(map[string]string{
			// Default kindest/haproxy image does not have /etc/environment, leading to issues downstream.
			"/etc/environment": "",
		})
}
