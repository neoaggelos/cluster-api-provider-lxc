package main

import "context"

type stagePostRunCommands struct{}

func (*stagePostRunCommands) name() string { return "post-run-commands" }

func (*stagePostRunCommands) run(ctx context.Context) error {
	return nil
}
