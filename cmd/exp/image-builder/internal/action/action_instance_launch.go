package action

import (
	"context"
	"fmt"
	"time"

	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
)

// LaunchInstance is an Action that launches an instance and waits for its agent to come up.
func LaunchInstance(lxcClient *lxc.Client, name string, launchOpts *lxc.LaunchOptions) Action {
	return func(ctx context.Context) error {
		// set size of root volume to 5GB, otherwise publishing virtual machine images takes a very long time.
		pools, err := lxcClient.GetStoragePools()
		if err != nil {
			return fmt.Errorf("failed to list storage pools: %w", err)
		}

		for _, pool := range pools {
			if pool.Status == api.StoragePoolStatusCreated {
				launchOpts = launchOpts.WithDevices(map[string]map[string]string{
					"root": {
						"type": "disk",
						"pool": pool.Name,
						"path": "/",
						"size": "5GiB",
					},
				})
			}
		}

		log.FromContext(ctx).V(1).Info("Launching instance")
		if _, err := lxcClient.WaitForLaunchInstance(ctx, name, launchOpts); err != nil {
			return fmt.Errorf("failed to launch instance: %w", err)
		}

		log.FromContext(ctx).V(1).Info("Waiting for instance agent to come up")
		waitInstanceCh := make(chan error, 1)
		go func() {
			select {
			case <-ctx.Done():
				waitInstanceCh <- ctx.Err()
			case <-time.After(5 * time.Minute):
				waitInstanceCh <- fmt.Errorf("timed out after 5 minutes")
			}
		}()
		go func() {
			for lxcClient.RunCommand(ctx, name, []string{"echo", "hi"}, nil, nil, nil) != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
			}
			waitInstanceCh <- nil
		}()

		if err := <-waitInstanceCh; err != nil {
			return fmt.Errorf("failed to wait for instance agent to come up: %w", err)
		}

		return nil
	}
}
