package main

import "context"

type stagePreRunCommands struct{}

func (*stagePreRunCommands) name() string { return "pre-run-commands" }

func (*stagePreRunCommands) run(ctx context.Context) error {
	return nil
}
