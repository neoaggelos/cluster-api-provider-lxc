package main

import "context"

type preRunCommandsStage struct{}

func (*preRunCommandsStage) name() string { return "pre-run-commands" }

func (*preRunCommandsStage) run(ctx context.Context) error {
	return nil
}
