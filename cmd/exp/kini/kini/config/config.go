package config

import (
	"context"
	"fmt"

	"github.com/lxc/incus/v6/shared/cliconfig"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// Manager can be used to manage the local Incus configuration file.
type Manager struct {
	// config is the loaded configuration.
	config *cliconfig.Config

	// path is the path from which the configuration was loaded.
	path string

	// mustUpdate tracks whether the configuration has been changed from the one on disk.
	mustUpdate bool
}

func NewManager(ctx context.Context, configFile string) (*Manager, error) {
	_, path, err := lxc.ConfigurationFromLocal(configFile, "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to read local configuration: %w", err)
	}

	log.FromContext(ctx).V(1).Info("Found local configuration file", "path", path)

	cfg, err := cliconfig.LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return &Manager{config: cfg, path: path}, nil
}

func (m *Manager) AddSimplestreamsRemoteIfNotExist(ctx context.Context, name string, server string) error {
	log := log.FromContext(ctx).WithValues("name", name, "server", server)
	if _, ok := m.config.Remotes[name]; ok {
		log.V(1).Info("Remote already exists, will not do anything")
		return nil
	}

	log.V(1).Info("Adding remote to local configuration")
	m.config.Remotes[name] = cliconfig.Remote{
		Addr:     server,
		Protocol: lxc.Simplestreams,
		Public:   true,
	}
	m.mustUpdate = true

	return nil
}

func (m *Manager) Commit(ctx context.Context) error {
	if !m.mustUpdate {
		log.FromContext(ctx).V(1).Info("No configuration changes to commit")
		return nil
	}

	log.FromContext(ctx).V(1).Info("Committing changes to config file", "path", m.path)
	if err := m.config.SaveConfig(m.path); err != nil {
		return fmt.Errorf("failed to update config file on disk: %w", err)
	}

	m.mustUpdate = false
	return nil
}
