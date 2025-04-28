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

type stageCleanupInstance struct{}

func (*stageCleanupInstance) name() string { return "cleanup-instance" }

// cat cleanup-instance.sh | incus exec capn-builder -- bash -s
func (*stageCleanupInstance) run(ctx context.Context) error {
	log.FromContext(ctx).V(1).Info("Running cleanup-instance.sh script")

	var stdout, stderr io.Writer
	if log.FromContext(ctx).V(4).Enabled() {
		stdout = os.Stdout
		stderr = os.Stderr
	}

	stdin := bytes.NewBufferString(static.CleanupInstanceScript())
	if err := client.RunCommand(ctx, cfg.instanceName, []string{"bash", "-s"}, stdin, stdout, stderr); err != nil {
		return fmt.Errorf("failed to run cleanup-instance.sh script: %w", err)
	}

	return nil
}
