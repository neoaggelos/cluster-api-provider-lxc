package shared

import (
	"context"
	"fmt"

	"github.com/lxc/cluster-api-provider-incus/internal/lxc"
	"github.com/lxc/incus/v6/shared/api"

	. "github.com/onsi/gomega"
)

// UnprivilegedContainersClusterVariables returns cluster template variables to launch unprivileged containers.
//
// Also attempts (best-effort) to account for well-known Kubernetes limitations.
//
// Current well-known issues:
//
// ## VERSION: v1.36.0
// ## ISSUE: kubelet does not start when kubelet folder is on ZFS filesystem because of missing cadvisor plugin in 1.36.0
// ## GITHUB LINK: https://github.com/kubernetes/kubernetes/issues/138556
// ## MITIGATION: Do not use a ZFS filesystem.
// ## EXAMPLE ERROR: Apr 23 21:45:14 strike01 kubelet[326484]: I0423 21:45:14.680361  326484 fs.go:406] no plugin found for filesystem type: zfs
//
// ## VERSION: v1.36.1+
// ## ISSUE: kubelet does not work in privileged containers when using a ZFS filesystem
// ## GITHUB LINK: <TODO>
// ## MITIGATION: Do not use a ZFS filesystem for unprivileged containers.
// ## EXAMPLE ERROR: 09:10:26 capn-default-unprivileged-2fnr-47zq6-cbrt2 kubelet[697]: E0518 09:10:26.166372     697 container_manager_linux.go:987] "Unable to get rootfs data from cAdvisor interface" err="cannot find filesystem info for device \"zfs/containers/capn-default-unprivileged-2fnr-47zq6-cbrt2\""
func (e2eCtx *E2EContext) UnprivilegedContainersClusterVariables() map[string]string {
	lxcClient, err := lxc.New(context.TODO(), e2eCtx.Settings.LXCClientOptions)
	Expect(err).ToNot(HaveOccurred(), "Failed to initialize client")

	d := map[string]string{
		"PRIVILEGED": "false",
	}

	// if possible, use a storage pool of type "dir", to work around Kubernetes v1.36+ issues with zfs on unprivileged containers
	pools, err := lxcClient.GetStoragePools()
	Expect(err).ToNot(HaveOccurred(), "Failed to list storage pools")
	for _, pool := range pools {
		if pool.Status == api.StoragePoolStatusCreated && pool.Driver == "dir" {
			Logf("Use storage pool %q (dir) for unprivileged instances", pool)
			d["CONTROL_PLANE_MACHINE_DEVICES"] = fmt.Sprintf("['root,type=disk,path=/,pool=%s']", pool)
			d["WORKER_MACHINE_DEVICES"] = fmt.Sprintf("['root,type=disk,path=/,pool=%s']", pool)
			break
		}
	}

	return d
}
