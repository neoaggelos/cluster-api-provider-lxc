package incus

import (
	"context"
	"fmt"
	"io"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
)

// RunCommand on a specified instance, and allow retrieving the stdout and stderr. It returns an error if execution failed.
func (c *Client) RunCommand(ctx context.Context, instanceName string, command []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	var status int
	if err := c.wait(ctx, "ExecInstance", func() (incus.Operation, error) {
		op, err := c.Client.ExecInstance(instanceName, api.InstanceExecPost{
			Command:   command,
			WaitForWS: true,
		}, &incus.InstanceExecArgs{
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		})
		if err != nil {
			return nil, err
		}

		_, _ = op.AddHandler(func(o api.Operation) {
			if rc, ok := o.Metadata["return"].(float64); ok {
				status = int(rc)
			}
		})
		return op, nil
	}); err != nil {
		return err
	}

	if status != 0 {
		return fmt.Errorf("command failed with exit code %v", status)
	}
	return nil
}
