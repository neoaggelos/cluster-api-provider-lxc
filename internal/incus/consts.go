package incus

import "time"

const (
	// loadBalancerCreateTimeout is the timeout for creating and starting the load balancer container.
	loadBalancerCreateTimeout = 60 * time.Second

	// loadBalancerDeleteTimeout is the timeout for stopping and deleting the load balancer container.
	loadBalancerDeleteTimeout = 30 * time.Second

	// loadBalancerReconfigureTimeout is the timeout for updating load balancer configuration and reloading.
	loadBalancerReconfigureTimeout = 60 * time.Second

	// instanceCreateTimeout is the timeout for creating and starting an instance.
	instanceCreateTimeout = 180 * time.Second

	// instanceDeleteTimeout is the timeout for stopping and deleting an instance.
	instanceDeleteTimeout = 30 * time.Second

	// configClusterNameKey is the user config key that tracks the cluster name.
	configClusterNameKey = "user.cluster-name"

	// configClusterNamespaceKey is the user config key that tracks the cluster namespace.
	configClusterNamespaceKey = "user.cluster-namespace"

	// configMachineNameKey is the user config key that tracks the machine name.
	configMachineNameKey = "user.machine-name"

	// configInstanceRoleKey is the user config key that tracks the instance role.
	configInstanceRoleKey = "user.cluster-role"

	// configCloudInitKey is the config key that seeds cloud-init configuration into the instance.
	configCloudInitKey = "cloud-init.user-data"

	// defaultSimplestreamsServer is the default simplestreams server for fetching images.
	defaultSimplestreamsServer = "https://d14dnvi2l3tc5t.cloudfront.net"
)
