# Quick Start

In this tutorial, we will deploy a single-node Incus (or Canonical LXD) server, use a local kind as a management cluster, deploy cluster-api-provider-incus and create a secret with credentials. Finally, we will provision a development workload cluster and interact with it.

## Table Of Contents

<!-- toc -->

## Requirements

1. A host running Ubuntu 24.04 (4 cores, 4GB RAM, 20GB disk)
2. Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) on your local environment
3. Install [kind](https://kind.sigs.k8s.io/) and [Docker](https://www.docker.com/)
4. Install [clusterctl](https://cluster-api.sigs.k8s.io/user/quick-start#install-clusterctl)

## Install pre-requisites

First, install necessary tools for launching and interacting with the management cluster:

```bash
# docker
curl https://get.docker.com | bash -x

# kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# clusterctl
curl -L {{#releaselink repo:"https://github.com/kubernetes-sigs/cluster-api" gomodule:"sigs.k8s.io/cluster-api" asset:"clusterctl-linux-amd64" version:"1.10.x" }} -o clusterctl
chmod +x ./clusterctl
sudo mv ./clusterctl /usr/local/bin/clusterctl

# kubectl
curl -L --remote-name-all "https://dl.k8s.io/release/v1.33.0/bin/linux/amd64/kubectl" -o ./kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```

## Setup management cluster

The easiest way to setup a management cluster is to use `kind`:

```bash
kind create cluster
```

> *NOTE*: If this fails, your user might not have permissions to call `docker` commands. One way to address is to use `sudo usermod -a -G docker $(whoami)`, and then start a new shell.

Initialize kind cluster as a ClusterAPI management cluster:

```bash
# Enable the ClusterTopology feature gate
export CLUSTER_TOPOLOGY=true

clusterctl init
```

## Prepare infrastructure

First, ensure the iptables FORWARD policy is set to ACCEPT. This is required because of how `docker` mangles the iptables rules on the host:

```bash
sudo iptables -P FORWARD ACCEPT
```

> **NOTE**: Unless the above is configured, LXC containers will not be able to contact each other.

{{#tabs name:"install" tabs:"Incus,Canonical LXD" }}
{{#tab Incus }}

Install `incus` from latest stable version:

```bash
curl https://pkgs.zabbly.com/get/incus-stable | sudo bash -x
```

Initialize incus with a default bridge and local disk, then expose HTTPS API on port 8443:

```bash
# get node IP address
ip_address="$(ip -o route get to 1.1.1.1 | sed -n 's/.*src \([0-9.]\+\).*/\1/p')"

sudo incus admin init --auto --network-address "$ip_address"
sudo incus network set incusbr0 ipv6.address=none
sudo incus cluster enable "$ip_address"
```

Generate a client certificate and key, and add it as a trusted client certificate:

```bash
incus remote generate-certificate
sudo incus config trust add-certificate ~/.config/incus/client.crt
```

Configure HTTPS remote to use `incus` without sudo:

```bash
incus remote add local-https "https://$(sudo incus config get core.https_address)" --accept-certificate
incus remote set-default local-https
```

Generate a Kubernetes secret `lxc-secret` with credentials to access the Incus HTTPS endpoint:

```bash
kubectl create secret generic lxc-secret \
  --from-literal=server="https://$(incus config get core.https_address)" \
  --from-literal=server-crt="$(cat ~/.config/incus/servercerts/local-https.crt)" \
  --from-literal=client-crt="$(cat ~/.config/incus/client.crt)" \
  --from-literal=client-key="$(cat ~/.config/incus/client.key)" \
  --from-literal=project="default"
```

{{#/tab }}
{{#tab Canonical LXD }}

Install `lxd`:

```bash
sudo snap install lxd --channel 5.21/stable
```

Initialize lxd with a default bridge and local disk, then expose HTTPS API on port 8443:

```bash
# get node IP address
ip_address="$(ip -o route get to 1.1.1.1 | sed -n 's/.*src \([0-9.]\+\).*/\1/p')"

sudo lxd init --auto --network-address "$ip_address"
sudo lxc network set lxdbr0 ipv6.address=none
sudo lxc cluster enable "$ip_address"
```

Generate a client certificate and key, and add it as a trusted client certificate:

```bash
token="$(sudo lxc config trust add --name client | tail -1)"

lxc remote add local-https --token "$token" "https://$(sudo lxc config get core.https_address)"
lxc remote set-default local-https
```

Generate a Kubernetes secret `lxc-secret` with credentials to access the LXD HTTPS endpoint:

```bash
kubectl create secret generic lxc-secret \
  --from-literal=server="https://$(lxc config get core.https_address)" \
  --from-literal=server-crt="$(cat ~/snap/lxd/common/config/servercerts/local-https.crt)" \
  --from-literal=client-crt="$(cat ~/snap/lxd/common/config/client.crt)" \
  --from-literal=client-key="$(cat ~/snap/lxd/common/config/client.key)" \
  --from-literal=project="default"
```

{{#/tab }}
{{#/tabs }}

After this step, you should now have your infrastructure ready and a Kubernetes secret with client credentials to access it.

## Deploy cluster-api-provider-incus

First, we need to configure clusterctl so that it knows about cluster-api-provider-incus:

```yaml
# ~/.cluster-api/clusterctl.yaml
{{#include ../static/v0.1/clusterctl.yaml }}
```

This can be done with the following commands:

```bash
mkdir -p ~/.cluster-api

curl -o ~/.cluster-api/clusterctl.yaml \
  https://lxc.github.io/cluster-api-provider-incus/static/v0.1/clusterctl.yaml
```

Then, initialize `incus` infrastructure provider:

```bash
clusterctl init -i incus
```

Wait for `capn-controller-manager` to become healthy

```bash
kubectl get pod -n capn-system
```

The output should look similar to this:

```bash
NAME                                       READY   STATUS    RESTARTS   AGE
capn-controller-manager-6668b99f89-sstlp   1/1     Running   0          2m33s
```

## Generate cluster manifest

We will create a cluster manifest using the [`default`](../reference/templates/default.md) flavor, which is also suitable for single-node testing.

List the cluster template variables:

```bash
clusterctl generate cluster c1 -i incus --flavor development --list-variables
```

Example output:

```bash
Required Variables:
  - KUBERNETES_VERSION
  - LOAD_BALANCER
  - LXC_SECRET_NAME

Optional Variables:
  - CLUSTER_NAME                    (defaults to c1)
  - CONTROL_PLANE_MACHINE_COUNT     (defaults to 1)
  - CONTROL_PLANE_MACHINE_DEVICES   (defaults to "[]")
  - CONTROL_PLANE_MACHINE_FLAVOR    (defaults to "c2-m4")
  - CONTROL_PLANE_MACHINE_PROFILES  (defaults to "[default]")
  - CONTROL_PLANE_MACHINE_TYPE      (defaults to "container")
  - DEPLOY_KUBE_FLANNEL             (defaults to "false")
  - INSTALL_KUBEADM                 (defaults to "false")
  - LXC_IMAGE_NAME                  (defaults to "")
  - POD_CIDR                        (defaults to "[10.244.0.0/16]")
  - PRIVILEGED                      (defaults to "true")
  - SERVICE_CIDR                    (defaults to "[10.96.0.0/12]")
  - WORKER_MACHINE_COUNT            (defaults to 0)
  - WORKER_MACHINE_DEVICES          (defaults to "[]")
  - WORKER_MACHINE_FLAVOR           (defaults to "c2-m4")
  - WORKER_MACHINE_PROFILES         (defaults to "[default]")
  - WORKER_MACHINE_TYPE             (defaults to "container")
```

Set configuration values (for more details, refer to the page of the [default cluster template](../reference/templates/default.md#configuration)):

```bash
# Use a haproxy container for cluster load balancer (sufficient for local development).
# Use the 'lxc-secret' secret with infrastructure credentials we generated previously.
# Deploy kube-flannel CNI on the workload cluster.

export LOAD_BALANCER='lxc: {}'
export LXC_SECRET_NAME=lxc-secret
export DEPLOY_KUBE_FLANNEL=true
```

Then generate a cluster manifest for a cluster with 1 control plane and 1 worker node, using:

```bash
# generate manifest in 'cluster.yaml'
clusterctl generate cluster c1 -i incus \
  --kubernetes-version v1.33.0 \
  --control-plane-machine-count 1 \
  --worker-machine-count 1 \
  > cluster.yaml
```

## Deploy cluster

```bash
kubectl apply -f cluster.yaml
```

The output should look similar to this:

```bash
clusterclass.cluster.x-k8s.io/capn-default created
kubeadmcontrolplanetemplate.controlplane.cluster.x-k8s.io/capn-default-control-plane created
lxcclustertemplate.infrastructure.cluster.x-k8s.io/capn-default-lxc-cluster created
lxcmachinetemplate.infrastructure.cluster.x-k8s.io/capn-default-control-plane created
lxcmachinetemplate.infrastructure.cluster.x-k8s.io/capn-default-default-worker created
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/capn-default-default-worker created
cluster.cluster.x-k8s.io/c1 created
```

## Wait for cluster to finish deployment

```bash
# describe cluster and infrastructure resources, useful to track deployment progress
clusterctl describe cluster c1

# get overview of running machines
kubectl get cluster,lxccluster,machine,lxcmachine
```

Example output while the cluster is being deployed:

```bash
# clusterctl describe cluster c1
NAME                                           READY  SEVERITY  REASON                                  SINCE  MESSAGE
Cluster/c1                                     False  Info      Bootstrapping @ Machine/c1-6n84z-lxj6v  4s     0 of 1 completed
├─ClusterInfrastructure - LXCCluster/c1-vtf7d  True                                                     18s
├─ControlPlane - KubeadmControlPlane/c1-6n84z  False  Info      Bootstrapping @ Machine/c1-6n84z-lxj6v  4s     0 of 1 completed
│ └─Machine/c1-6n84z-lxj6v                     False  Info      Bootstrapping                           4s     1 of 2 completed
└─Workers
  └─MachineDeployment/c1-md-0-v42br            False  Warning   WaitingForAvailableMachines             22s    Minimum availability requires 1 replicas, current 0 available
    └─Machine/c1-md-0-v42br-vh2wd-7sn5p        False  Info      WaitingForControlPlaneAvailable         6s     0 of 2 completed

# kubectl get cluster,lxccluster,machine,lxcmachine
NAME                          CLUSTERCLASS   PHASE         AGE   VERSION
cluster.cluster.x-k8s.io/c1   capn-default   Provisioned   22s   v1.33.0

NAME                                                  CLUSTER   LOAD BALANCER   READY   AGE
lxccluster.infrastructure.cluster.x-k8s.io/c1-vtf7d   c1        10.130.1.162    true    22s

NAME                                                 CLUSTER   NODENAME   PROVIDERID   PHASE          AGE   VERSION
machine.cluster.x-k8s.io/c1-6n84z-lxj6v              c1                                Provisioning   17s   v1.33.0
machine.cluster.x-k8s.io/c1-md-0-v42br-vh2wd-7sn5p   c1                                Pending        6s    v1.33.0

NAME                                                                   CLUSTER   MACHINE                     PROVIDERID   READY   AGE
lxcmachine.infrastructure.cluster.x-k8s.io/c1-6n84z-lxj6v              c1        c1-6n84z-lxj6v                                   17s
lxcmachine.infrastructure.cluster.x-k8s.io/c1-md-0-v42br-vh2wd-7sn5p   c1        c1-md-0-v42br-vh2wd-7sn5p                        6s
```

Once the cluster is deployed successfully, the output should look similar to:

```bash
# clusterctl describe cluster c1
NAME                                           READY  SEVERITY  REASON  SINCE  MESSAGE
Cluster/c1                                     True                     23s
├─ClusterInfrastructure - LXCCluster/c1-vtf7d  True                     54s
├─ControlPlane - KubeadmControlPlane/c1-6n84z  True                     23s
│ └─Machine/c1-6n84z-lxj6v                     True                     30s
└─Workers
  └─MachineDeployment/c1-md-0-v42br            True                     8s
    └─Machine/c1-md-0-v42br-vh2wd-7sn5p        True                     10s

# kubectl get cluster,lxccluster,machine,lxcmachine
NAME                          CLUSTERCLASS   PHASE         AGE   VERSION
cluster.cluster.x-k8s.io/c1   capn-default   Provisioned   59s   v1.33.0

NAME                                                  CLUSTER   LOAD BALANCER   READY   AGE
lxccluster.infrastructure.cluster.x-k8s.io/c1-vtf7d   c1        10.130.1.162    true    59s

NAME                                                 CLUSTER   NODENAME                    PROVIDERID                         PHASE     AGE   VERSION
machine.cluster.x-k8s.io/c1-6n84z-lxj6v              c1        c1-6n84z-lxj6v              lxc:///c1-6n84z-lxj6v              Running   54s   v1.33.0
machine.cluster.x-k8s.io/c1-md-0-v42br-vh2wd-7sn5p   c1        c1-md-0-v42br-vh2wd-7sn5p   lxc:///c1-md-0-v42br-vh2wd-7sn5p   Running   43s   v1.33.0

NAME                                                                   CLUSTER   MACHINE                     PROVIDERID                         READY   AGE
lxcmachine.infrastructure.cluster.x-k8s.io/c1-6n84z-lxj6v              c1        c1-6n84z-lxj6v              lxc:///c1-6n84z-lxj6v              true    54s
lxcmachine.infrastructure.cluster.x-k8s.io/c1-md-0-v42br-vh2wd-7sn5p   c1        c1-md-0-v42br-vh2wd-7sn5p   lxc:///c1-md-0-v42br-vh2wd-7sn5p   true    43s
```

> *NOTE*: `MachineDeployment` status requires the `Node` objects on the workload cluster to become `Ready`. If you did not set `DEPLOY_KUBE_FLANNEL=true`, the status of the `MachineDeployment` statusill not become `Ready` until you have deployed a CNI. You can do this in the next step.

We can also see the containers that have been created:

{{#tabs name:"launch" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
incus list user.cluster-name=c1
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
lxc list user.cluster-name=c1
```

{{#/tab }}

{{#/tabs }}

The output should look similar to:

```bash
+---------------------------+---------+------------------------+------+-----------+-----------+
|           NAME            |  STATE  |          IPV4          | IPV6 |   TYPE    | SNAPSHOTS |
+---------------------------+---------+------------------------+------+-----------+-----------+
| c1-6n84z-lxj6v            | RUNNING | 10.244.0.1 (cni0)      |      | CONTAINER | 0         |
|                           |         | 10.244.0.0 (flannel.1) |      |           |           |
|                           |         | 10.130.1.97 (eth0)     |      |           |           |
+---------------------------+---------+------------------------+------+-----------+-----------+
| c1-md-0-v42br-vh2wd-7sn5p | RUNNING | 10.244.1.0 (flannel.1) |      | CONTAINER | 0         |
|                           |         | 10.130.1.195 (eth0)    |      |           |           |
+---------------------------+---------+------------------------+------+-----------+-----------+
| c1-vtf7d-37a8e-lb         | RUNNING | 10.130.1.162 (eth0)    |      | CONTAINER | 0         |
+---------------------------+---------+------------------------+------+-----------+-----------+
```

## Access the cluster

First retrieve the kubeconfig file for the workload cluster

```bash
clusterctl get kubeconfig c1 > ~/.kube/c1.config
```

Then, retrieve the list of pods and nodes on the cluster with:

```bash
KUBECONFIG=~/.kube/c1.config kubectl get pod,node -A -o wide
```

Output should look similar to:

```bash
NAMESPACE      NAME                                         READY   STATUS    RESTARTS   AGE     IP             NODE                        NOMINATED NODE   READINESS GATES
kube-flannel   pod/kube-flannel-ds-d69xh                    1/1     Running   0          112s    10.130.1.195   c1-md-0-v42br-vh2wd-7sn5p   <none>           <none>
kube-flannel   pod/kube-flannel-ds-vh6rm                    1/1     Running   0          2m8s    10.130.1.97    c1-6n84z-lxj6v              <none>           <none>
kube-system    pod/coredns-674b8bbfcf-58976                 1/1     Running   0          2m8s    10.244.0.3     c1-6n84z-lxj6v              <none>           <none>
kube-system    pod/coredns-674b8bbfcf-bclrt                 1/1     Running   0          2m8s    10.244.0.2     c1-6n84z-lxj6v              <none>           <none>
kube-system    pod/etcd-c1-6n84z-lxj6v                      1/1     Running   0          2m13s   10.130.1.97    c1-6n84z-lxj6v              <none>           <none>
kube-system    pod/kube-apiserver-c1-6n84z-lxj6v            1/1     Running   0          2m16s   10.130.1.97    c1-6n84z-lxj6v              <none>           <none>
kube-system    pod/kube-controller-manager-c1-6n84z-lxj6v   1/1     Running   0          2m16s   10.130.1.97    c1-6n84z-lxj6v              <none>           <none>
kube-system    pod/kube-proxy-8cx9m                         1/1     Running   0          112s    10.130.1.195   c1-md-0-v42br-vh2wd-7sn5p   <none>           <none>
kube-system    pod/kube-proxy-zkwcc                         1/1     Running   0          2m8s    10.130.1.97    c1-6n84z-lxj6v              <none>           <none>
kube-system    pod/kube-scheduler-c1-6n84z-lxj6v            1/1     Running   0          2m16s   10.130.1.97    c1-6n84z-lxj6v              <none>           <none>

NAMESPACE   NAME                             STATUS   ROLES           AGE     VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION     CONTAINER-RUNTIME
            node/c1-6n84z-lxj6v              Ready    control-plane   2m18s   v1.33.0   10.130.1.97    <none>        Ubuntu 24.04.2 LTS   6.8.0-59-generic   containerd://2.1.0
            node/c1-md-0-v42br-vh2wd-7sn5p   Ready    <none>          112s    v1.33.0   10.130.1.195   <none>        Ubuntu 24.04.2 LTS   6.8.0-59-generic   containerd://2.1.0
```

## Delete cluster

Delete the workload cluster:

```bash
kubectl delete cluster c1
```

Delete the management cluster:

```bash
kind delete cluster
```

## Next Steps

- Expore the v1alpha2 [CRDs](../reference/api/v1alpha2/api.md)
- See list of example [Cluster Templates](../reference/templates/index.md)
- Read about the [Default Simplestreams Server](../reference/default-simplestreams-server.md)
