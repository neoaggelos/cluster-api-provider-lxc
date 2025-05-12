# Default Simplestreams Server

The `cluster-api-provider-incus` project runs a simplestreams server with pre-built kubeadm images for specific Kubernetes versions.

The default simplestreams server is available through an Amazon CloudFront distribution at [https://d14dnvi2l3tc5t.cloudfront.net](https://d14dnvi2l3tc5t.cloudfront.net).

Running infrastructure costs are kindly subsidized by the [National Technical University Of Athens].

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

| Image Alias | Base Image | Description | amd64 | arm64 |
|-|-|-|-|-|
| haproxy | Ubuntu 24.04 | Haproxy image for development clusters | X | X |
| kubeadm/v1.31.5 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.31.5 | X | |
| kubeadm/v1.32.0 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.32.0 | X | |
| kubeadm/v1.32.1 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.32.1 | X | |
| kubeadm/v1.32.2 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.32.2 | X | |
| kubeadm/v1.32.3 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.32.3 | X | |
| kubeadm/v1.32.4 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.32.4 | X | X |
| kubeadm/v1.33.0 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.33.0 | X | X |

Note that the table above might be out of date. See [streams/v1/index.json] and [streams/v1/images.json] for the list of versions currently available.

## Check available images supported by your infrastructure

{{#tabs name:"images" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

Configure the `capi` remote:

```bash
incus remote add capi https://d14dnvi2l3tc5t.cloudfront.net --protocol=simplestreams
```

List available images (with filters):

```bash
incus image list capi:                                  # list all images
incus image list capi: type=virtual-machine             # list kvm images
incus image list capi: release=v1.33.0                  # list v1.33.0 images
incus image list capi: arch=amd64                       # list amd64 images
```

Example output:

```bash
# incus image list capi: release=v1.33.0
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-----------------------+
|             ALIAS              | FINGERPRINT  | PUBLIC |             DESCRIPTION              | ARCHITECTURE |      TYPE       |    SIZE    |      UPLOAD DATE      |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-----------------------+
| kubeadm/v1.33.0 (3 more)       | 162f0a10f60b | yes    | kubeadm v1.33.0 amd64 (202505101746) | x86_64       | VIRTUAL-MACHINE | 1076.20MiB | 2025/05/10 03:00 EEST |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-----------------------+
| kubeadm/v1.33.0 (3 more)       | fb7876d1be57 | yes    | kubeadm v1.33.0 amd64 (202505101747) | x86_64       | CONTAINER       | 683.60MiB  | 2025/05/10 03:00 EEST |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-----------------------+
| kubeadm/v1.33.0/arm64 (1 more) | 5141a5f24713 | yes    | kubeadm v1.33.0 arm64 (202505101748) | aarch64      | CONTAINER       | 664.60MiB  | 2025/05/10 03:00 EEST |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-----------------------+
```

{{#/tab }}

{{#tab Canonical LXD }}

Configure the `capi` remote:

```bash
lxc remote add capi https://d14dnvi2l3tc5t.cloudfront.net --protocol=simplestreams
```

List available images (with filters):

```bash
lxc image list capi:                                  # list all images
lxc image list capi: type=virtual-machine             # list kvm images
lxc image list capi: release=v1.33.0                  # list v1.33.0 images
lxc image list capi: arch=amd64                       # list amd64 images
```

Example output:

```bash
# lxc image list capi: release=v1.33.0
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
|             ALIAS              | FINGERPRINT  | PUBLIC |             DESCRIPTION              | ARCHITECTURE |      TYPE       |    SIZE    |          UPLOAD DATE          |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.33.0 (3 more)       | 16240bc72afb | yes    | kubeadm v1.33.0 amd64 (202505031313) | x86_64       | VIRTUAL-MACHINE | 1064.22MiB | May 3, 2025 at 12:00am (UTC)  |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.33.0 (3 more)       | fb7876d1be57 | yes    | kubeadm v1.33.0 amd64 (202505101747) | x86_64       | CONTAINER       | 683.60MiB  | May 10, 2025 at 12:00am (UTC) |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.33.0/arm64 (1 more) | 5141a5f24713 | yes    | kubeadm v1.33.0 arm64 (202505101748) | aarch64      | CONTAINER       | 664.60MiB  | May 10, 2025 at 12:00am (UTC) |
+--------------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
```

{{#/tab }}
{{#/tabs }}

<!-- links -->
[National Technical University Of Athens]: https://ntua.gr/en
[Incus]: https://linuxcontainers.org/incus/docs/main/
[Canonical LXD]: https://canonical-lxd.readthedocs-hosted.com/en/
[streams/v1/index.json]: https://d14dnvi2l3tc5t.cloudfront.net/streams/v1/index.json
[streams/v1/images.json]: https://d14dnvi2l3tc5t.cloudfront.net/streams/v1/images.json
