package action

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// DeleteInstance is an Action that deletes the instance with specified name.
func DeleteInstance(name string) Action {
	return func(ctx context.Context, lxcClient *lxc.Client) error {
		log.FromContext(ctx).V(1).Info("Deleting instance")
		if err := lxcClient.WaitForDeleteInstance(ctx, name); err != nil {
			return fmt.Errorf("failed to delete instance: %w", err)
		}
		return nil
	}
}
