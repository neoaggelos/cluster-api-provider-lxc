package main

import (
	"context"
	"fmt"
)

type stageRemoveInstance struct{}

func (*stageRemoveInstance) name() string { return "remove-instance" }

// incus rm capn-builder --force
func (*stageRemoveInstance) run(ctx context.Context) error {
	if err := client.ForceRemoveInstance(ctx, cfg.instanceName); err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}
	return nil
}
