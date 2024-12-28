package incus

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	incus "github.com/lxc/incus/v6/client"
	"github.com/lxc/incus/v6/shared/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CloudInitStatus defines different possible values for the status of cloud-init in an instance.
type CloudInitStatus string

const (
	// CloudInitStatusUnknown means that it was impossible to retrieve the status of cloud-init,
	// perhaps because the "cloud-init status" command was not successful.
	CloudInitStatusUnknown CloudInitStatus = "Unknown"

	// CloudInitStatusDone means that cloud-init completed successfully.
	CloudInitStatusDone CloudInitStatus = "Done"

	// CloudInitStatusRunning means that cloud-init is stil running on the instance.
	CloudInitStatusRunning CloudInitStatus = "Running"

	// CloudInitStatusError means that cloud-init failed.
	CloudInitStatusError CloudInitStatus = "Error"
)

// CheckCloudInitStatus checks the cloud-init status of an instance.
//
// CheckCloudInitStatus returns one of the following tuples:
//
// - (CloudInitStatusDone, nil)
// - (CloudInitStatusRunning, nil)
// - (CloudInitStatusError, nil)
// - (CloudInitStatusUnknown, <error describing why status is unknown>)
func (c *Client) CheckCloudInitStatus(ctx context.Context, name string) (result CloudInitStatus, rerr error) {
	var stdout, stderr bytes.Buffer
	err := c.wait(ctx, name, func() (incus.Operation, error) {
		return c.Client.ExecInstance(name, api.InstanceExecPost{
			Command:      []string{"cloud-init", "status"},
			RecordOutput: true,
		}, &incus.InstanceExecArgs{
			Stdout: &stdout,
			Stderr: &stderr,
		})
	})

	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		/*
			// NOTE: op.Wait() for an ExecInstance command may return errors because it could not delete files
			// and mask actual errors or break the command output.
			//
			// For this reason, we ignore the error if it contains "no such file or directory".

			{
				status, err := client.CheckCloudInitStatus(context.TODO(), "t1")
				g.Expect(err).To(BeNil())
			}

			raised the following test failure:

			--- FAIL: TestCloudInitStatus (0.21s)
					client_test.go:47:
							Expected
									<*fmt.wrapError | 0xc0002a24e0>:
									cloud-init status command failed: failed to t1: "remove /var/lib/incus/containers/t1/exec-output/exec_ed6faba9-6647-4f38-83ad-b19bbce584a4.stdout: no such file or directory"
									{
											msg: "cloud-init status command failed: failed to t1: \"remove /var/lib/incus/containers/t1/exec-output/exec_ed6faba9-6647-4f38-83ad-b19bbce584a4.stdout: no such file or directory\"",
											err: <*errors.errorString | 0xc00062c830>{
													s: "failed to t1: \"remove /var/lib/incus/containers/t1/exec-output/exec_ed6faba9-6647-4f38-83ad-b19bbce584a4.stdout: no such file or directory\"",
											},
									}
							to be nil
		*/
		err = nil
	}

	defer func() {
		log.FromContext(ctx).V(4).WithValues("status", result, "stdout", strings.TrimSpace(stdout.String()), "stderr", strings.TrimSpace(stderr.String()), "error", rerr).Info("Checked cloud-init status on instance")
	}()
	switch {
	case strings.Contains(stdout.String(), "status: error"):
		return CloudInitStatusError, nil
	case strings.Contains(stdout.String(), "status: done"):
		return CloudInitStatusDone, nil
	case strings.Contains(stdout.String(), "status: running"):
		return CloudInitStatusRunning, nil
	case err != nil:
		return CloudInitStatusUnknown, fmt.Errorf("cloud-init status command failed: %w", err)
	default:
		return CloudInitStatusUnknown, fmt.Errorf("could not understand cloud-init status command output (stdout=%q, stderr=%q)", stdout.String(), stderr.String())
	}
}
