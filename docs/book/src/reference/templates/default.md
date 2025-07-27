# Default cluster template

The default cluster-template uses the [`capn-default` cluster class](#cluster-class).

All load balancer types are supported through configuration options. Further, it allows deploying the default kube-flannel CNI on the cluster.

## Table Of Contents

<!-- toc -->

## Requirements

1. ClusterAPI `ClusterTopology` Feature Gate is enabled (initialize providers with `CLUSTER_TOPOLOGY=true`).
2. The management cluster can reach the load balancer endpoint, so that it can connect to the workload cluster.

## Configuration

```bash
{{#include ../../../../../templates/cluster-template.rc }}
```

## Generate cluster

```bash
clusterctl generate cluster example-cluster -i incus
```

## Configuration notes

### `LXC_SECRET_NAME`

Name of Kubernetes secret with [infrastructure credentials](../identity-secret.md#identity-secret-format).

### `LOAD_BALANCER`

You must choose between one of the options above to configure the load balancer for the infrastructure. See [Cluster Load Balancer Types](../../explanation/load-balancer.md) for more details.

{{#tabs name:"load-balancer-type" tabs:"LXC,OCI,Kube VIP,OVN" }}

{{#tab LXC }}

Use an LXC container for the load balancer. The instance size will be 1 core, 1 GB RAM and will have the `default` profile attached.

```bash
export LOAD_BALANCER="lxc: {profiles: [default], flavor: c1-m1}"
```

{{#/tab }}

{{#tab OCI }}

Use an OCI container for the load balancer. The instance size will be 1 core, 1 GB RAM and will have the `default` profile attached.

```bash
export LOAD_BALANCER="oci: {profiles: [default], flavor: c1-m1}"
```

{{#/tab }}

{{#tab Kube VIP }}

Deploy `kube-vip` with static pods on the control plane nodes. The VIP address will be `10.0.42.1`.

```bash
export LOAD_BALANCER="kube-vip: {host: 10.0.42.1}"
```

{{#/tab }}

{{#tab OVN }}

Create an OVN network load balancer with IP `10.100.42.1` on the OVN network `ovn-0`.

```bash
export LOAD_BALANCER="ovn: {host: 10.100.42.1, networkName: ovn-0}"
```

{{#/tab }}

{{#/tabs }}

### `PRIVILEGED`

Set `PRIVILEGED=false` to use [unprivileged containers](../../explanation/unprivileged-containers.md).

### `DEPLOY_KUBE_FLANNEL`

Set `DEPLOY_KUBE_FLANNEL=true` to deploy the default kube-flannel CNI on the cluster. If not set, you must manually deploy a CNI before the cluster is usable.

### `LXC_IMAGE_NAME` and `INSTALL_KUBEADM`

`LXC_IMAGE_NAME` must be set if creating a cluster with a Kubernetes version for which no [pre-built Kubeadm images](../default-simplestreams-server.md#provided-images) are available. It is recommended to build [custom images](../../howto/images/kubeadm.md) in this case.

Note that Incus and Canonical LXD use incompatible image servers. To help mitigate this issue, the following image names are recognized:

{{#tabs name:"lxc-image-name" tabs:"Incus,Canonical LXD" }}
{{#tab Incus }}

- `ubuntu:VERSION` => `ubuntu/VERSION/cloud` from [https://images.linuxcontainers.org](https://images.linuxcontainers.org)
- `debian:VERSION` => `debian/VERSION/cloud` from [https://images.linuxcontainers.org](https://images.linuxcontainers.org)
- `images:IMAGE` => `IMAGE` from [https://images.linuxcontainers.org](https://images.linuxcontainers.org)
- `capi:IMAGE` => `IMAGE` from [https://d14dnvi2l3tc5t.cloudfront.net](https://d14dnvi2l3tc5t.cloudfront.net) ([default simplestreams server](../default-simplestreams-server.md))
- `capi-stg:IMAGE` => `IMAGE` from [https://djapqxqu5n2qu.cloudfront.net](https://djapqxqu5n2qu.cloudfront.net) (staging simplestreams server)

{{#/tab }}
{{#tab Canonical LXD }}

- `ubuntu:VERSION` => `VERSION` from [https://cloud-images.ubuntu.com/releases](https://cloud-images.ubuntu.com/releases)
- `debian:VERSION` => `debian/VERSION/cloud` from [https://images.lxd.canonical.com](https://images.lxd.canonical.com)
- `images:IMAGE` => `IMAGE` from [https://images.lxd.canonical.com](https://images.lxd.canonical.com)
- `capi:IMAGE` => `IMAGE` from [https://d14dnvi2l3tc5t.cloudfront.net](https://d14dnvi2l3tc5t.cloudfront.net) ([default simplestreams server](../default-simplestreams-server.md))
- `capi-stg:IMAGE` => `IMAGE` from [https://djapqxqu5n2qu.cloudfront.net](https://djapqxqu5n2qu.cloudfront.net) (staging simplestreams server)

{{#/tab }}
{{#/tabs }}

`INSTALL_KUBEADM=true` should be set if using an image without kubeadm installed. In this case, the cluster template will inject `preKubeadmCommands` to install Kubeadm using the [install-kubeadm.sh](../../static/v0.1/install-kubeadm.sh) script. Note that this is only meant for development purposes.

Finally, if the image name contains the text `VERSION`, it will be replaced with the Kubernetes version of the cluster (sourced from the Machine object).

### `CONTROL_PLANE_MACHINE_TYPE` and `WORKER_MACHINE_TYPE`

These must be set to `container`, `virtual-machine` or `kind`.

It is a common scenario to use `container` instances for the control plane nodes, and then `virtual-machine` instances for the worker nodes of a cluster.

See [Instance Types](../../explanation/instance-types.md) for more details.

### `CONTROL_PLANE_MACHINE_PROFILES` and `WORKER_MACHINE_PROFILES`

A list of [profile](https://linuxcontainers.org/incus/docs/main/profiles/) names to attach to the created instances. The [default kubeadm profile](../profile/kubeadm.md) will be automatically added to the list, if not already present. For local development, this should be `[default]`.

### `CONTROL_PLANE_MACHINE_DEVICES` and `WORKER_MACHINE_DEVICES`

A list of [device](https://linuxcontainers.org/incus/docs/main/reference/devices/) configuration overrides for the created instances. This can be used to override the network interface or the root disk of the instance.

Devices are specified as an array of strings with the following syntax: `<device>,<key>=<value>`. For example, to override the network of the created instances, you can specify:

```bash
export CONTROL_PLANE_MACHINE_DEVICES="['eth0,type=nic,network=my-network']"
export WORKER_MACHINE_DEVICES="['eth0,type=nic,network=my-network']"
```

Similarly, to override the network and also specify a custom root disk size, you can use:

```bash
export CONTROL_PLANE_MACHINE_DEVICES="['eth0,type=nic,network=my-network', 'root,type=disk,path=/,pool=local,size=50GB']"
export WORKER_MACHINE_DEVICES="['eth0,type=nic,network=my-network', 'root,type=disk,path=/,pool=local,size=50GB']"
```

### `CONTROL_PLANE_MACHINE_FLAVOR` and `WORKER_MACHINE_FLAVOR`

Instance size for the control plane and worker instances. This is typically specified as `cX-mY`, in which case the instance size will be `X cores` and `Y GB RAM`.

### `CONTROL_PLANE_MACHINE_TARGET` and `WORKER_MACHINE_TARGET`

When infrastructure is a cluster, specify target cluster member or cluster group for control plane and worker machines. See [Machine Placement](../../howto/machine-placement.md) for more details.

## Cluster Template

```yaml
{{#include ../../../../../templates/cluster-template.yaml }}
```

## Cluster Class Definition

```yaml
{{#include ../../../../../templates/clusterclass-capn-default.yaml }}
```
