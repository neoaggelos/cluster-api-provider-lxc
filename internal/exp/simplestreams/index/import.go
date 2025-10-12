package index

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// NOTE(neoaggelos): GitHub Actions always zips the artifacts, so images will be a .zip file containing
// the .tar.gz unified tarball. Handle this case automatically.
func maybeExtractUnifiedTarballFromZIP(path string) (string, func() error, error) {
	if !strings.HasSuffix(path, ".zip") {
		return "", nil, nil
	}
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		// if not a zip archive, do nothing
		return "", nil, nil
	}
	defer func() {
		_ = zipReader.Close()
	}()
	if len(zipReader.File) != 1 || !strings.HasSuffix(zipReader.File[0].Name, ".tar.gz") {
		// if not contains a single unified tarball tar.gz, do nothing
		return "", nil, nil
	}

	f, err := zipReader.Open(zipReader.File[0].Name)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read unified tarball from zip file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	ofName := filepath.Join(tmpDir, "image.tar.gz")
	of, err := os.Create(ofName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		_ = of.Close()
	}()
	if _, err := io.Copy(of, f); err != nil {
		return "", nil, fmt.Errorf("failed to extract unified tarball: %w", err)
	}

	return ofName, func() error { return os.RemoveAll(tmpDir) }, nil
}

// ImportImage imports a container or a virtual machine image into the simplestreams index.
func (i *Index) ImportImage(ctx context.Context, imageType string, imagePath string, aliases []string, incus bool, lxd bool) error {
	if tarballPath, cleanup, err := maybeExtractUnifiedTarballFromZIP(imagePath); err != nil {
		return fmt.Errorf("failed to extract unified tarball from archive: %w", err)
	} else if tarballPath != "" {
		log.FromContext(ctx).Info("Detected zip archive with unified tarball image", "archive", imagePath, "tarball", tarballPath)
		defer func() {
			_ = cleanup()
		}()
		imagePath = tarballPath
	}

	switch imageType {
	case lxc.Container:
		return i.importContainerUnifiedTarball(ctx, imagePath, aliases, incus, lxd)
	case lxc.VirtualMachine:
		return i.importVirtualMachineUnifiedTarball(ctx, imagePath, aliases, incus, lxd)
	default:
		return fmt.Errorf("unknown image type %q", imageType)
	}
}
