package main

import (
	"context"
	"fmt"
)

type removeInstanceStage struct{}

func (*removeInstanceStage) name() string { return "remove-instance" }

// incus rm capn-builder --force
func (*removeInstanceStage) run(ctx context.Context) error {
	if err := client.ForceRemoveInstance(ctx, cfg.instanceName); err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}
	return nil
}
