# Default Simplestreams Server

The `cluster-api-provider-lxc` project runs a simplestreams server with pre-built kubeadm images for specific Kubernetes versions.

The default simplestreams server is available through an Amazon CloudFront distribution at [https://d14dnvi2l3tc5t.cloudfront.net](https://d14dnvi2l3tc5t.cloudfront.net).

Running infrastructure costs are kindly subsidized by the [National Technical University Of Athens].

## Support-level disclaimer

- The simplestreams server may terminate at any time, and should only be used for evaluation purposes.
- The images are provided as-is.
- The images are based on the upstream Ubuntu 24.04 cloud images.
- The images might not include latest security updates.
- Container and virtual-machine images are provided, compatible and tested with both [Incus] and [Canonical LXD].
- The images only support the amd64 architecture. There are no current plans to support more architectures (e.g. arm64), but that is subject to change in the future.
- Availability and support of Kubernetes versions is primarily driven by CI testing requirements.
- New Kubernetes versions are added on a best-effort basis, mainly as needed for development and CI testing.
- Images for Kubernetes versions might be removed from the simplestreams after the Kubernetes version reaches [End of Life](https://kubernetes.io/releases/patch-releases/#support-period).

It is recommended that production environments [build their own custom images](./../howto/build-base-images.md) instead.

## Provided images

The following images are currently provided:

| Image Alias | Base Image | Description |
|-|-|-|
| haproxy | Ubuntu 24.04 | Haproxy image for development clusters |
| kubeadm/v1.32.0 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.32.0 |

Note that the table above might be out of date. See [streams/v1/index.json] and [streams/v1/images.json] for the list of versions currently available.

## Check available images supported by your infrastructure

{{#tabs name:"images" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
$ incus remote add capi https://d14dnvi2l3tc5t.cloudfront.net --protocol=simplestreams
$ incus image list capi:
```

Example output for server that offers the `haproxy` container image, as well as kubeadm images only for `v1.32.0`:

```bash
+--------------------------+--------------+--------+-----------------------------------+--------------+-----------------+-----------+----------------------+
|          ALIAS           | FINGERPRINT  | PUBLIC |            DESCRIPTION            | ARCHITECTURE |      TYPE       |   SIZE    |     UPLOAD DATE      |
+--------------------------+--------------+--------+-----------------------------------+--------------+-----------------+-----------+----------------------+
| haproxy (7 more)         | 328da3b7ba2d | yes    | ubuntu noble amd64 (202501120613) | x86_64       | CONTAINER       | 133.87MiB | 2025/01/12 02:00 EET |
+--------------------------+--------------+--------+-----------------------------------+--------------+-----------------+-----------+----------------------+
| kubeadm/v1.32.0 (7 more) | 5053ee6cac52 | yes    | ubuntu noble amd64 (202501120531) | x86_64       | CONTAINER       | 668.13MiB | 2025/01/12 02:00 EET |
+--------------------------+--------------+--------+-----------------------------------+--------------+-----------------+-----------+----------------------+
| kubeadm/v1.32.0 (7 more) | 74062c0bcb2e | yes    | ubuntu noble amd64 (202501121335) | x86_64       | VIRTUAL-MACHINE | 912.87MiB | 2025/01/12 02:00 EET |
+--------------------------+--------------+--------+-----------------------------------+--------------+-----------------+-----------+----------------------+
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
$ lxc remote add capi https://d14dnvi2l3tc5t.cloudfront.net --protocol=simplestreams
$ lxc image list capi:
```
Example output for server that offers the `haproxy` container image, as well as kubeadm images only for `v1.32.0`:

```bash
+--------------------------+--------------+--------+-----------------------------------+--------------+-----------------+------------+-------------------------------+
|          ALIAS           | FINGERPRINT  | PUBLIC |            DESCRIPTION            | ARCHITECTURE |      TYPE       |    SIZE    |          UPLOAD DATE          |
+--------------------------+--------------+--------+-----------------------------------+--------------+-----------------+------------+-------------------------------+
| haproxy (7 more)         | 328da3b7ba2d | yes    | ubuntu noble amd64 (202501120613) | x86_64       | CONTAINER       | 133.87MiB  | Jan 12, 2025 at 12:00am (UTC) |
+--------------------------+--------------+--------+-----------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.32.0 (7 more) | 5053ee6cac52 | yes    | ubuntu noble amd64 (202501120531) | x86_64       | CONTAINER       | 668.13MiB  | Jan 12, 2025 at 12:00am (UTC) |
+--------------------------+--------------+--------+-----------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.32.0 (7 more) | cc772d5b71dc | yes    | ubuntu noble amd64 (202501081536) | x86_64       | VIRTUAL-MACHINE | 1118.44MiB | Jan 8, 2025 at 12:00am (UTC)  |
+--------------------------+--------------+--------+-----------------------------------+--------------+-----------------+------------+-------------------------------+
```

{{#/tab }}
{{#/tabs }}

<!-- links -->
[National Technical University Of Athens]: https://ntua.gr/en
[Incus]: https://linuxcontainers.org/incus/docs/main/
[Canonical LXD]: https://canonical-lxd.readthedocs-hosted.com/en/
[streams/v1/index.json]: https://d14dnvi2l3tc5t.cloudfront.net/streams/v1/index.json
[streams/v1/images.json]: https://d14dnvi2l3tc5t.cloudfront.net/streams/v1/images.json
