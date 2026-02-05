# Talos cluster template

The talos cluster template will create a cluster using [Talos Control Plane](https://github.com/siderolabs/cluster-api-control-plane-provider-talos) provider.

A container running a haproxy server is used for the cluster load balancer endpoint. The load balancer endpoint will be the IP address of the haproxy container.

> **WARNING**: The load balancer container is a single point of failure for the control plane of the workload cluster, therefore should only be used for development or evaluation purposes. For production grade clusters, you should instead use a different option for the load balancer, e.g. [OVN](./ovn.md)

> **WARNING**: This cluster template is not currently tested in CI, therefore could be broken. Please raise a [GitHub issue](https://github.com/lxc/cluster-api-provider-incus/issues) if that is the case.

## Table Of Contents

<!-- toc -->

## Requirements

### 1. CAPN version

Talos cluster templates require CAPN version `v0.8.5` or newer.

### 2. Install Talos control plane and bootstrap providers on the cluster

```bash
clusterctl init -c talos -b talos
```

**Tested Versions**:
- [cluster-api-control-plane-provider-talos v0.5.12](https://github.com/siderolabs/cluster-api-control-plane-provider-talos/releases/tag/v0.5.12)
- [cluster-api-bootstrap-provider-talos v0.6.11](https://github.com/siderolabs/cluster-api-bootstrap-provider-talos/releases/tag/v0.6.11)
- [talos linux version v1.12.2](https://docs.siderolabs.com/talos/v1.12/overview/what-is-talos)

### 3. Import Talos nocloud server image into Incus

Start from [https://factory.talos.dev/](https://factory.talos.dev/)
- Select **Cloud Server**
- Select **Talos Linux Version** (v1.12.2)
- Select **nocloud** variant
- Select **Machine architecture** (amd64)
- Select **Secure Boot** (true or false)
- Select optional **System Extensions** (none required)
- Select **Bootloader** (auto)
- Select **BIOS Only** boot (**NOTE**: UEFI boot seems to not work correctly)

{{#tabs name:"image-configuration" tabs:"SecureBoot disabled,SecureBoot enabled" }}

{{#tab SecureBoot disabled }}

[Example configuration with SecureBoot disabled](https://factory.talos.dev/?arch=amd64&bootloader=auto&cmdline-set=true&extensions=-&platform=nocloud&target=cloud&version=1.12.2)

```bash
# 1. download nocloud-amd64.raw.xz, extract and convert to "rootfs.qcow2"
curl -L "https://factory.talos.dev/image/9ed5fecdacb36b5c5427b87d409f1065cfb2df69b0f71c58b868d9d466d8dab3/v1.12.2/nocloud-amd64.raw.xz" -o nocloud-amd64.raw.xz
unxz nocloud-amd64.raw.xz
qemu-img convert -f raw -O qcow2 nocloud-amd64.raw rootfs.qcow2
```

{{#/tab }}

{{#tab SecureBoot enabled }}

> **WARNING**: Avoid SecureBoot enabled images if unsure of how they work.

[Example configuration with SecureBoot enabled](https://factory.talos.dev/?arch=amd64&bootloader=auto&cmdline-set=true&extensions=-&platform=nocloud&secureboot=true&target=cloud&version=1.12.2)

```bash
# 1. download nocloud-amd64-secureboot.raw.xz, extract and convert to "rootfs.qcow2"
curl -L "https://factory.talos.dev/image/376567988ad370138ad8b2698212367b8edcb69b5fd68c80be1f2ec7d603b4ba/v1.12.2/nocloud-amd64-secureboot.raw.xz" -o nocloud-amd64-secureboot.raw.xz
unxz nocloud-amd64-secureboot.raw.xz
qemu-img convert -f raw -O qcow2 nocloud-amd64-secureboot.raw rootfs.qcow2
```

{{#/tab }}

{{#/tabs }}

```bash
# 2. generate image metadata tarball "metadata.tar.gz"
SERIAL="$(date '+%s')"
VERSION=v1.12.2
echo "
architecture: amd64
creation_date: $SERIAL
expiry_date: 0
properties:
  architecture: amd64
  description: talos $VERSION amd64 ($SERIAL)
  name: talos-$VERSION
  os: talos
  release: $VERSION
  serial: $SERIAL
  variant: default
templates: {}
" | tee metadata.yaml
tar cvzf metadata.tar.gz metadata.yaml
```

Finally, import the image into Incus:

```bash
# c. import image into Incus
incus image import metadata.tar.gz rootfs.qcow2 --alias talos-v1.12.2
```

Make a note of:
- The image alias `talos-v1.12.2`, as we will later refer to it when generating the cluster manifest.
- Whether you used the SecureBoot image variant (configuration below needs to match).

## Configuration

```bash
{{#include ../../../../../templates/cluster-template-talos.rc }}
```

## Generate cluster

```bash
clusterctl generate cluster example-cluster -i incus --flavor talos
```

## Cluster Template

```yaml
{{#include ../../../../../templates/cluster-template-talos.yaml }}
```
