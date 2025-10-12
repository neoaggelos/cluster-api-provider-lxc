package action

import (
	"context"
	"fmt"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// Chain is a meta Action that runs multiple Action one after the other.
func Chain(actions ...Action) Action {
	return func(ctx context.Context, lxcClient *lxc.Client) error {
		for idx, action := range actions {
			if err := action(ctx, lxcClient); err != nil {
				return fmt.Errorf("action %d/%d failed: %w", idx, len(actions), err)
			}
		}
		return nil
	}
}
