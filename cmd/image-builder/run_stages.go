package main

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type stage interface {
	name() string
	run(ctx context.Context) error
}

func runStages(stages ...stage) error {
	for idx, stage := range stages {
		ctx := log.IntoContext(gCtx, log.FromContext(gCtx).WithValues("stage.name", stage.name(), "stage.index", fmt.Sprintf("%d/%d", idx+1, len(stages))))

		log.FromContext(ctx).Info("Starting stage")
		if err := stage.run(ctx); err != nil {
			return fmt.Errorf("failure during stage %q: %w", stage.name(), err)
		}
		log.FromContext(ctx).Info("Completed stage")
	}

	return nil
}
