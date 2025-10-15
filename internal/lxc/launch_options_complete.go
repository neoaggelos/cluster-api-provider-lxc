package lxc

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/lxc/incus/v6/shared/api"

	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

// complete launch options and verify
func (o *LaunchOptions) complete(serverName string) error {
	// virtual machines instances do not support unixSocket, createFiles or replacements
	if o.instanceType == api.InstanceTypeVM {
		if o.unixSocket {
			return utils.TerminalError(fmt.Errorf("mounting unix socket not supported for virtual-machine instances"))
		}
		if len(o.createFiles) > 0 {
			ops := make([]string, 0, len(o.createFiles))
			for _, f := range o.createFiles {
				ops = append(ops, f.action())
			}
			return utils.TerminalError(fmt.Errorf("operations not supported for virtual-machine instances: %v", ops))
		}
		if len(o.replacements) > 0 {
			return utils.TerminalError(fmt.Errorf("replacements not supported for virtual-machine instances: %v", slices.Collect(maps.Keys(o.replacements))))
		}
	}

	// complete image configuration
	if o.image == nil {
		return utils.TerminalError(fmt.Errorf("cannot launch instance without image"))
	}
	image, err := o.image.For(serverName)
	if err != nil {
		return fmt.Errorf("unsupported instance image: %w", err)
	}
	// if OCI image is specified as `IMG[:TAG]@sha256:HASH`, replace with `IMG@sha256:HASH`
	if image.Protocol == OCI {
		if imageWithoutHash, hash, ok := strings.Cut(image.Alias, "@"); ok {
			imageWithoutTag, _, _ := strings.Cut(imageWithoutHash, ":")
			image.Alias = fmt.Sprintf("%s@%s", imageWithoutTag, hash)
		}
	}
	o.image = image

	// load unix socket to /run-unix.socket inside the instance
	if o.unixSocket {
		path, err := GetDefaultUnixSocketPathFor(serverName)
		if err != nil {
			return fmt.Errorf("failed to get unix socket path: %w", err)
		}

		// NOTE(neoaggelos): disable SA4006: this value of o is never used (staticcheck)
		//nolint:staticcheck
		o = o.WithDevices(map[string]map[string]string{
			"00-unix-socket": {
				"type":   "disk",
				"source": path,
				"path":   "/run-unix.socket",
				"shift":  "true",
			},
		})
	}

	return nil
}
