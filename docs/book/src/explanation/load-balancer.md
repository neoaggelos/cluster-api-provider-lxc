# Cluster Load Balancer types

Part of the responsibilities of the infrastructure provider is to provision a Load Balancer for the control plane endpoint of workload clusters.

`cluster-api-provider-incus` supports a number of different options for how to provision the Load Balancer. They are mostly a tradeoff between simplicity, infrastructure requirements and production readiness.

## Load balancer types

In the LXCCluster resource, `spec.loadBalancer.type` can be one of:

{{#tabs name:"load-balancer-type" tabs:"lxc,oci,ovn,external" }}

{{#tab lxc }}

When using the `lxc` load balancer type, the infrastructure provider will launch an LXC container running haproxy. As control plane machines are created and deleted, the provider will update and automatically reload the backend configuration of the haproxy instance. This is similar to the behavior of the haproxy load balancer container in cluster-api-provider-docker.

The control plane endpoint of the cluster will be set to the IP address of the haproxy container. The haproxy container is a single-point-of-failure for accessing the control plane of the workload cluster, so it is not suitable for production deployments. However, it requires zero configuration, therefore it can be used for evaulation or development purposes.

The load balancer instance can be configured through the `spec.loadBalancer.lxc.instanceSpec` configuration fields. Unless a custom image source is set, the `haproxy` image is used from the [default simplestreams server](../reference/default-simplestreams-server.md).

The only requirement to use the `lxc` load balancer type is that the management cluster **must** be able to reach the load balancer container through its IP.

An example LXCCluster spec follows:

```yaml,hidelines=#
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: LXCCluster
metadata:
  name: example-cluster
spec:
#  secretRef:
#    name: example-secret
  loadBalancer:
    lxc:
      instanceSpec:
        flavor: c1-m1
        profiles: [default]
```

{{#/tab }}
{{#tab oci }}

- **Required server extensions**: [`oci_instance`](https://linuxcontainers.org/incus/docs/main/api-extensions/#instance-oci)

The `oci` load balancer type is similar to `lxc`. The only difference is that an OCI container running the kindest haproxy image is used instead. Similarly to `lxc`, when control plane machines are added or removed from the cluster, the provider will keep the haproxy configuration up to date.

The control plane endpoint of the cluster will be set to the IP address of the haproxy container. The haproxy container is a single-point-of-failure for accessing the control plane of the workload cluster, so it is not suitable for production deployments. However, it requires zero configuration, therefore it can be used for evaulation or development purposes.

The load balancer instance can be configured through the `spec.loadBalancer.oci.instanceSpec` configuration fields. Unless a custom image source is set, the `ghcr.io/lxc/cluster-api-provider-incus/haproxy:v20230606-42a2262b` (mirror of `kindest/haproxy`) image will be used.

Support for OCI containers was first added in Incus 6.5. Using the `oci` load balancer type when the `oci_instance` API extension is not supported will raise an error during the LXCCluster provisioning process.

The only requirement to use the `oci` load balancer type is that the management cluster **must** be able to reach the load balancer container through its IP.

An example LXCCluster spec follows:

```yaml,hidelines=#
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: LXCCluster
metadata:
  name: example-cluster
spec:
#  secretRef:
#    name: example-secret
  loadBalancer:
    oci:
      instanceSpec:
        flavor: c1-m1
        profiles: [default]
```

{{#/tab }}

{{#tab ovn }}

- **Required server extensions**: [`network_load_balancer`](https://linuxcontainers.org/incus/docs/main/api-extensions/#network-load-balancer), [`network_load_balancer_health_check`](https://linuxcontainers.org/incus/docs/main/api-extensions/#network-load-balancer-health-check)

The `ovn` load balancer type will create and manage an [OVN network load balancer](https://linuxcontainers.org/incus/docs/main/howto/network_load_balancers/) for the control plane endpoint. A backend is configured for each control plane machine on the cluster. As control plane machines are added or removed from the cluster, cluster-api-provider-incus will reconcile the backends of the network load balancer object accordingly.

Using the `ovn` load balancer type when the `network_load_balancer` and `network_load_balancer_health_check` API extensions are not supported will raise an error during the LXCCluster provisioning process.

As mentioned in the documentation, network load balancers are only supported for [OVN networks](https://linuxcontainers.org/incus/docs/main/reference/network_ovn/). The load balancer address must be chosen from the uplink network. The cluster administrator must ensure that:

- The management cluster can reach the OVN uplink network, so that it can connect to the workload cluster.
- The name of the ovn network is set in `spec.loadBalancer.ovn.networkName`.
- The list of profiles used for control plane machines use the same OVN network (such that the load balancer backends can be configured).
- The load balancer IP address is set in `spec.controlPlaneEndpoint.host`

### Example

Let's assume the following scenario:

- We have 3 cluster nodes `w01`, `w02`, `w03`.
- We have a network `UPLINK` network using the uplink interface `eno1.100` with subnet `10.100.0.0/16`, gateway `10.100.255.254/16` and DNS `1.1.1.1,1.0.0.1`. The range `10.100.3.10-10.100.3.100` has been allocated for OVN networks.
- We have a network `OVN` of type OVN, with subnet `192.168.1.1/24`. The external address of the OVN router is `10.100.3.10` (assigned automatically during creation).
- Profile `default` is using the `OVN` network, so instances are created in the OVN network.
- We want to use IP address `10.100.42.1` for the load balancer address.

{{#tabs name:"network" tabs:"uplink network,ovn network,profile" }}

{{#tab uplink network }}

```bash
incus network show UPLINK
```

```yaml,hidelines=#
config:
  dns.nameservers: 1.1.1.1,1.0.0.1
  ipv4.gateway: 10.100.255.254/16
  ipv4.ovn.ranges: 10.100.3.10-10.100.3.100
#description: Physical network for OVN routers
name: UPLINK
type: physical
used_by:
- /1.0/networks/default
managed: true
#status: Created
#locations:
#- w03
#- w01
#- w02
#project: default
```

{{#/tab }}

{{#tab ovn network }}

```bash
incus network show OVN
```

```yaml,hidelines=#
config:
  bridge.mtu: "1442"
  ipv4.address: 192.168.1.1/24
  ipv4.nat: "true"
  ipv6.address: none
  network: UPLINK
  volatile.network.ipv4.address: 10.100.3.10
#description: Initial OVN network
name: default
type: ovn
used_by:
- /1.0/profiles/default
managed: true
#status: Created
#locations:
#- w01
#project: default
```

{{#/tab }}

{{#tab profile }}

```bash
incus show profile default
```

```yaml,hidelines=#
#config: {}
#description: Default Incus profile
devices:
  eth0:
    name: eth0
    network: OVN
    type: nic
  root:
    path: /
    pool: local
    type: disk
name: default
#used_by: []
#project: default
```

{{#/tab }}
{{#/tabs }}

The appropriate LXCCluster spec would look like this:

```yaml,hidelines=#
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: LXCCluster
metadata:
  name: example-cluster
spec:
#  secretRef:
#    name: example-secret
  controlPlaneEndpoint:
    host: 10.100.42.1
    port: 6443
  loadBalancer:
    ovn:
      networkName: OVN
```

{{#/tab }}

{{#tab external }}

The `external` load balancer type will not provision anything for the cluster load balancer. Instead, something else like `kube-vip` should be used to configure a VIP for the control plane endpoint.

The cluster administrator must manually specify the control plane endpoint.

Consider the following scenario:
- We have a network `incusbr0` with CIDR `10.217.28.1/24`. We have limited the DHCP ranges to `10.217.28.10-10.217.28.200`, so we are free to use the rest of the IPs without conflicts.
- We want to use `10.217.28.242` as the control plane VIP.

```bash
incus network show incusbr0
```

```yaml,hidelines=#
config:
  ipv4.address: 10.217.28.1/24
  ipv4.dhcp.ranges: 10.217.28.10-10.217.28.200
  ipv4.nat: "true"
#description: ""
name: incusbr0
type: bridge
used_by:
- /1.0/profiles/default
managed: true
#status: Created
#locations:
#- none
#project: default
```

The LXCCluster in that case would be:

```yaml,hidelines=#
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: LXCCluster
metadata:
  name: example-cluster
spec:
#  secretRef:
#    name: example-secret
  controlPlaneEndpoint:
    host: 10.217.28.242
    port: 6443
  loadBalancer:
    external: {}
```

> **NOTE**: More configuration is needed to deploy kube-vip. For a full example, see the [kube-vip cluster template](../reference/templates/kube-vip.md)

{{#/tab }}

{{#/tabs }}

<!-- links -->
[`lxc`]: ./lxc.md
