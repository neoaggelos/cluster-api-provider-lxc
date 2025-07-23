package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
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
	"sigs.k8s.io/yaml"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

func importVirtualMachineImage(index simplestreams.Stream, products simplestreams.Products) error {
	log.Info("Importing virtual-machine image", "image", importCfg.imagePath)

	f, err := os.Open(importCfg.imagePath)
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
	//   * outRootfs: qcow2 rootfs for vm image

	info, err := getVirtualMachineImageInfo(outMetadata, outRootfs)
	if err != nil {
		return fmt.Errorf("failed to calculate size and sha256 of virtual machine image: %w", err)
	}

	productName := fmt.Sprintf("%s:%s:%s:%s", metadata.Properties["os"], metadata.Properties["release"], metadata.Properties["variant"], metadata.Properties["architecture"])
	versionName := time.Unix(metadata.CreationDate, 0).Format("200601021504")

	metadataFType := vmMetadataFTypeByServer[importCfg.serverType]
	metadataTarget := filepath.Join("images", metadata.Properties["os"], metadata.Properties["release"], fmt.Sprintf("%s.%s", info.MetaSha256, metadataFType))

	rootfsFType := vmRootfsFTypeByServer[importCfg.serverType]
	rootfsTarget := filepath.Join("images", metadata.Properties["os"], metadata.Properties["release"], fmt.Sprintf("%s.%s", info.MetaSha256, rootfsFType))

	log.Info("Adding product version item", "product", productName, "version", versionName, "info", info)

	// update index
	if !slices.Contains(index.Index["images"].Products, productName) {
		log.Info("Adding product in streams/v1/index.json")

		newImages := index.Index["images"]
		newImages.Products = append(newImages.Products, productName)
		slices.Sort(newImages.Products)
		index.Index["images"] = newImages
	}

	// update product versions
	var product simplestreams.Product
	if existingProduct, ok := products.Products[productName]; ok {
		product = existingProduct
	} else {
		log.Info("Creating product", "product", productName)

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

	if len(importCfg.imageAliases) > 0 {
		log.Info("Setting product aliases", "product", productName, "aliases", importCfg.imageAliases)

		product.Aliases = strings.Join(importCfg.imageAliases, ",")
	}

	if product.Versions == nil {
		product.Versions = make(map[string]simplestreams.ProductVersion)
	}
	newProductVersions := product.Versions[versionName]
	if newProductVersions.Items == nil {
		newProductVersions.Items = make(map[string]simplestreams.ProductVersionItem)
	}

	log.Info("Adding rootfs product version item", "ftype", rootfsFType, "path", rootfsTarget, "size", info.RootSize)
	newProductVersions.Items[rootfsFType] = simplestreams.ProductVersionItem{
		FileType:   rootfsFType,
		HashSha256: info.RootSha256,
		Size:       info.RootSize,
		Path:       rootfsTarget,
	}

	log.Info("Adding metadata product version item", "ftype", metadataFType, "path", metadataTarget, "size", info.MetaSize)
	metadataProductVersionItem := simplestreams.ProductVersionItem{
		FileType:   metadataFType,
		HashSha256: info.MetaSha256,
		Size:       info.MetaSize,
		Path:       metadataTarget,
	}

	switch importCfg.serverType {
	case lxc.Incus:
		metadataProductVersionItem.CombinedSha256DiskKvmImg = info.CombinedSha256
	case lxc.LXD:
		metadataProductVersionItem.CombinedSha256DiskImg = info.CombinedSha256
	}

	newProductVersions.Items[metadataFType] = metadataProductVersionItem
	product.Versions[versionName] = newProductVersions
	products.Products[productName] = product

	// copy image
	log.Info("Copying image rootfs into simplestreams index", "destination", filepath.Join(cfg.rootDir, rootfsTarget))
	if err := os.WriteFile(filepath.Join(cfg.rootDir, rootfsTarget), outRootfs, 0644); err != nil {
		return fmt.Errorf("failed to write rootfs: %w", err)
	}
	log.Info("Copying image metadata into simplestreams index", "destination", filepath.Join(cfg.rootDir, metadataTarget))
	if err := os.WriteFile(filepath.Join(cfg.rootDir, metadataTarget), outMetadata, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// update simplestreams index
	log.Info("Updating streams/v1/index.json")
	if indexJSON, err := json.Marshal(index); err != nil {
		return fmt.Errorf("failed to encode new streams/v1/index.json: %w", err)
	} else if err := os.WriteFile(filepath.Join(cfg.rootDir, "streams", "v1", "index.json"), indexJSON, 0644); err != nil {
		return fmt.Errorf("failed to write streams/v1/index.json: %w", err)
	}

	// update products index
	log.Info("Updating streams/v1/images.json")
	if productsJSON, err := json.Marshal(products); err != nil {
		return fmt.Errorf("failed to encode new streams/v1/images.json: %w", err)
	} else if err := os.WriteFile(filepath.Join(cfg.rootDir, "streams", "v1", "images.json"), productsJSON, 0644); err != nil {
		return fmt.Errorf("failed to write streams/v1/images.json: %w", err)
	}

	return nil
}
