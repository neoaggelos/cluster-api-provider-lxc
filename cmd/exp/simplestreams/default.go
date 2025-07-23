package main

import "github.com/lxc/cluster-api-provider-incus/internal/lxc"

var (
	containerFTypeByServer = map[string]string{
		lxc.Incus: "incus_combined.tar.gz",
		lxc.LXD:   "lxd_combined.tar.gz",
	}
	vmMetadataFTypeByServer = map[string]string{
		lxc.Incus: "incus.tar.xz",
		lxc.LXD:   "lxd.tar.xz",
	}
	vmRootfsFTypeByServer = map[string]string{
		lxc.Incus: "disk-kvm.img",
		lxc.LXD:   "disk1.img",
	}
)
