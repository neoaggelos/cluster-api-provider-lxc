package action

import (
	"context"
	"fmt"
)

// Chain is a meta Action that runs multiple Action one after the other.
func Chain(actions ...Action) Action {
	return func(ctx context.Context) error {
		for idx, action := range actions {
			if err := action(ctx); err != nil {
				return fmt.Errorf("action %d/%d failed: %w", idx, len(actions), err)
			}
		}
		return nil
	}
}
