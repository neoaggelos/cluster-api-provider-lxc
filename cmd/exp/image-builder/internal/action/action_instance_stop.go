package action

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// StopInstance is an Action that stops the instance with specified name.
func StopInstance(lxcClient *lxc.Client, name string) Action {
	return func(ctx context.Context) error {
		log.FromContext(ctx).V(1).Info("Stopping instance")
		if err := lxcClient.WaitForStopInstance(ctx, name); err != nil {
			return fmt.Errorf("failed to stop instance: %w", err)
		}
		return nil
	}
}
