package main

import (
	"context"
	"fmt"
)

type stopInstanceStage struct{}

func (*stopInstanceStage) name() string { return "stop-instance" }

// incus stop capn-builder
func (*stopInstanceStage) run(ctx context.Context) error {
	if err := client.StopInstance(ctx, cfg.instanceName); err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	return nil
}
