package index

import (
	"archive/tar"
	"bytes"
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

func (i *Index) importVirtualMachineUnifiedTarball(ctx context.Context, imagePath string, aliases []string, incus bool, lxd bool) error {
	log.FromContext(ctx).Info("Importing virtual-machine image", "image", imagePath)

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

	var outMetadataBuffer bytes.Buffer
	gzWriter := gzip.NewWriter(&outMetadataBuffer)
	tarWriter := tar.NewWriter(gzWriter)

	var metadata api.ImageMetadata
	var outRootfs []byte
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

			if err := tarWriter.WriteHeader(hdr); err != nil {
				return fmt.Errorf("failed to write header for metadata.yaml: %w", err)
			}
			if _, err := tarWriter.Write(b); err != nil {
				return fmt.Errorf("failed to write metadata.yaml: %w", err)
			}
		} else if strings.HasPrefix(hdr.Name, "templates/") {
			if err := tarWriter.WriteHeader(hdr); err != nil {
				return fmt.Errorf("failed to write header for %q: %w", hdr.Name, err)
			}
			if _, err := io.Copy(tarWriter, tarReader); err != nil {
				return fmt.Errorf("failed to write %q: %w", hdr.Name, err)
			}
		} else if hdr.Name == "rootfs.img" {
			b, err := io.ReadAll(tarReader)
			if err != nil {
				return fmt.Errorf("failed to read rootfs.img from image: %w", err)
			}

			outRootfs = b
		}
	}

	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to generate metadata.tar: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to generate metadata.tar.gz: %w", err)
	}

	outMetadata := outMetadataBuffer.Bytes()
	if len(outMetadata) == 0 {
		return fmt.Errorf("no metadata found in image")
	}
	if len(outRootfs) == 0 {
		return fmt.Errorf("no rootfs.img found in image")
	}

	if metadata.Architecture == "" {
		return fmt.Errorf("no metadata.yaml found in image")
	}

	// we now have:
	//   * metadata: parsed instance metadata
	//   * outMetadata: metadata archive for vm image
	//   * outRootfs: compressed qcow2 rootfs for vm image

	info, err := getVirtualMachineImageInfo(outMetadata, outRootfs)
	if err != nil {
		return fmt.Errorf("failed to calculate size and sha256 of virtual machine image: %w", err)
	}

	productName := fmt.Sprintf("%s:%s:%s:%s", metadata.Properties["os"], metadata.Properties["release"], metadata.Properties["variant"], metadata.Properties["architecture"])
	versionName := time.Unix(metadata.CreationDate, 0).Format("200601021504")
	metadataTarget := filepath.Join("images", metadata.Properties["os"], metadata.Properties["release"], metadata.Properties["architecture"], fmt.Sprintf("%s.incus.tar.xz", info.MetaSha256))
	rootfsTarget := filepath.Join("images", metadata.Properties["os"], metadata.Properties["release"], metadata.Properties["architecture"], fmt.Sprintf("%s.disk-kvm.img", info.MetaSha256))

	if err := os.MkdirAll(filepath.Join(i.rootDir, filepath.Dir(rootfsTarget)), 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %w", err)
	}

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
		log.FromContext(ctx).Info("Adding rootfs product version item", "ftype", vmRootfsFTypeIncus, "path", rootfsTarget, "size", info.RootSize)
		newProductVersions.Items[vmRootfsFTypeIncus] = simplestreams.ProductVersionItem{
			FileType:   vmRootfsFTypeIncus,
			Size:       info.RootSize,
			Path:       rootfsTarget,
			HashSha256: info.RootSha256,
		}

		log.FromContext(ctx).Info("Adding metadata product version item", "ftype", vmMetadataFTypeIncus, "path", metadataTarget, "size", info.MetaSize)
		newProductVersions.Items[vmMetadataFTypeIncus] = simplestreams.ProductVersionItem{
			FileType:                 vmMetadataFTypeIncus,
			Size:                     info.MetaSize,
			Path:                     metadataTarget,
			HashSha256:               info.MetaSha256,
			CombinedSha256DiskKvmImg: info.CombinedSha256,
		}
	}
	if lxd {
		log.FromContext(ctx).Info("Adding rootfs product version item", "ftype", vmRootfsFTypeLXD, "path", rootfsTarget, "size", info.RootSize)
		newProductVersions.Items[vmRootfsFTypeLXD] = simplestreams.ProductVersionItem{
			FileType:   vmRootfsFTypeLXD,
			Size:       info.RootSize,
			Path:       rootfsTarget,
			HashSha256: info.RootSha256,
		}

		log.FromContext(ctx).Info("Adding metadata product version item", "ftype", vmMetadataFTypeLXD, "path", metadataTarget, "size", info.MetaSize)
		newProductVersions.Items[vmMetadataFTypeLXD] = simplestreams.ProductVersionItem{
			FileType:              vmMetadataFTypeLXD,
			Size:                  info.MetaSize,
			Path:                  metadataTarget,
			HashSha256:            info.MetaSha256,
			CombinedSha256DiskImg: info.CombinedSha256,
		}
	}

	product.Versions[versionName] = newProductVersions
	i.Products.Products[productName] = product

	// copy image
	log.FromContext(ctx).Info("Copying image rootfs into simplestreams index", "destination", filepath.Join(i.rootDir, rootfsTarget))
	if err := os.WriteFile(filepath.Join(i.rootDir, rootfsTarget), outRootfs, 0644); err != nil {
		return fmt.Errorf("failed to write rootfs: %w", err)
	}
	log.FromContext(ctx).Info("Copying image metadata into simplestreams index", "destination", filepath.Join(i.rootDir, metadataTarget))
	if err := os.WriteFile(filepath.Join(i.rootDir, metadataTarget), outMetadata, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
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
