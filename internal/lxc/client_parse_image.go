package lxc

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/lxc/incus/v6/shared/api"

	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

type imageInfo struct {
	server    string
	transform func(in string) string
}

var (
	// wellKnownImagePrefixes defines (by server type) well-known image prefixes.
	wellKnownImagePrefixes = map[string]map[string]imageInfo{
		Incus: {
			"ubuntu": { // "ubuntu:24.04" -> "ubuntu/24.04/cloud" from "https://images.linuxcontainers.org"
				server:    DefaultIncusSimplestreamsServer,
				transform: func(in string) string { return fmt.Sprintf("ubuntu/%s/cloud", in) },
			},
			"debian": { // "debian:12" -> "debian/12/cloud" from "https://images.linuxcontainers.org"
				server:    DefaultIncusSimplestreamsServer,
				transform: func(in string) string { return fmt.Sprintf("debian/%s/cloud", in) },
			},
			"images": { // "images:IMAGE" -> "IMAGE" from "https://images.linuxcontainers.org"
				server:    DefaultIncusSimplestreamsServer,
				transform: func(in string) string { return in },
			},
			"capi": { // "capi:IMAGE" -> "IMAGE" from "https://d14dnvi2l3tc5t.cloudfront.net"
				server:    DefaultSimplestreamsServer,
				transform: func(in string) string { return in },
			},
			"capi-stg": { // "capi-stg:IMAGE" -> "IMAGE" from "https://djapqxqu5n2qu.cloudfront.net"
				server:    DefaultStagingSimplestreamsServer,
				transform: func(in string) string { return in },
			},
		},
		LXD: {
			"ubuntu": { // "ubuntu:24.04" -> "24.04" from "https://cloud-images.ubuntu.com/releases/"
				server:    DefaultLXDUbuntuSimplestreamsServer,
				transform: func(in string) string { return in },
			},
			"debian": { // "debian:12" -> "debian/12/cloud" from "https://images.lxd.canonical.com"
				server:    DefaultLXDSimplestreamsServer,
				transform: func(in string) string { return fmt.Sprintf("debian/%s/cloud", in) },
			},
			"images": { // "images:IMAGE" -> "IMAGE" from "https://images.lxd.canonical.com"
				server:    DefaultLXDSimplestreamsServer,
				transform: func(in string) string { return in },
			},
			"capi": { // "capi:IMAGE" -> "IMAGE" from "https://d14dnvi2l3tc5t.cloudfront.net"
				server:    DefaultSimplestreamsServer,
				transform: func(in string) string { return in },
			},
			"capi-stg": { // "capi-stg:IMAGE" -> "IMAGE" from "https://djapqxqu5n2qu.cloudfront.net"
				server:    DefaultStagingSimplestreamsServer,
				transform: func(in string) string { return in },
			},
		},
	}
)

func TryParseImageSource(serverName, imageName string) (api.InstanceSource, bool, error) {
	parts := strings.Split(imageName, ":")
	if len(parts) != 2 {
		return api.InstanceSource{}, false, nil
	}

	if info, ok := wellKnownImagePrefixes[serverName][parts[0]]; ok {
		return api.InstanceSource{
			Type:     "image",
			Protocol: "simplestreams",
			Server:   info.server,
			Alias:    info.transform(parts[1]),
		}, true, nil
	}
	if prefixes := slices.Collect(maps.Keys(wellKnownImagePrefixes[serverName])); len(prefixes) > 0 {
		return api.InstanceSource{}, false, utils.TerminalError(fmt.Errorf("unknown image prefix %q for server %q. must be one of %v", parts[0], serverName, prefixes))
	}
	return api.InstanceSource{}, false, utils.TerminalError(fmt.Errorf("server type %q does not spuport any image prefixes", serverName))
}
