package main

import (
	"context"
	"fmt"
	"slices"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type stage interface {
	name() string
	run(ctx context.Context) error
}

func runStages(stages ...stage) error {
	for idx, stage := range stages {
		ctx := log.IntoContext(gCtx, log.FromContext(gCtx).WithValues("stage.name", stage.name(), "stage.index", fmt.Sprintf("%d/%d", idx+1, len(stages))))

		if slices.Contains(cfg.skipStages, stage.name()) {
			log.FromContext(ctx).Info("Skipping stage", "skip", cfg.skipStages)
			continue
		}

		if len(cfg.onlyStages) > 0 && !slices.Contains(cfg.onlyStages, stage.name()) {
			log.FromContext(ctx).Info("Skipping stage", "only", cfg.onlyStages)
			continue
		}

		log.FromContext(ctx).Info("Starting stage")
		if err := stage.run(ctx); err != nil {
			return fmt.Errorf("failure during stage %q: %w", stage.name(), err)
		}
		log.FromContext(ctx).Info("Completed stage")
	}

	return nil
}
