package lxc

import (
	"context"
	"fmt"
	"strings"

	"github.com/lxc/incus/v6/shared/api"

	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

func (c *Client) TryParseImageSource(ctx context.Context, imageName string) (api.InstanceSource, bool, error) {
	if version, ok := strings.CutPrefix(imageName, "ubuntu:"); ok {
		return c.getDefaultUbuntuImage(ctx, version)
	}
	if version, ok := strings.CutPrefix(imageName, "debian:"); ok {
		return c.getDefaultDebianImage(ctx, version)
	}
	if image, ok := strings.CutPrefix(imageName, "images:"); ok {
		return c.getDefaultRepoImage(ctx, image)
	}

	return api.InstanceSource{}, false, nil
}

func (c *Client) getDefaultUbuntuImage(ctx context.Context, version string) (api.InstanceSource, bool, error) {
	serverName := c.GetServerName(ctx)
	switch serverName {
	case Incus:
		return api.InstanceSource{
			Type:     "image",
			Alias:    fmt.Sprintf("ubuntu/%s/cloud", version),
			Server:   "https://images.linuxcontainers.org",
			Protocol: "simplestreams",
		}, true, nil
	case LXD:
		return api.InstanceSource{
			Type:     "image",
			Alias:    version,
			Server:   "https://cloud-images.ubuntu.com/releases/",
			Protocol: "simplestreams",
		}, true, nil
	default:
		return api.InstanceSource{}, false, utils.TerminalError(fmt.Errorf("image name is %q, but server is %q. Images with 'ubuntu:' prefix are only allowed for Incus and LXD", version, serverName))
	}
}

func (c *Client) getDefaultDebianImage(ctx context.Context, version string) (api.InstanceSource, bool, error) {
	serverName := c.GetServerName(ctx)
	switch serverName {
	case Incus:
		return api.InstanceSource{
			Type:     "image",
			Alias:    fmt.Sprintf("debian/%s/cloud", version),
			Server:   "https://images.linuxcontainers.org",
			Protocol: "simplestreams",
		}, true, nil
	case LXD:
		return api.InstanceSource{
			Type:     "image",
			Alias:    fmt.Sprintf("debian/%s/cloud", version),
			Server:   "https://images.lxd.canonical.com",
			Protocol: "simplestreams",
		}, true, nil
	default:
		return api.InstanceSource{}, false, utils.TerminalError(fmt.Errorf("image name is %q, but server is %q. Images with 'debian:' prefix are only allowed for Incus and LXD", version, serverName))
	}
}

func (c *Client) getDefaultRepoImage(ctx context.Context, image string) (api.InstanceSource, bool, error) {
	serverName := c.GetServerName(ctx)
	switch serverName {
	case Incus:
		return api.InstanceSource{
			Type:     "image",
			Alias:    image,
			Server:   "https://images.linuxcontainers.org",
			Protocol: "simplestreams",
		}, true, nil
	case LXD:
		return api.InstanceSource{
			Type:     "image",
			Alias:    image,
			Server:   "https://images.lxd.canonical.com",
			Protocol: "simplestreams",
		}, true, nil
	default:
		return api.InstanceSource{}, false, utils.TerminalError(fmt.Errorf("image name is %q, but server is %q. Images with 'images:' prefix are only allowed for Incus and LXD", image, serverName))
	}
}
