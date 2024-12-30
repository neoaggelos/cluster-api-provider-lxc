package incus

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/neoaggelos/cluster-api-provider-lxc/internal/cloudinit"
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
		log.FromContext(ctx).Error(err, "Could not read cloud-init status file")
		return cloudinit.StatusUnknown, nil
	}

	defer func() {
		log.FromContext(ctx).V(4).WithValues("result", result, "error", rerr).Info("Checking cloud-init status on instance")
	}()
	defer reader.Close()
	return cloudinit.ParseStatus(reader)
}
