package lxc

import (
	"fmt"
)

func UbuntuImage(version string) ImageFamily {
	return &imageFamily{
		Description: "ubuntu",
		Sources: map[string]Image{
			Incus: { // "ubuntu:24.04" -> "ubuntu/24.04/cloud" from "https://images.linuxcontainers.org"
				Protocol: Simplestreams,
				Server:   DefaultIncusSimplestreamsServer,
				Alias:    fmt.Sprintf("ubuntu/%s/cloud", version),
			},
			LXD: { // "ubuntu:24.04" -> "24.04" from "https://cloud-images.ubuntu.com/releases/"
				Protocol: Simplestreams,
				Server:   DefaultLXDUbuntuSimplestreamsServer,
				Alias:    version,
			},
		},
	}
}

func DebianImage(version string) ImageFamily {
	return &imageFamily{
		Description: "debian",
		Sources: map[string]Image{
			Incus: { // "debian:13" -> "debian/13/cloud" from "https://images.linuxcontainers.org"
				Protocol: Simplestreams,
				Server:   DefaultIncusSimplestreamsServer,
				Alias:    fmt.Sprintf("debian/%s/cloud", version),
			},
			LXD: { // "debian:13" -> "debian/13/cloud" from "https://images.lxd.canonical.com"
				Protocol: Simplestreams,
				Server:   DefaultLXDSimplestreamsServer,
				Alias:    fmt.Sprintf("debian/%s/cloud", version),
			},
		},
	}
}

func DefaultImage(image string) ImageFamily {
	return &imageFamily{
		Description: "images",
		Sources: map[string]Image{
			Incus: { // "images:IMAGE" -> "IMAGE" from "https://images.linuxcontainers.org"
				Protocol: Simplestreams,
				Server:   DefaultIncusSimplestreamsServer,
				Alias:    image,
			},
			LXD: { // "images:IMAGE" -> "IMAGE" from "https://images.lxd.canonical.com"
				Protocol: Simplestreams,
				Server:   DefaultLXDSimplestreamsServer,
				Alias:    image,
			},
		},
	}
}

func CapnImage(image string) Image {
	return Image{ // "capi:IMAGE" -> "IMAGE" from "https://images.linuxcontainers.org/capn/"
		Protocol: Simplestreams,
		Server:   DefaultSimplestreamsServer,
		Alias:    image,
	}
}

func CapnStagingImage(image string) Image {
	return Image{ // "capi-stg:IMAGE" -> "IMAGE" from "https://images-stg.capn.open-cloud.xyz/capn/staging/"
		Protocol: Simplestreams,
		Server:   DefaultStagingSimplestreamsServer,
		Alias:    image,
	}
}

func KindestNodeImage(version string) Image {
	return Image{ // "kind:VERSION" -> "kindest/node:VERSION" from "https://docker.io"
		Protocol: OCI,
		Server:   DockerHubServer,
		Alias:    fmt.Sprintf("kindest/node:%s", version),
	}
}
