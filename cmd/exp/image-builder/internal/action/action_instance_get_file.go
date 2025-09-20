package action

import (
	"context"
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// CopyInstanceFile is an Action that copies the contents of an instance file on the local disk.
func CopyInstanceFile(lxcClient *lxc.Client, name string, instanceFile string, outputFile string) Action {
	return func(ctx context.Context) error {
		of, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		log.FromContext(ctx).V(1).Info("Copy instance file", "source", instanceFile, "destination", outputFile)

		if reader, _, err := lxcClient.GetInstanceFile(name, instanceFile); err != nil {
			return fmt.Errorf("failed to read file %q from instance %q: %w", instanceFile, name, err)
		} else if _, err := io.Copy(of, reader); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}

		return nil
	}
}
