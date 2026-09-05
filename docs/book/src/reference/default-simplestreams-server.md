# Default Simplestreams Server

The `cluster-api-provider-incus` project offers pre-built kubeadm images for specific Kubernetes versions.

The default simplestreams server is available under [https://images.linuxcontainers.org/capn/](https://images.linuxcontainers.org/capn/streams/v1/index.json).

> **WARNING: In 2025-12-31, the default simplestreams server has migrated to a new location. If using version `v0.8.2` or older, you are kindly requested to update to `v0.8.3` or newer.**
>
> **The new URL is at [https://images.linuxcontainers.org/capn/](https://images.linuxcontainers.org/capn/streams/v1/index.json).**
>
> **For more details, refer to [https://github.com/lxc/cluster-api-provider-incus/issues/180](https://github.com/lxc/cluster-api-provider-incus/issues/180)**

## Table Of Contents

<!-- toc -->

## Support-level disclaimer

- The simplestreams server may terminate at any time, and should only be used for evaluation purposes.
- The images are provided "as-is", based on the upstream Ubuntu 24.04 cloud images, and do not include latest security updates.
- Container and virtual-machine amd64 images are provided, compatible and tested with both [Incus] and [Canonical LXD].
- Container arm64 images are provided, compatible and tested with both [Incus] and [Canonical LXD]. Virtual machine images for arm64 are currently not available, due to lack of CI infrastructure to build and test the images.
- Availability and support of Kubernetes versions is primarily driven by CI testing requirements. New Kubernetes versions are added on a best-effort basis, mainly as needed for development and CI testing.
- Images for Kubernetes versions might be removed from the simplestreams server after the Kubernetes version reaches [End of Life](https://kubernetes.io/releases/patch-releases/#support-period).

It is recommended that production environments [build their own custom images](../howto/images/index.md) instead.

## Provided images

Provided images are built in [GitHub Actions](https://github.com/lxc/cluster-api-provider-incus/actions/workflows/build-kubeadm-images.yml).

The following images are currently provided:

| Image Alias     | Base Image   | Description                            | amd64 | arm64 |
| --------------- | ------------ | -------------------------------------- | ----- | ----- |
| haproxy         | Ubuntu 24.04 | Haproxy image for development clusters | X     | X     |
| kubeadm/v1.33.0 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.33.0   | X     | X     |
| kubeadm/v1.33.5 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.33.5   | X     | X     |
| kubeadm/v1.34.0 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.34.0   | X     | X     |
| kubeadm/v1.35.0 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.35.0   | X     | X     |
| kubeadm/v1.35.5 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.35.5   | X     | X     |
| kubeadm/v1.36.1 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.36.1   | X     | X     |
| kubeadm/v1.36.4 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.36.4   | X     | X     |
| kubeadm/v1.37.0 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.37.0   | X     | X     |

Note that the table above might be out of date. See [streams/v1/index.json] and [streams/v1/images.json] for the list of versions currently available.

## Check available images supported by your infrastructure

{{#tabs name:"images" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

Configure the `capi` remote:

```bash
incus remote add capi https://images.linuxcontainers.org/capn/ --protocol=simplestreams
```

List available images (with filters):

```bash
incus image list capi:                                  # list all images
incus image list capi: type=virtual-machine             # list kvm images
incus image list capi: release=v1.35.0                  # list v1.35.0 images
incus image list capi: arch=amd64                       # list amd64 images
```

Example output:

```bash
# incus image list capi: release=v1.35.0
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
|             ALIAS              | FINGERPRINT  | PUBLIC |             DESCRIPTION              | ARCHITECTURE |      TYPE       |    SIZE    |     UPLOAD DATE      |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
| kubeadm/v1.35.0 (3 more)       | a2a90287e692 | yes    | kubeadm v1.35.0 amd64 (202512182137) | x86_64       | CONTAINER       | 586.96MiB  | 2025/12/18 02:00 EET |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
| kubeadm/v1.35.0 (3 more)       | dfcc37a45736 | yes    | kubeadm v1.35.0 amd64 (202512182138) | x86_64       | VIRTUAL-MACHINE | 1107.94MiB | 2025/12/18 02:00 EET |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
| kubeadm/v1.35.0/arm64 (1 more) | e19b3bedafc1 | yes    | kubeadm v1.35.0 arm64 (202512182140) | aarch64      | CONTAINER       | 540.61MiB  | 2025/12/18 02:00 EET |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
```

{{#/tab }}

{{#tab Canonical LXD }}

Configure the `capi` remote:

```bash
lxc remote add capi https://images.linuxcontainers.org/capn/ --protocol=simplestreams
```

List available images (with filters):

```bash
lxc image list capi:                                  # list all images
lxc image list capi: type=virtual-machine             # list kvm images
lxc image list capi: release=v1.35.0                  # list v1.35.0 images
lxc image list capi: arch=amd64                       # list amd64 images
```

Example output:

```bash
# lxc image list capi: release=v1.35.0
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
|             ALIAS              | FINGERPRINT  | PUBLIC |             DESCRIPTION              | ARCHITECTURE |      TYPE       |    SIZE    |          UPLOAD DATE          |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.35.0 (3 more)       | a2a90287e692 | yes    | kubeadm v1.35.0 amd64 (202512182137) | x86_64       | CONTAINER       | 586.96MiB  | Dec 18, 2025 at 12:00am (UTC) |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.35.0 (3 more)       | dfcc37a45736 | yes    | kubeadm v1.35.0 amd64 (202512182138) | x86_64       | VIRTUAL-MACHINE | 1107.94MiB | Dec 18, 2025 at 12:00am (UTC) |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.35.0/arm64 (1 more) | e19b3bedafc1 | yes    | kubeadm v1.35.0 arm64 (202512182140) | aarch64      | CONTAINER       | 540.61MiB  | Dec 18, 2025 at 12:00am (UTC) |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
```

{{#/tab }}
{{#/tabs }}

<!-- links -->
[National Technical University Of Athens]: https://ntua.gr/en
[Incus]: https://linuxcontainers.org/incus/docs/main/
[Canonical LXD]: https://canonical-lxd.readthedocs-hosted.com/en/
[streams/v1/index.json]: https://images.linuxcontainers.org/capn//streams/v1/index.json
[streams/v1/images.json]: https://images.linuxcontainers.org/capn//streams/v1/images.json
