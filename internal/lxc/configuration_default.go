package lxc

import (
	"errors"
	"fmt"
	"os"

	"github.com/lxc/incus/v6/shared/util"

	"github.com/lxc/cluster-api-provider-incus/internal/utils"
)

func getDefaultConfigFiles() []string {
	var files []string

	for _, file := range []string{
		os.ExpandEnv("${HOME}/.config/incus/config.yml"),
		os.ExpandEnv("${HOME}/snap/lxd/common/config/config.yml"),
	} {
		if _, err := os.Stat(file); err == nil {
			files = append(files, file)
		}
	}

	return append(files, "")
}

func findDefaultUnixSocketPath() (string, error) {
	var errs []error

	for _, file := range []string{
		"/var/lib/incus/unix.socket",
		"/run/incus/unix.socket", // alternate incus unix socket path
		"/var/snap/lxd/common/lxd/unix.socket",
		"/run-unix.socket", // unix socket path used as default fallback
	} {
		if _, err := os.Stat(file); err != nil {
			errs = append(errs, fmt.Errorf("failed to stat %q: %w", file, err))
			continue
		}
		if !util.PathIsWritable(file) {
			errs = append(errs, fmt.Errorf("%q is not writeable", file))
			continue
		}

		return file, nil
	}

	return "", errors.Join(errs...)
}

func GetDefaultUnixSocketPathFor(serverName string) (string, error) {
	// FIXME(neoaggelos): we should consider alternate paths for the unix socket, e.g. /run/incus/unix.socket
	// however, do not have a way to check if the socket path is valid without creating an instance
	if path, ok := map[string]string{
		LXD:   "/var/snap/lxd/common/lxd/unix.socket",
		Incus: "/var/lib/incus/unix.socket",
	}[serverName]; !ok {
		return "", utils.TerminalError(fmt.Errorf("unknown default unix socket path for server %q", serverName))
	} else {
		return path, nil
	}
}
