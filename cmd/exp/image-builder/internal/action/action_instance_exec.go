package action

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// ExecInstance is an Action that executes a bash script on the instance, with an optional list of arguments.
func ExecInstance(lxcClient *lxc.Client, name string, bashScript string, args ...string) Action {
	return func(ctx context.Context) error {
		var stdout, stderr io.Writer
		if log.FromContext(ctx).V(4).Enabled() {
			stdout = os.Stdout
			stderr = os.Stderr
		}
		stdin := bytes.NewBufferString(bashScript)

		log.FromContext(ctx).V(1).Info("Running script", "args", args)
		if err := lxcClient.RunCommand(ctx, name, append([]string{"bash", "-s", "--"}, args...), stdin, stdout, stderr); err != nil {
			return fmt.Errorf("failed to run script: %w", err)
		}

		return nil
	}
}
