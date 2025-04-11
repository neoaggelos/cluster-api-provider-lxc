package incus

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/lxc/cluster-api-provider-incus/internal/cloudinit"
)

// CheckCloudInitStatus checks the cloud-init status of an instance.
//
// CheckCloudInitStatus returns one of the following:
//
// - (cloudinit.StatusDone, nil)
// - (cloudinit.StatusRunning, nil)
// - (cloudinit.StatusError, nil)
// - (cloudinit.StatusUnknown, <error describing why status is unknown>)
func (c *Client) CheckCloudInitStatus(ctx context.Context, name string) (result cloudinit.Status, rerr error) {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithValues("instance", name, "path", "/var/lib/cloud/data/status.json"))

	reader, _, err := c.Client.GetInstanceFile(name, "/var/lib/cloud/data/status.json")
	if err != nil || reader == nil {
		return cloudinit.StatusUnknown, fmt.Errorf("failed to read cloud-init status file: failed to GetInstanceFile: %w", err)
	}
	defer func() { _ = reader.Close() }()

	defer func() {
		log.FromContext(ctx).V(2).WithValues("result", result, "error", rerr).Info("Checking cloud-init status on instance")
	}()
	return cloudinit.ParseStatus(reader)
}
