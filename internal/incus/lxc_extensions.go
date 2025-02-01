package incus

import "fmt"

// SupportsInstanceOCI checks if the necessary API extensions for OCI instances are supported by the server.
//
// If instance_oci is not supported, a terminalError is returned.
func (c *Client) SupportsInstanceOCI() error {
	if unsupported, err := c.serverSupportsExtensions("instance_oci"); err != nil {
		return fmt.Errorf("failed to check if server supports 'instance_oci' extension: %w", err)
	} else if len(unsupported) > 0 {
		return terminalError{fmt.Errorf("server cannot create OCI containers, required extensions are missing: %v", unsupported)}
	}
	return nil
}

// SupportsNetworkLoadBalancer checks if the necessary API extensions for Network Load Balancers are supported by the server.
//
// If network_load_balancer or network_load_balancer_health_check are not supported, a terminalError is returned.
func (c *Client) SupportsNetworkLoadBalancer() error {
	if unsupported, err := c.serverSupportsExtensions("network_load_balancer", "network_load_balancer_health_check"); err != nil {
		return fmt.Errorf("failed to check if server supports network load balancer extensions: %w", err)
	} else if len(unsupported) > 0 {
		return terminalError{fmt.Errorf("server cannot create network load balancers, required extensions are missing: %v", unsupported)}
	}
	return nil
}
