package action

import (
	"context"
	"time"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Wait is an Action that blocks for the specified amount of time.
func Wait(duration time.Duration) Action {
	return func(ctx context.Context, _ *lxc.Client) error {
		log.FromContext(ctx).V(1).Info("Waiting for interval", "timeout", duration)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(duration):
			return nil
		}
	}
}
