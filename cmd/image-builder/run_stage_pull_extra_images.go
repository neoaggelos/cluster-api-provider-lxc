package main

import (
	"context"
	"fmt"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type stagePullExtraImages struct{}

func (*stagePullExtraImages) name() string { return "pull-extra-images" }

// incus exec capn-builder -- crictl pull $image
func (*stagePullExtraImages) run(ctx context.Context) error {
	for _, image := range kubeadmCfg.pullExtraImages {
		log.FromContext(ctx).V(1).WithValues("image", image).Info("Pulling image")
		if err := client.RunCommand(ctx, cfg.instanceName, []string{"crictl", "pull", image}, nil, os.Stdout, os.Stderr); err != nil {
			return fmt.Errorf("failed to pull image %q: %w", image, err)
		}
	}

	return nil
}
