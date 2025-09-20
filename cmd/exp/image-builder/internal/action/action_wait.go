package action

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Wait is an Action that blocks for the specified amount of time.
func Wait(duration time.Duration) Action {
	return func(ctx context.Context) error {
		log.FromContext(ctx).V(1).Info("Waiting for interval", "timeout", duration)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(duration):
			return nil
		}
	}
}
