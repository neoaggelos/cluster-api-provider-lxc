package action

import (
	"context"
	"fmt"
	"os"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/ioprogress"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// ExportImage is an Action that downloads a unified image tarball and saves to a local file.
func ExportImage(imageAliasName string, outputFile string) Action {
	return func(ctx context.Context, lxcClient *lxc.Client) (rerr error) {
		image, _, err := lxcClient.GetImageAlias(imageAliasName)
		if err != nil {
			return fmt.Errorf("failed to find image for alias %q: %w", imageAliasName, err)
		}

		output, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			_ = output.Close()
			if rerr != nil {
				_ = os.Remove(outputFile)
			}
		}()

		log.FromContext(ctx).V(1).Info("Downloading image")
		resp, err := lxcClient.GetImageFile(image.Target, incus.ImageFileRequest{
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

		// NOTE(neoaggelos): https://github.com/lxc/incus/commit/76804eedd6ac061fb4d974806be65ee78fb62c74
		// Incus no longer compresses rootfs when exporting a unified tarball, so we have to
		if lxcClient.GetServerName() == lxc.Incus && image.Type == lxc.VirtualMachine {
			if err := compressUnifiedImageTarballRootfs(ctx, outputFile); err != nil {
				return fmt.Errorf("failed to compress rootfs.img in unified tarball: %w", err)
			}
		}

		return nil
	}
}
