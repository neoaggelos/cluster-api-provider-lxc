package kini

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SetupEnvironment creates a temporary directory and symlinks "docker" and "kind" to the current binary.
// SetupEnvironment updates the value of the PATH environment variable, and returns the directory path and a cleanup function to revert.
// SetupEnvironment is used to setup the environment such that kind commands work through the kini docker shim command.
func SetupEnvironment(ctx context.Context, docker bool, kind bool) (string, func() error, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	log := log.FromContext(ctx).WithValues("dir", dir, "docker", docker, "kind", kind)
	log.V(4).Info("Setting up")

	cleanup := func() error {
		log.V(4).Info("Cleaning up")
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to clean up temporary directory: %w", err)
		}
		return nil
	}

	self, err := filepath.Abs(os.Args[0])
	if err != nil {
		_ = cleanup()
		return "", nil, fmt.Errorf("failed to identity absolute path to %q: %w", os.Args[0], err)
	}

	if docker {
		if err := os.Symlink(self, filepath.Join(dir, "docker")); err != nil {
			_ = cleanup()
			return "", nil, fmt.Errorf("failed to create symlink as docker for self: %w", err)
		}
	}
	if kind {
		if err := os.Symlink(self, filepath.Join(dir, "kind")); err != nil {
			_ = cleanup()
			return "", nil, fmt.Errorf("failed to create symlink as kind for self: %w", err)
		}
	}

	currentPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", fmt.Sprintf("%s:%s", dir, currentPath)); err != nil {
		_ = cleanup()
		return "", nil, fmt.Errorf("failed to set PATH: %w", err)
	}

	return dir, cleanup, nil
}
