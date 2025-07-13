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
- A `cloud-init-launch.service` is injected into the instance and enabled by default. This allows the cloud-init scripts to run once when the instance starts for the first time. See [cloud-init support for kind instances](#cloud-init-support-for-kind-instances)
- By default, `/usr/local/bin/entrypoint` will always attempt to overwrite `/etc/resolv.conf`. In Incus, this is not required, and also fails for unprivileged instances because `/etc/resolv.conf` is read-only. Therefore, we mutate the script to write to `/etc/local-resolv.conf` instead (and nullify this change).
- `kind` mounts `/lib/modules` into the containers. However, this fails under Incus. Instead, we mount `/boot` into `/usr/lib/ostree-boot`, so that kubeadm is able to retrieve the kernel configuration.

## cloud-init support for kind instances

ClusterAPI uses cloud-init configuration for cluster instances to bootstrap or join a cluster, but `kindest/node` do not have `cloud-init` preinstalled, meaning that nodes would not be configured without further action.

The ClusterAPI docker provider, which also uses the `kindest/node` images, addresses this by manually parsing the cloud-init configuration on the provider, then executing the cloud-init commands on the node. This was not deemed as a useful approach for CAPN.

Instead, CAPN will inject a `cloud-init-launch.service` on the nodes, and enable it by default (so it starts when the systemd entrypoint takes over). There are two ways that the cloud-init configuration can be applied:

1.  `/hack/cloud-init.py /hack/cloud-init.json` [**default**]

    In this (default) mode, CAPN will parse the YAML cloud-init config that the bootstrap provider has generated for the instance, and render it in JSON format as `/hack/cloud-init.json` inside the instance. A python script `/hack/cloud-init.py` is also injected, which applies these configs on the instance.

    Currently, only `write_files` and `runcmd` configuration is supported in this mode. This should cover kubeadm and most other bootstrap providers (e.g. RKE2, K3s, etc). However, it will not work if other cloud-init configuration is required (e.g. `users`). If you are affected by this, you are kindly requested to create an issue in [GitHub](https://github.com/lxc/cluster-api-provider-incus/issues) with more details, and support may be added.

    This mode is the default, as it requires no external dependencies for the cloud-init configuration to be applied.

2.  `apt update; apt install cloud-init; systemctl start cloud-final.service`.

    An alternative approach is to install the full cloud-init apt packages on the instance, and then start the `cloud-final` service, which executes cloud-init.

    This does a `apt update; apt install cloud-init` to install the `cloud-init` package from the archives, which requires external network connection and might not be desirable. However, it supports any cloud-init configuration that the bootstrap provider has set.

    To use this mode, the `LXCMachineTemplate` instance spec must be updated to set the `user.capn.x-kind-apt-install-cloud-init` config to true, for example:

    ```yaml,hidelines=#
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
    kind: LXCMachineTemplate
    metadata:
      name: "kind-machine-template"
    spec:
      template:
        spec:
          instanceType: kind
          config:
            user.capn.x-kind-apt-install-cloud-init: "true"
    ```
