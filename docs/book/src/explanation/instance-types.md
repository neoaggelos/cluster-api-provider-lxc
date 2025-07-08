# Instance Types

CAPN will launch one [instance](https://linuxcontainers.org/incus/docs/main/explanation/instances/) per LXCMachine. The LXCMachine `.spec.instanceType` field defines the type of instance that will be launched. The following instance types are supported:

## `container`

**Extra requirements**: none

When setting `.spec.instanceType: container`, an LXC system container is spawned for the LXCMachine. By default, CAPN will spawn privileged containers so that Kubernetes services run properly, but [unprivileged containers](./unprivileged-containers.md) are also supported.

CAPN offers pre-built kubeadm container images for select Kubernetes versions, see [Default Simplestreams Server](../reference/default-simplestreams-server.md) for details. Alternatively, [image-builder](../howto/images/kubeadm.md) can be used.

Custom images may be used, e.g. if using different control plane and bootstrap providers. It is also possible to use [stock Ubuntu images](../reference/templates/default.md#lxc_image_name-and-install_kubeadm) and install kubeadm manually, as part of the cloud-init.

If using custom images, you have to make sure that they support cloud-init.

## `virtual-machine`

**Extra requirements**: kvm support on the underlying hypervisor

When setting `.spec.instanceType: virtual-machine`, a QEMU-KVM based virtual machine is spawned for the LXCMachine. Virtual machines are not as lightweight as LXC containers, but they offer full isolation from the hypervisor where the instance is running, different kernel versions, etc.

CAPN offers pre-built kubeadm virtual-machine images for select Kubernetes versions, see [Default Simplestreams Server](../reference/default-simplestreams-server.md) for details. Alternatively, [image-builder](../howto/images/kubeadm.md) can be used.

Custom images may be used, e.g. if using different control plane and bootstrap providers. It is also possible to use [stock Ubuntu images](../reference/templates/default.md#lxc_image_name-and-install_kubeadm) and install kubeadm manually, as part of the cloud-init.

If using custom images, you have to make sure that they support cloud-init.

## `kind`

**Extra requirements**: `instance_oci` and `instance_oci_entrypoint` API extensions

When setting `.spec.instanceType: kind`, an OCI application container is spawned for the LXCMachine. By default, CAPN will spawn privileged containers so that Kubernetes services run properly, but [unprivileged containers](./unprivileged-containers.md) are also supported.

`kind` instances use the `kindest/node` images published to DockerHub by the [kind](https://kind.sigs.k8s.io) project, which are not maintained by CAPN, but should support all released Kubernetes versions. Note that `kindest/node` images do not come with cloud-init out of the box, therefore CAPN will manually execute the cloud-init script to ensure instances are configured.

Custom images may be used, but they are expected to be "binary-compatible" with the `kindest/node` image. If you encounter any issues, you are kindly requested to create an issue in [GitHub](https://github.com/lxc/cluster-api-provider-incus/issues).

**NOTE**: OCI containers are currently supported only for Incus 6.11 or newer ([`instance_oci`](https://linuxcontainers.org/incus/docs/main/api-extensions/#instance-oci) was added in version [6.5](https://github.com/lxc/incus/releases/v6.5.0), and [`instance_oci_entrypoint`](https://linuxcontainers.org/incus/docs/main/api-extensions/#instance-oci-entrypoint) in version [6.11](https://github.com/lxc/incus/releases/v6.11.0)). Canonical LXD does not currently support OCI containers.

Running the kindest/node containers under Incus requires a few in-place modifications on the instance configuration files:
- A symlink is created at `/init`, which points to `/usr/local/bin/entrypoint`. This is because kind must run as PID 1 for systemd to run properly, but Incus will override PID 1 of containers unless the entrypoint is one of `/init`, `/sbin/init`, or `/s6-init`. Therefore, the resulting `oci.entrypoint` becomes `/init /sbin/init`
- A `cloud-init-launch.service` is injected into the instance and enabled by default. This allows the cloud-init scripts to run once when the instance starts for the first time.
  - The `cloud-init-launch.service` is able to `apt install cloud-init` (not pre-installed in the image) and start the cloud-final target. However, this makes instances depend on external packages, and is disabled
  - Instead, CAPN will parse a limited subset of the cloud-init manifest from the bootstrap provider, render it to `/hack/cloud-init.json`, and apply it on the host using `/hack/cloud-init.py`
- By default, `/usr/local/bin/entrypoint` will always attempt to overwrite `/etc/resolv.conf`. In Incus, this is not required, and also fails for unprivileged instances because `/etc/resolv.conf` is read-only. Therefore, we mutate the script to write to `/etc/local-resolv.conf` instead (and nullify this change).
- `kind` mounts `/lib/modules` into the containers. However, this fails under Incus. Instead, we mount `/boot` into `/usr/lib/ostree-boot`, so that kubeadm is able to retrieve the kernel configuration.
