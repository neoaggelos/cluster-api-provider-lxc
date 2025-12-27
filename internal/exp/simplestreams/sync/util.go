package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func calculateFileSha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	hash := sha256.New()
	if _, err = io.Copy(hash, f); err != nil {
		return "", fmt.Errorf("failed to calculate sha256 sum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
