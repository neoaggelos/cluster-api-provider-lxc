package stage

import (
	"context"
	"fmt"
	"slices"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Run runs a list of stages, with optional filtering criteria.
func Run(ctx context.Context, skipStages, onlyStages []string, dryRun bool, stages ...Stage) error {
	for idx, stage := range stages {
		ctx := log.IntoContext(ctx, log.FromContext(ctx).WithValues("stage.name", stage.Name, "stage.index", fmt.Sprintf("%d/%d", idx+1, len(stages))))

		if dryRun {
			log.FromContext(ctx).Info("Skipping stage", "dry-run", true)
			continue
		}

		if slices.Contains(skipStages, stage.Name) {
			log.FromContext(ctx).Info("Skipping stage", "skip", skipStages)
			continue
		}

		if len(onlyStages) > 0 && !slices.Contains(onlyStages, stage.Name) {
			log.FromContext(ctx).Info("Skipping stage", "only", onlyStages)
			continue
		}

		log.FromContext(ctx).Info("Starting stage")
		if err := stage.Action(ctx); err != nil {
			return fmt.Errorf("failure during stage %q: %w", stage.Name, err)
		}
		log.FromContext(ctx).Info("Completed stage")
	}

	return nil
}
