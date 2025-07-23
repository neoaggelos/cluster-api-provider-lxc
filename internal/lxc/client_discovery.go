package lxc

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

// GetServerName returns one of "incus", "lxd" or "unknown", depending on the server type.
func (c *Client) GetServerName(ctx context.Context) string {
	switch c.serverInfo.Environment.Server {
	case Incus, LXD:
		return c.serverInfo.Environment.Server
	default:
		return "unknown"
	}
}

// The built-in Client.HasExtension() from Incus cannot be trusted, as it returns true if we skip the GetServer call.
// Return the list of extensions that are NOT supported by the server, if any.
func (c *Client) serverSupportsExtensions(extensions ...string) error {
	if missing := sets.New(extensions...).Difference(sets.New(c.serverInfo.APIExtensions...)).UnsortedList(); len(missing) > 0 {
		return utils.TerminalError(fmt.Errorf("required extensions %v are not supported", missing))
	}
	return nil
}

func (c *Client) SupportsInstanceOCI() error {
	return c.serverSupportsExtensions("instance_oci", "instance_oci_entrypoint")
}

func (c *Client) SupportsNetworkLoadBalancers() error {
	return c.serverSupportsExtensions("network_load_balancer", "network_load_balancer_health_check")
}

func (c *Client) SupportsInstanceKVM() error {
	if !slices.Contains(strings.Split(c.serverInfo.Environment.Driver, " | "), "qemu") {
		return utils.TerminalError(fmt.Errorf("server is missing driver qemu, supported drivers are: %q", c.serverInfo.Environment.Driver))
	}
	return nil
}

func (c *Client) SupportsArchitectures() []string {
	return slices.Clone(c.serverInfo.Environment.Architectures)
}

func (c *Client) SupportsInstanceTarget() error {
	if !c.serverInfo.Environment.ServerClustered {
		return fmt.Errorf("server is not part of a cluster")
	}
	return nil
}
