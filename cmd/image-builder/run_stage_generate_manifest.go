package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/static"
)

type stageGenerateManifest struct{}

func (*stageGenerateManifest) name() string { return "generate-manifest" }

// cat generate-manifest.sh | incus exec capn-builder -- bash -s
func (*stageGenerateManifest) run(ctx context.Context) error {
	log.FromContext(ctx).V(1).Info("Running generate-manifest.sh script")

	var stdout, stderr io.Writer
	if log.FromContext(ctx).V(4).Enabled() {
		stdout = os.Stdout
		stderr = os.Stderr
	}

	stdin := bytes.NewBufferString(static.GenerateManifestScript())
	if err := client.RunCommand(ctx, cfg.instanceName, []string{"bash", "-s"}, stdin, stdout, stderr); err != nil {
		return fmt.Errorf("failed to run generate-manifest.sh script: %w", err)
	}

	return nil
}
