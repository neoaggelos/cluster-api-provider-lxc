package lxc

import (
	"context"
	"fmt"
	"strings"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func loggingProgressHandler(ctx context.Context, name string) func(api.Operation) {
	idx := 0
	return func(o api.Operation) {
		// reduce log noise for command execute operations unless --v=5 or higher
		if _, ok := o.Metadata["command"]; ok && !log.FromContext(ctx).V(5).Enabled() {
			delete(o.Metadata, "environment")
			delete(o.Metadata, "fds")
		}

		log := log.FromContext(ctx).WithValues("operation.name", name, "operation.uuid", o.ID, "operation.metadata", o.Metadata, "operation.status", o.Status)

		switch {
		case o.StatusCode == api.Failure:
			log.Error(fmt.Errorf("%v", o.Err), "Operation failed")
		case o.StatusCode.IsFinal():
			log.V(2).Info("Operation completed")
		default:
			// use log level 2 for every 5th message
			level := 4
			if idx%5 == 0 {
				level = 2
			}
			idx++

			log.V(level).Info("Operation in progress")
		}
	}
}

func (c *Client) WaitForOperation(ctx context.Context, name string, f func() (incus.Operation, error)) error {
	op, err := f()
	if err != nil {
		return fmt.Errorf("failed to %s: %w", name, err)
	}

	// configure progress handler
	handler := c.progressHandler
	if handler == nil {
		handler = loggingProgressHandler(ctx, name)
	}
	target, _ := op.AddHandler(handler)
	defer func() {
		_ = op.RemoveHandler(target)
	}()

	if err := op.WaitContext(ctx); err != nil && !strings.Contains(err.Error(), "Operation not found") {
		return fmt.Errorf("failed to wait for %s operation: %w", name, err)
	}
	return nil
}
