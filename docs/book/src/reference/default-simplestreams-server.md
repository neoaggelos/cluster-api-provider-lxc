# Default Simplestreams Server

The `cluster-api-provider-lxc` project runs a simplestreams server with pre-built kubeadm images for specific Kubernetes versions.

The default simplestreams server is available through an Amazon CloudFront distribution at [https://d14dnvi2l3tc5t.cloudfront.net](https://d14dnvi2l3tc5t.cloudfront.net).

Running infrastructure costs are kindly subsidized by the [National Technical University Of Athens].

## Table Of Contents

<!-- toc -->

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

It is recommended that production environments [build their own custom images](../howto/images/index.md) instead.

## Provided images

The following images are currently provided:

| Image Alias | Base Image | Description |
|-|-|-|
| haproxy | Ubuntu 24.04 | Haproxy image for development clusters |
| kubeadm/v1.31.5 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.31.5 |
| kubeadm/v1.32.0 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.32.0 |
| kubeadm/v1.32.1 | Ubuntu 24.04 | Kubeadm image for Kubernetes v1.32.1 |

Note that the table above might be out of date. See [streams/v1/index.json] and [streams/v1/images.json] for the list of versions currently available.

## Check available images supported by your infrastructure

{{#tabs name:"images" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
incus remote add capi https://d14dnvi2l3tc5t.cloudfront.net --protocol=simplestreams
incus image list capi:
```

Example output for server that offers the `haproxy` container image, as well as kubeadm images for `v1.31.5`, `v1.32.0` and `v1.32.1`:

```bash
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
|          ALIAS           | FINGERPRINT  | PUBLIC |             DESCRIPTION              | ARCHITECTURE |      TYPE       |    SIZE    |     UPLOAD DATE      |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
| haproxy (3 more)         | 9260bf4ebaee | yes    | haproxy noble amd64 (202502041428)   | x86_64       | CONTAINER       | 122.57MiB  | 2025/02/04 02:00 EET |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
| kubeadm/v1.31.5 (3 more) | 77b06d558f89 | yes    | kubeadm v1.31.5 amd64 (202502051122) | x86_64       | CONTAINER       | 719.11MiB  | 2025/02/05 02:00 EET |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
| kubeadm/v1.31.5 (3 more) | e193205417e2 | yes    | kubeadm v1.31.5 amd64 (202502051127) | x86_64       | VIRTUAL-MACHINE | 1040.99MiB | 2025/02/05 02:00 EET |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
| kubeadm/v1.32.0 (3 more) | 8674fb99246b | yes    | kubeadm v1.32.0 amd64 (202502041405) | x86_64       | CONTAINER       | 725.99MiB  | 2025/02/04 02:00 EET |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
| kubeadm/v1.32.0 (3 more) | b3bb90fdf849 | yes    | kubeadm v1.32.0 amd64 (202502041414) | x86_64       | VIRTUAL-MACHINE | 1023.19MiB | 2025/02/04 02:00 EET |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
| kubeadm/v1.32.1 (3 more) | 22a41fe2beb1 | yes    | kubeadm v1.32.1 amd64 (202502042122) | x86_64       | CONTAINER       | 725.97MiB  | 2025/02/04 02:00 EET |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
| kubeadm/v1.32.1 (3 more) | b2fca2f871ca | yes    | kubeadm v1.32.1 amd64 (202502042127) | x86_64       | VIRTUAL-MACHINE | 1026.84MiB | 2025/02/04 02:00 EET |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+----------------------+
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
lxc remote add capi https://d14dnvi2l3tc5t.cloudfront.net --protocol=simplestreams
lxc image list capi:
```

Example output for server that offers the `haproxy` container image, as well as kubeadm images for `v1.31.5`, `v1.32.0` and `v1.32.1`:

```bash
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
|          ALIAS           | FINGERPRINT  | PUBLIC |             DESCRIPTION              | ARCHITECTURE |      TYPE       |    SIZE    |          UPLOAD DATE          |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| haproxy (3 more)         | 9260bf4ebaee | yes    | haproxy noble amd64 (202502041428)   | x86_64       | CONTAINER       | 122.57MiB  | Feb 4, 2025 at 12:00am (UTC)  |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.31.5 (3 more) | 77b06d558f89 | yes    | kubeadm v1.31.5 amd64 (202502051122) | x86_64       | CONTAINER       | 719.11MiB  | Feb 5, 2025 at 12:00am (UTC)  |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.31.5 (3 more) | 2017519d44d5 | yes    | kubeadm v1.31.5 amd64 (202501150903) | x86_64       | VIRTUAL-MACHINE | 1154.42MiB | Jan 15, 2025 at 12:00am (UTC) |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.32.0 (3 more) | 8674fb99246b | yes    | kubeadm v1.32.0 amd64 (202502041405) | x86_64       | CONTAINER       | 725.99MiB  | Feb 4, 2025 at 12:00am (UTC)  |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.32.0 (3 more) | a7db592d1905 | yes    | kubeadm v1.32.0 amd64 (202501150903) | x86_64       | VIRTUAL-MACHINE | 1162.60MiB | Jan 15, 2025 at 12:00am (UTC) |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.32.1 (3 more) | 22a41fe2beb1 | yes    | kubeadm v1.32.1 amd64 (202502042122) | x86_64       | CONTAINER       | 725.97MiB  | Feb 4, 2025 at 12:00am (UTC)  |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
| kubeadm/v1.32.1 (3 more) | f68a0391d16d | yes    | kubeadm v1.32.1 amd64 (202501150903) | x86_64       | VIRTUAL-MACHINE | 1158.56MiB | Jan 15, 2025 at 12:00am (UTC) |
+--------------------------+--------------+--------+--------------------------------------+--------------+-----------------+------------+-------------------------------+
```

{{#/tab }}
{{#/tabs }}

<!-- links -->
[National Technical University Of Athens]: https://ntua.gr/en
[Incus]: https://linuxcontainers.org/incus/docs/main/
[Canonical LXD]: https://canonical-lxd.readthedocs-hosted.com/en/
[streams/v1/index.json]: https://d14dnvi2l3tc5t.cloudfront.net/streams/v1/index.json
[streams/v1/images.json]: https://d14dnvi2l3tc5t.cloudfront.net/streams/v1/images.json
