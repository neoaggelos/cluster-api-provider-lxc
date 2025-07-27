package lxc

import "time"

const (
	// DefaultSimplestreamsServer is the default simplestreams server for fetching images.
	DefaultSimplestreamsServer = "https://d14dnvi2l3tc5t.cloudfront.net"

	// DefaultStagingSimplestreamsServer is the default staging simplestreams server for fetching images.
	DefaultStagingSimplestreamsServer = "https://djapqxqu5n2qu.cloudfront.net"

	// DefaultIncusSimplestreamsServer is the default simplestreams server for Incus.
	DefaultIncusSimplestreamsServer = "https://images.linuxcontainers.org"

	// DefaultLXDSimplestreamsServer is the default simplestreams server for Canonical LXD.
	DefaultLXDSimplestreamsServer = "https://images.lxd.canonical.com"

	// DefaultLXDUbuntuSimplestreamsServer is the default simplestreams server for Ubuntu images for Canonical LXD.
	DefaultLXDUbuntuSimplestreamsServer = "https://cloud-images.ubuntu.com/releases/"

	// Container is the instance type for container instances.
	Container = "container"

	// VirtualMachine is the instance type for virtual-machine instances.
	VirtualMachine = "virtual-machine"

	// Incus is the server name for Incus servers.
	Incus = "incus"

	// LXD is the server name for Canonical LXD servers.
	LXD = "lxd"

	// instanceCreateTimeout is the timeout for creating an instance.
	instanceCreateTimeout = 180 * time.Second

	// instanceStartTimeout is the timeout for starting an instance.
	instanceStartTimeout = 60 * time.Second

	// instanceStopTimeout is the timeout for stopping an instance.
	instanceStopTimeout = 30 * time.Second

	// instanceDeleteTimeout is the timeout for stopping and deleting an instance.
	instanceDeleteTimeout = 30 * time.Second
)
