package index

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type containerImageInfo struct {
	Sha256 string
	Size   int64
}

func getContainerImageInfo(path string) (containerImageInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return containerImageInfo{}, fmt.Errorf("failed to stat: %w", err)
	}

	// get the image sha256
	f, err := os.Open(path)
	if err != nil {
		return containerImageInfo{}, fmt.Errorf("failed to open: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	hash := sha256.New()
	if _, err = io.Copy(hash, f); err != nil {
		return containerImageInfo{}, fmt.Errorf("failed to calculate sha256 sum: %w", err)
	}

	return containerImageInfo{
		Size:   stat.Size(),
		Sha256: fmt.Sprintf("%x", hash.Sum(nil)),
	}, nil
}

type virtualMachineImageInfo struct {
	MetaSize   int64
	MetaSha256 string

	RootSize   int64
	RootSha256 string

	CombinedSha256 string
}

func getVirtualMachineImageInfo(metadata []byte, rootfs []byte) (virtualMachineImageInfo, error) {
	info := virtualMachineImageInfo{
		MetaSize: int64(len(metadata)),
		RootSize: int64(len(rootfs)),
	}
	hash := sha256.New()
	if _, err := hash.Write(metadata); err != nil {
		return virtualMachineImageInfo{}, fmt.Errorf("failed to calculate metadata sha256 sum: %w", err)
	}
	info.MetaSha256 = fmt.Sprintf("%x", hash.Sum(nil))

	if _, err := hash.Write(rootfs); err != nil {
		return virtualMachineImageInfo{}, fmt.Errorf("failed to calculate combined sha256 sum: %w", err)
	}
	info.CombinedSha256 = fmt.Sprintf("%x", hash.Sum(nil))

	hash.Reset()
	if _, err := hash.Write(rootfs); err != nil {
		return virtualMachineImageInfo{}, fmt.Errorf("failed to calculate rootfs sha256 sum: %w", err)
	}
	info.RootSha256 = fmt.Sprintf("%x", hash.Sum(nil))

	return info, nil
}

func copyFile(source, destination string) error {
	f, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("failed to create directory for destination: %w", err)
	}
	of, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to open destination: %w", err)
	}

	if n, err := io.Copy(of, f); err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	} else if err := of.Truncate(n); err != nil {
		return fmt.Errorf("failed to truncate output file: %w", err)
	}

	if err := of.Close(); err != nil {
		return fmt.Errorf("failed to flush file: %w", err)
	}

	return nil
}
