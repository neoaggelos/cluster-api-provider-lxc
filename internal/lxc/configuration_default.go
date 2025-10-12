package lxc

import (
	"errors"
	"fmt"
	"os"

	"github.com/lxc/incus/v6/shared/util"
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

func getDefaultUnixSocketPath() (string, error) {
	var errs []error

	for _, file := range []string{
		"/var/lib/incus/unix.socket",
		"/var/snap/lxd/common/lxd/unix.socket",
		"/run/unix.socket", // unix socket path used as default fallback
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
