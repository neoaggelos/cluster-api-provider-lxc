package sync

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	"github.com/lxc/cluster-api-provider-incus/internal/exp/simplestreams/index"
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
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("failed to download image: failed to fetch: %w", err)
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

func (m *manager) syncManifest(ctx context.Context, manifest Manifest) error {
	log.FromContext(ctx).Info("Writing images.yaml")
	b, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to format YAML: %w", err)
	}
	if err := os.WriteFile(filepath.Join(m.indexDir, "images.yaml"), b, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (m *manager) syncHashes(ctx context.Context, index *index.Index) error {
	hashes := map[string]string{}
	for _, path := range []string{
		"streams/v1/index.json",
		"streams/v1/images.json",
		"images.yaml",
	} {
		hash, err := calculateFileSha256(filepath.Join(m.indexDir, path))
		if err != nil {
			return fmt.Errorf("failed to calculate sha256 of %q: %w", path, err)
		}
		hashes[path] = hash
	}

	for _, product := range index.Products.Products {
		for _, version := range product.Versions {
			for _, item := range version.Items {
				hashes[item.Path] = item.HashSha256
			}
		}
	}

	log.FromContext(ctx).Info("Writing files.sha256", "items", len(hashes))
	lines := make([]string, 0, len(hashes)+1)
	lines = append(lines, fmt.Sprintf("# generated at %v", time.Now()))
	for _, key := range slices.Sorted(maps.Keys(hashes)) {
		lines = append(lines, fmt.Sprintf("%s  %s", hashes[key], key))
	}

	if err := os.WriteFile(filepath.Join(m.indexDir, "files.sha256"), []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write files.sha256: %w", err)
	}
	return nil
}

// Sync
func (m *manager) Sync(ctx context.Context, manifest Manifest) error {
	index, err := index.GetOrCreateIndex(m.indexDir)
	if err != nil {
		return fmt.Errorf("failed to open index at %q: %w", m.indexDir, err)
	}
	for _, imageID := range slices.Sorted(maps.Keys(manifest.Images)) {
		image := manifest.Images[imageID]
		imageCtx := log.IntoContext(ctx, log.FromContext(ctx, "image", imageID))
		if err := m.syncImage(imageCtx, index, imageID, image); err != nil {
			return fmt.Errorf("failed to sync %q: %w", imageID, err)
		}
	}

	if err := m.syncManifest(ctx, manifest); err != nil {
		return fmt.Errorf("failed to sync images.yaml: %w", err)
	}

	if err := m.syncHashes(ctx, index); err != nil {
		return fmt.Errorf("failed to sync files.sha256: %w", err)
	}

	return nil
}
