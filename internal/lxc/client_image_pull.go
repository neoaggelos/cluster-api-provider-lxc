package lxc

import (
	"context"
	"fmt"

	incus "github.com/lxc/incus/v7/client"
	"github.com/lxc/incus/v7/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PullImage pulls a remote image.
func (c *Client) PullImage(ctx context.Context, imageFamily ImageFamily) error {
	image, err := imageFamily.For(c.GetServerName())
	if err != nil {
		return fmt.Errorf("unsupported image: %w", err)
	}

	log.FromContext(ctx).V(4).Info("Pulling image", "server", image.Server, "name", image.Alias, "protocol", image.Protocol)
	return c.WaitForOperation(ctx, "PullImage", func() (incus.Operation, error) {
		return c.CreateImage(api.ImagesPost{
			Source: &api.ImagesPostSource{
				Type:        "image",
				ImageSource: image.ImageSource(),
			},
		}, nil)
	})
}
