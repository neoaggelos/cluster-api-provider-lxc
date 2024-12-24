package incus

import "time"

const (
	// loadBalancerCreateTimeout is the timeout for creating and starting the load balancer container.
	loadBalancerCreateTimeout = 60 * time.Second

	// loadBalancerDeleteTimeout is the timeout for stopping and deleting the load balancer container.
	loadBalancerDeleteTimeout = 30 * time.Second

	// loadBalancerReconfigureTimeout is the timeout for updating load balancer configuration and reloading.
	loadBalancerReconfigureTimeout = 60 * time.Second

	// loadBalancerDefaultHaproxyImage is the default image to use for the load balancer container.
	// TODO(neoaggelos): mirror and use our own image
	loadBalancerDefaultHaproxyImage = "kindest/haproxy:v20230606-42a2262b"

	// loadBalancerDefaultHaproxyImageRegistry is the default OCI registry we will pull the load balancer haproxy image.
	loadBalancerDefaultHaproxyImageRegistry = "https://docker.io"

	// loadBalancerDefaultHaproxyConfigPath is the path where haproxy config is created.
	loadBalancerDefaultHaproxyConfigPath = "/usr/local/etc/haproxy/haproxy.cfg"

	// configClusterNameKey is the user config key that tracks the cluster name.
	configClusterNameKey = "user.cluster-name"

	// configClusterNamespaceKey is the user config key that tracks the cluster namespace.
	configClusterNamespaceKey = "user.cluster-namespace"

	// configInstanceRoleKey is the user config key that tracks the instance role.
	configInstanceRoleKey = "user.cluster-role"
)
