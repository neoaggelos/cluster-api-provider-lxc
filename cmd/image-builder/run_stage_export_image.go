package main

import (
	"context"
	"fmt"
	"os"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/ioprogress"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type stageExportImage struct{}

func (*stageExportImage) name() string { return "export-image" }

// incus image export capn-builder-image
func (s *stageExportImage) run(ctx context.Context) (rerr error) {
	image, _, err := client.Client.GetImageAlias(cfg.imageAlias)
	if err != nil {
		return fmt.Errorf("failed to find image for alias %q: %w", cfg.imageAlias, err)
	}

	output, err := os.Create(cfg.outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		_ = output.Close()
		if rerr != nil {
			_ = os.Remove(cfg.outputFile)
		}
	}()

	resp, err := client.Client.GetImageFile(image.Target, incus.ImageFileRequest{
		MetaFile: output,
		ProgressHandler: func(progress ioprogress.ProgressData) {
			log.FromContext(ctx).V(2).WithValues("progress", progress.Text).Info("Downloading image")
		},
	})
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}

	log.FromContext(ctx).V(1).WithValues("image", resp).Info("Downloaded image")
	if err := output.Truncate(resp.MetaSize); err != nil {
		return fmt.Errorf("failed to truncate output file: %w", err)
	}

	return nil
}
