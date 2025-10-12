package action

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// qemuCompressImage accepts raw bytes of a qcow2 image, compresses with `qemu-img convert -c` and returns the raw bytes of the compressed image
func qemuCompressImage(ctx context.Context, raw []byte) ([]byte, error) {
	log.FromContext(ctx).V(2).Info("Compressing rootfs.img, this might take a while", "uncompressed", len(raw))
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	if err := os.WriteFile(filepath.Join(tmpDir, "uncompressed.qcow2"), raw, 0644); err != nil {
		return nil, fmt.Errorf("failed to write uncompressed rootfs to temporary file: %w", err)
	}

	// attempt to use qemu-img from path, or fallbcak to /opt/incus/bin/qemu-img
	var extraEnv []string
	qemuImg, err := exec.LookPath("qemu-img")
	if err != nil {
		qemuImg = "/opt/incus/bin/qemu-img"
		extraEnv = append(extraEnv, "LD_LIBRARY_PATH=/opt/incus/lib")
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, qemuImg, "convert", "-O", "qcow2", "-c", filepath.Join(tmpDir, "uncompressed.qcow2"), filepath.Join(tmpDir, "compressed.qcow2"))
	cmd.Env = append(os.Environ(), extraEnv...)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("qemu-img convert -c command failed with stderr=%q: %w", stderr.String(), err)
	}

	b, err := os.ReadFile(filepath.Join(tmpDir, "compressed.qcow2"))
	if err != nil {
		return nil, fmt.Errorf("failed to read rootfs after compression: %w", err)
	}

	log.FromContext(ctx).Info("Compressed rootfs.img", "uncompressed", len(raw), "compressed", len(b))
	return b, nil
}

func compressUnifiedImageTarballRootfs(ctx context.Context, path string) error {
	log.FromContext(ctx).V(2).Info("Updating unified tarball to compress rootfs.img", "path", path)

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("failed to create a temporary directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	f, err := os.Open(path)
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

	tmpFile := filepath.Join(tmpDir, "image.tar.gz")

	log.FromContext(ctx).V(2).Info("Create temporary tarball", "path", tmpFile)
	of, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to open: %w", err)
	}
	defer func() {
		_ = of.Close()
	}()
	gzWriter := gzip.NewWriter(of)
	defer func() {
		_ = gzWriter.Close()
	}()
	tarWriter := tar.NewWriter(gzWriter)
	defer func() {
		_ = tarWriter.Close()
	}()

	for {
		if hdr, err := tarReader.Next(); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read tar.gz archive: %w", err)
		} else if hdr.Name != "rootfs.img" {
			log.FromContext(ctx).V(2).Info("Copying file to temporary tarball", "path", tmpFile, "file", hdr.Name)
			if err := tarWriter.WriteHeader(hdr); err != nil {
				return fmt.Errorf("failed to write header for %q: %w", hdr.Name, err)
			} else if _, err := io.Copy(tarWriter, tarReader); err != nil {
				return fmt.Errorf("failed to copy file %q: %w", hdr.Name, err)
			}
		} else {
			log.FromContext(ctx).V(2).Info("Copying rootfs.img to temporary tarball", "path", tmpFile, "file", hdr.Name)
			raw, err := io.ReadAll(tarReader)
			if err != nil {
				return fmt.Errorf("failed to read rootfs.img: %w", err)
			}
			compressed, err := qemuCompressImage(ctx, raw)
			if err != nil {
				return fmt.Errorf("failed to compress rootfs.img from tarball: %w", err)
			}
			hdr.Size = int64(len(compressed))
			if err := tarWriter.WriteHeader(hdr); err != nil {
				return fmt.Errorf("failed to write header for compressed rootfs.img: %w", err)
			} else if _, err := tarWriter.Write(compressed); err != nil {
				return fmt.Errorf("failed to copy compressed rootfs.img: %w", err)
			}
		}
	}

	log.FromContext(ctx).V(2).Info("Update tarball with compressed rootfs.img", "path", path)
	if err := os.Rename(tmpFile, path); err != nil {
		return fmt.Errorf("failed to move temporary file: %w", err)
	}

	return nil
}
