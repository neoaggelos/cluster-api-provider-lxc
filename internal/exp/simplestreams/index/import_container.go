package index

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/lxc/incus/v6/shared/api"
	"github.com/lxc/incus/v6/shared/simplestreams"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

func (i *Index) importContainerUnifiedTarball(ctx context.Context, imagePath string, aliases []string, incus bool, lxd bool) error {
	log.FromContext(ctx).Info("Importing container image", "image", imagePath)

	f, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz archive: %w", err)
	}
	defer func() {
		_ = gzReader.Close()
	}()
	tarReader := tar.NewReader(gzReader)

	var metadata api.ImageMetadata
	for {
		if hdr, err := tarReader.Next(); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read tar.gz archive: %w", err)
		} else if hdr.Name == "metadata.yaml" {
			b, err := io.ReadAll(tarReader)
			if err != nil {
				return fmt.Errorf("failed to read metadata.yaml from image: %w", err)
			}

			if err := yaml.Unmarshal(b, &metadata); err != nil {
				return fmt.Errorf("failed to parse metadata.yaml from image: %w", err)
			}

			break
		}
	}

	if metadata.Architecture == "" {
		return fmt.Errorf("no metadata.yaml found for image")
	}

	info, err := getContainerImageInfo(imagePath)
	if err != nil {
		return fmt.Errorf("failed to retrieve image information: %w", err)
	}

	productName := fmt.Sprintf("%s:%s:%s:%s", metadata.Properties["os"], metadata.Properties["release"], metadata.Properties["variant"], metadata.Properties["architecture"])
	versionName := time.Unix(metadata.CreationDate, 0).Format("200601021504")
	target := filepath.Join("images", metadata.Properties["os"], metadata.Properties["release"], metadata.Properties["architecture"], fmt.Sprintf("%s.incus_combined.tar.gz", info.Sha256))

	log.FromContext(ctx).Info("Adding product version item", "product", productName, "version", versionName, "info", info)

	// update index
	if !slices.Contains(i.Index.Index["images"].Products, productName) {
		log.FromContext(ctx).Info("Adding product in streams/v1/index.json")

		newImages := i.Index.Index["images"]
		newImages.Products = append(newImages.Products, productName)
		slices.Sort(newImages.Products)
		i.Index.Index["images"] = newImages
	}

	// update product versions
	var product simplestreams.Product
	if existingProduct, ok := i.Products.Products[productName]; ok {
		product = existingProduct
	} else {
		log.FromContext(ctx).Info("Creating product", "product", productName)
		product = simplestreams.Product{
			Architecture:    metadata.Properties["architecture"],
			OperatingSystem: metadata.Properties["os"],
			Release:         metadata.Properties["release"],
			ReleaseTitle:    metadata.Properties["release"],
			Variant:         metadata.Properties["variant"],
			Versions: map[string]simplestreams.ProductVersion{
				versionName: {},
			},
		}
	}

	if len(aliases) > 0 {
		log.FromContext(ctx).Info("Setting product aliases", "product", productName, "aliases", aliases)
		product.Aliases = strings.Join(aliases, ",")
	}

	if product.Versions == nil {
		product.Versions = make(map[string]simplestreams.ProductVersion)
	}
	newProductVersions := product.Versions[versionName]
	if newProductVersions.Items == nil {
		newProductVersions.Items = make(map[string]simplestreams.ProductVersionItem)
	}

	if incus {
		log.FromContext(ctx).Info("Adding product version item", "ftype", containerFTypeIncus, "path", target)
		newProductVersions.Items[containerFTypeIncus] = simplestreams.ProductVersionItem{
			FileType:   containerFTypeIncus,
			HashSha256: info.Sha256,
			Size:       info.Size,
			Path:       target,
		}
	}
	if lxd {
		log.FromContext(ctx).Info("Adding product version item", "ftype", containerFTypeLXD, "path", target)
		newProductVersions.Items[containerFTypeLXD] = simplestreams.ProductVersionItem{
			FileType:   containerFTypeLXD,
			HashSha256: info.Sha256,
			Size:       info.Size,
			Path:       target,
		}
	}
	product.Versions[versionName] = newProductVersions
	i.Products.Products[productName] = product

	// copy image
	log.FromContext(ctx).Info("Copying image file into simplestreams index", "source", imagePath, "destination", filepath.Join(i.rootDir, target))
	if err := copyFile(imagePath, filepath.Join(i.rootDir, target)); err != nil {
		return fmt.Errorf("failed to copy image file: %w", err)
	}

	// update simplestreams index
	log.FromContext(ctx).Info("Updating streams/v1/index.json")
	if indexJSON, err := json.Marshal(i.Index); err != nil {
		return fmt.Errorf("failed to encode new streams/v1/index.json: %w", err)
	} else if err := os.WriteFile(filepath.Join(i.rootDir, "streams", "v1", "index.json"), indexJSON, 0644); err != nil {
		return fmt.Errorf("failed to write streams/v1/index.json: %w", err)
	}

	// update products index
	log.FromContext(ctx).Info("Updating streams/v1/images.json")
	if productsJSON, err := json.Marshal(i.Products); err != nil {
		return fmt.Errorf("failed to encode new streams/v1/images.json: %w", err)
	} else if err := os.WriteFile(filepath.Join(i.rootDir, "streams", "v1", "images.json"), productsJSON, 0644); err != nil {
		return fmt.Errorf("failed to write streams/v1/images.json: %w", err)
	}

	return nil
}
