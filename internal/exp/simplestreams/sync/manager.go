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

	client *http.Client
}

func NewManager(indexDir string, stagingDir string, client *http.Client) *manager {
	return &manager{
		indexDir:   indexDir,
		stagingDir: stagingDir,
		client:     client,
	}
}

func (m *manager) markImageSynced(imageID string) error {
	if err := os.MkdirAll(filepath.Join(m.indexDir, "image-ids"), 0755); err != nil {
		return fmt.Errorf("failed to create image-ids directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(m.indexDir, "image-ids", imageID), nil, 0644); err != nil {
		return fmt.Errorf("failed to create image-id mark: %w", err)
	}
	return nil
}

func (m *manager) isImageSynced(imageID string) (bool, error) {
	if _, err := os.Stat(filepath.Join(m.indexDir, "image-ids", imageID)); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("failed to check image-ids file: %w", err)
	}
}

func (m *manager) syncImage(ctx context.Context, index *index.Index, image Image) (rerr error) {
	if v, err := m.isImageSynced(image.ID); err != nil {
		return fmt.Errorf("failed to check if image is synced: %w", err)
	} else if v {
		log.FromContext(ctx).Info("Image is already synced")
		return nil
	}
	defer func() {
		if rerr == nil {
			log.FromContext(ctx).Info("Marking image as synced")
			if err := m.markImageSynced(image.ID); err != nil {
				rerr = fmt.Errorf("failed to mark %q as synced: %w", image.ID, err)
			}
		}
	}()

	log.FromContext(ctx).Info("Downloading image", "source", image.Source)
	resp, err := m.client.Get(image.Source)
	if err != nil {
		return fmt.Errorf("failed to download image: HTTP request GET %q failed: %w", image.Source, err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to download image: GET %q failed: %s", image.Source, resp.Status)
	}
	defer resp.Body.Close()

	f, err := os.Create(filepath.Join(m.stagingDir, image.ID))
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
	if err := index.ImportImage(ctx, image.Type, filepath.Join(m.stagingDir, image.ID), image.Alias, true, true); err != nil {
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
	for _, image := range manifest.Images {
		imageCtx := log.IntoContext(ctx, log.FromContext(ctx, "image", image.ID))
		if err := m.syncImage(imageCtx, index, image); err != nil {
			return fmt.Errorf("failed to sync %q: %w", image.ID, err)
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
