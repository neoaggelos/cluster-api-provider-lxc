package sync

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lxc/cluster-api-provider-incus/internal/exp/simplestreams/index"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

// manager handles syncing images into a target directory.
type manager struct {
	indexDir   string
	stagingDir string

	// manifest is the currently synced manifest
	manifest Manifest

	client *http.Client
}

func NewManager(indexDir string, stagingDir string, client *http.Client) (*manager, error) {
	var manifest Manifest
	if b, err := os.ReadFile(filepath.Join(indexDir, "images.yaml")); err == nil {
		if err := yaml.UnmarshalStrict(b, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse images.yaml in index, cowardly refusing to proceed: %w", err)
		}
	}
	return &manager{
		indexDir:   indexDir,
		stagingDir: stagingDir,
		manifest:   manifest,
		client:     client,
	}, nil
}

func (m *manager) syncImage(ctx context.Context, index *index.Index, imageID string, image Image) (rerr error) {
	if _, ok := m.manifest.Images[imageID]; ok {
		log.FromContext(ctx).Info("Image is already synced")
		return nil
	}

	log.FromContext(ctx).Info("Downloading image", "source", image.Source)
	resp, err := m.client.Get(image.Source)
	if err != nil {
		return fmt.Errorf("failed to download image: HTTP request GET %q failed: %w", image.Source, err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to download image: GET %q failed: %s", image.Source, resp.Status)
	}
	defer resp.Body.Close()

	f, err := os.Create(filepath.Join(m.stagingDir, imageID))
	if err != nil {
		return fmt.Errorf("failed to download image: failed to create file: %w", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to download image: failed to write file: %w", err)
	}

	if image.Checksum != "" {
		// TODO: validate checksum
		log.FromContext(ctx).Info("Ignoring checksum check", "checksum", image.Checksum)
	}

	log.FromContext(ctx).Info("Syncing image")
	if err := index.ImportImage(ctx, image.Type, filepath.Join(m.stagingDir, imageID), image.Alias, true, true); err != nil {
		return fmt.Errorf("failed to import image: %w", err)
	}

	return nil
}

// Sync
func (m *manager) Sync(ctx context.Context, manifest Manifest) error {
	index, err := index.GetOrCreateIndex(m.indexDir)
	if err != nil {
		return fmt.Errorf("failed to open index at %q: %w", m.indexDir, err)
	}
	for imageID, image := range manifest.Images {
		imageCtx := log.IntoContext(ctx, log.FromContext(ctx, "image", imageID))
		if err := m.syncImage(imageCtx, index, imageID, image); err != nil {
			return fmt.Errorf("failed to sync %q: %w", imageID, err)
		}
	}

	log.FromContext(ctx).Info("Writing images.yaml")
	b, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to add images.yaml manifest: failed to format YAML: %w", err)
	}
	if err := os.WriteFile(filepath.Join(m.indexDir, "images.yaml"), b, 0644); err != nil {
		return fmt.Errorf("failed to add images.yaml manifest: failed to write file: %w", err)
	}

	return nil
}
