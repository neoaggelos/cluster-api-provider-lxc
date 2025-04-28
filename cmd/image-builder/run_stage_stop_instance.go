package main

import (
	"context"
	"fmt"
)

type stageStopInstance struct{}

func (*stageStopInstance) name() string { return "stop-instance" }

// incus stop capn-builder
func (*stageStopInstance) run(ctx context.Context) error {
	if err := client.StopInstance(ctx, cfg.instanceName); err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	return nil
}
