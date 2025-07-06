# Machine Placement

CAPN works both for single node infrastructure (aimed at local development), as well as production clusters.

In a production cluster, it is usually desirable to ensure that cluster machines are scheduled on a specific hypervisor. For example, control plane machines may run on overprovisioned CPU hypervisors, whereas worker nodes can run on machines with GPUs.

In this page, we explain how to configure cluster groups in an existing cluster, and use them to launch CAPN machines on specific hypervisors.

## Table Of Contents

<!-- toc -->

## Cluster Members and Cluster Groups

Incus uses the concepts of [cluster members](https://linuxcontainers.org/incus/docs/main/explanation/clustering/#cluster-members) (individual hypervisors that are part of the cluster) and [cluster groups](https://linuxcontainers.org/incus/docs/main/explanation/clustering/#cluster-groups) (members grouped by the user based on specific criteria).

When launching an instance, the target may be set to:

- `<member>`, where `<member>` is the name of a cluster member.
- `@<group>`, where `<group>` is the name of a cluster group.

## Example cluster

Let's assume a cluster with 6 nodes. 3 CPU nodes `cpu-01`, `cpu-02`, `cpu-03` and 3 GPU nodes `gpu-01`, `gpu-02`, `gpu-03`.

We can see the list of cluster members with:

```bash
incus cluster list -c nuasm
```

Example output:

```bash
+--------+-----------------------+--------------+--------+-------------------+
|  NAME  |          URL          | ARCHITECTURE | STATUS |      MESSAGE      |
+--------+-----------------------+--------------+--------+-------------------+
| cpu-01 | https://10.0.1.1:8443 | x86_64       | ONLINE | Fully operational |
+--------+-----------------------+--------------+--------+-------------------+
| cpu-02 | https://10.0.1.2:8443 | x86_64       | ONLINE | Fully operational |
+--------+-----------------------+--------------+--------+-------------------+
| cpu-03 | https://10.0.1.3:8443 | x86_64       | ONLINE | Fully operational |
+--------+-----------------------+--------------+--------+-------------------+
| gpu-01 | https://10.0.2.1:8443 | x86_64       | ONLINE | Fully operational |
+--------+-----------------------+--------------+--------+-------------------+
| gpu-02 | https://10.0.2.2:8443 | x86_64       | ONLINE | Fully operational |
+--------+-----------------------+--------------+--------+-------------------+
| gpu-03 | https://10.0.2.3:8443 | x86_64       | ONLINE | Fully operational |
+--------+-----------------------+--------------+--------+-------------------+
```

By default, all cluster members are part of the `default` cluster group:

```bash
incus cluster group show default
```

Command output can be seen below:

```yaml
description: Default cluster group
members:
- cpu-01
- cpu-02
- cpu-03
- gpu-01
- gpu-02
- gpu-03
config: {}
name: default
```

## Configure cluster groups

We want to deploy Kubernetes clusters with control plane machines running on the `cpu-xx` hypervisors, and worker machines running on the `gpu-xx` hypervisors.

In order to do this, we can define two cluster groups, called `cpu-nodes` and `gpu-nodes` respectively:

```bash
incus cluster group create cpu-nodes
incus cluster group create gpu-nodes
```

Then, we assign each cluster member on the respective group:

```bash
incus cluster group assign cpu-01 cpu-nodes,default
incus cluster group assign cpu-02 cpu-nodes,default
incus cluster group assign cpu-03 cpu-nodes,default

incus cluster group assign gpu-01 gpu-nodes,default
incus cluster group assign gpu-02 gpu-nodes,default
incus cluster group assign gpu-03 gpu-nodes,default
```

You can check that the cluster group members have been configured properly:

```bash
incus cluster group show gpu-nodes
```

Example output:

```yaml
description: ""
members:
- gpu-01
- gpu-02
- gpu-03
config: {}
name: gpu
```

We have now configured our `cpu-nodes` and `gpu-nodes` cluster groups.

## Launch a cluster

Generate a cluster manifest using the [default cluster template](../reference/templates/default.md) and set the following additional configuration:

```bash
export CONTROL_PLANE_MACHINE_TARGET="@cpu-nodes"
export WORKER_MACHINE_TARGET="@gpu-nodes"
```

This will ensure control plane machines are scheduled on a cluster member that is part of the `cpu-nodes` group we configured earlier. Similarly, worker machines will be scheduled on an available member of the `gpu-nodes` group.
