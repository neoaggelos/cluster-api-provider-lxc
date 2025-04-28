package main

import "context"

type postRunCommandsStage struct{}

func (*postRunCommandsStage) name() string { return "post-run-commands" }

func (*postRunCommandsStage) run(ctx context.Context) error {
	return nil
}
