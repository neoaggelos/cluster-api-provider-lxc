# Quick Start

In this tutorial, we will deploy a single-node Incus (or Canonical LXD) server, use a local kind as a management cluster, deploy cluster-api-provider-incus and create a secret with credentials. Finally, we will provision a development workload cluster and interact with it.

## Table Of Contents

<!-- toc -->

## Requirements

1. A host running Ubuntu 24.04
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
curl -L {{#releaselink repo:"https://github.com/kubernetes-sigs/cluster-api" gomodule:"sigs.k8s.io/cluster-api" asset:"clusterctl-linux-amd64" version:"1.9.x" }} -o clusterctl
chmod +x ./clusterctl
sudo mv ./clusterctl /usr/local/bin/clusterctl

# kubectl
curl -L --remote-name-all "https://dl.k8s.io/release/v1.32.0/bin/linux/amd64/kubectl" -o ./kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```

## Setup management cluster

The easiest way to setup a management cluster is to use `kind`:

```bash
mkdir -p ~/.kube
sudo kind create cluster --kubeconfig ~/.kube/config

sudo chown $(id -u):$(id -g) ~/.kube/config
```

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
NAME                                      READY   STATUS    RESTARTS   AGE
capn-controller-manager-b6f789559-vtdvw   1/1     Running   0          4m20s
```

## Generate cluster manifest

We will create a cluster manifest of the `development` flavor, which is suitable for single-node testing.

List the cluster template variables:

```bash
clusterctl generate cluster c1 -i incus --flavor development --list-variables
```

Example output:

```bash
Required Variables:
  - CONTROL_PLANE_MACHINE_FLAVOR
  - CONTROL_PLANE_MACHINE_TYPE
  - KUBERNETES_VERSION
  - LXC_SECRET_NAME
  - WORKER_MACHINE_FLAVOR
  - WORKER_MACHINE_TYPE

Optional Variables:
  - CLUSTER_NAME                    (defaults to c1)
  - CONTROL_PLANE_MACHINE_COUNT     (defaults to 1)
  - CONTROL_PLANE_MACHINE_DEVICES   (defaults to "[]")
  - CONTROL_PLANE_MACHINE_PROFILES  (defaults to "[default]")
  - LOAD_BALANCER_MACHINE_FLAVOR    (defaults to "")
  - LOAD_BALANCER_MACHINE_PROFILES  (defaults to "[default]")
  - LXC_IMAGE_NAME                  (defaults to "")
  - LXC_LOAD_BALANCER_TYPE          (defaults to "lxc")
  - POD_CIDR                        (defaults to "[10.244.0.0/16]")
  - SERVICE_CIDR                    (defaults to "[10.96.0.0/12]")
  - WORKER_MACHINE_COUNT            (defaults to 0)
  - WORKER_MACHINE_DEVICES          (defaults to "[]")
  - WORKER_MACHINE_PROFILES         (defaults to "[default]")
```

Set configuration values:

```bash
{{#include ../../../../templates/cluster-template-development.rc }}
```

Then generate the cluster manifest using:

```bash
clusterctl generate cluster c1 -i incus --flavor development \
  --kubernetes-version v1.32.0 \
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
cluster.cluster.x-k8s.io/c1 created
lxccluster.infrastructure.cluster.x-k8s.io/c1 created
kubeadmcontrolplane.controlplane.cluster.x-k8s.io/c1-control-plane created
lxcmachinetemplate.infrastructure.cluster.x-k8s.io/c1-control-plane created
machinedeployment.cluster.x-k8s.io/c1-md-0 created
lxcmachinetemplate.infrastructure.cluster.x-k8s.io/c1-md-0 created
kubeadmconfigtemplate.bootstrap.cluster.x-k8s.io/c1-md-0 created
```

## Wait for cluster to finish deployment

```bash
# describe cluster and infrastructure resources, useful to track deployment progress
clusterctl describe cluster c1

# get overview of running machines
kubectl get cluster,lxccluster,machine,lxcmachine
```

Once the cluster is deployed successfully, the output should look similar to:

```bash
# clusterctl describe cluster c1
NAME                                                   READY  SEVERITY  REASON                       SINCE  MESSAGE
Cluster/c1                                             True                                          33m
├─ClusterInfrastructure - LXCCluster/c1                True                                          34m
├─ControlPlane - KubeadmControlPlane/c1-control-plane  True                                          33m
│ └─Machine/c1-control-plane-sn5ww                     True                                          34m
└─Workers
  └─MachineDeployment/c1-md-0                          False  Warning   WaitingForAvailableMachines  33m     Minimum availability requires 1 replicas, current 0 available
    └─Machine/c1-md-0-zj8bf-cbwnc                      True                                          33m

# kubectl get cluster,lxccluster,machine,lxcmachine
NAME                          CLUSTERCLASS   PHASE         AGE   VERSION
cluster.cluster.x-k8s.io/c1                  Provisioned   19m

NAME                                            CLUSTER   LOAD BALANCER   READY   AGE
lxccluster.infrastructure.cluster.x-k8s.io/c1   c1        10.225.32.110   true    19m

NAME                                              CLUSTER   NODENAME                 PROVIDERID                      PHASE     AGE   VERSION
machine.cluster.x-k8s.io/c1-control-plane-sn5ww   c1        c1-control-plane-sn5ww   lxc:///c1-control-plane-sn5ww   Running   18m   v1.32.0
machine.cluster.x-k8s.io/c1-md-0-zj8bf-cbwnc      c1        c1-md-0-zj8bf-cbwnc      lxc:///c1-md-0-zj8bf-cbwnc      Running   18m   v1.32.0

NAME                                                                CLUSTER   MACHINE                  PROVIDERID                      READY   AGE
lxcmachine.infrastructure.cluster.x-k8s.io/c1-control-plane-sn5ww   c1        c1-control-plane-sn5ww   lxc:///c1-control-plane-sn5ww   true    18m
lxcmachine.infrastructure.cluster.x-k8s.io/c1-md-0-zj8bf-cbwnc      c1        c1-md-0-zj8bf-cbwnc      lxc:///c1-md-0-zj8bf-cbwnc      true    18m
```

> *NOTE*: The `MachineDeployment` status is expected, as no CNI has been deployed for Nodes to become Ready. We will deploy kube-flannel in the next step.

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
+------------------------+---------+----------------------+------+-----------+-----------+---------------+
|          NAME          |  STATE  |         IPV4         | IPV6 |   TYPE    | SNAPSHOTS |   LOCATION    |
+------------------------+---------+----------------------+------+-----------+-----------+---------------+
| c1-control-plane-sn5ww | RUNNING | 10.225.32.7 (eth0)   |      | CONTAINER | 0         | 172.31.33.136 |
+------------------------+---------+----------------------+------+-----------+-----------+---------------+
| c1-md-0-zj8bf-cbwnc    | RUNNING | 10.225.32.246 (eth0) |      | CONTAINER | 0         | 172.31.33.136 |
+------------------------+---------+----------------------+------+-----------+-----------+---------------+
| default-c1-lb          | RUNNING | 10.225.32.110 (eth0) |      | CONTAINER | 0         | 172.31.33.136 |
+------------------------+---------+----------------------+------+-----------+-----------+---------------+
```

## Access the cluster

First retrieve the kubeconfig file for the workload cluster

```bash
clusterctl get kubeconfig c1 > ~/.kube/c1.config
```

Optionally, deploy kube-flannel on the cluster:

```bash
KUBECONFIG=~/.kube/c1.config kubectl apply -f \
    https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml
```

Then, retrieve the list of pods and nodes on the cluster with:

```bash
KUBECONFIG=~/.kube/c1.config kubectl get pod,node -A -o wide
```

Output should look similar to:

```bash
NAMESPACE      NAME                                                 READY   STATUS    RESTARTS   AGE   IP              NODE                     NOMINATED NODE   READINESS GATES
kube-flannel   pod/kube-flannel-ds-5z8h5                            1/1     Running   0          38s   10.225.32.7     c1-control-plane-sn5ww   <none>           <none>
kube-flannel   pod/kube-flannel-ds-f9qxb                            1/1     Running   0          38s   10.225.32.246   c1-md-0-zj8bf-cbwnc      <none>           <none>
kube-system    pod/coredns-668d6bf9bc-dkv9r                         1/1     Running   0          37m   10.244.1.3      c1-md-0-zj8bf-cbwnc      <none>           <none>
kube-system    pod/coredns-668d6bf9bc-kvc27                         1/1     Running   0          37m   10.244.1.2      c1-md-0-zj8bf-cbwnc      <none>           <none>
kube-system    pod/etcd-c1-control-plane-sn5ww                      1/1     Running   0          37m   10.225.32.7     c1-control-plane-sn5ww   <none>           <none>
kube-system    pod/kube-apiserver-c1-control-plane-sn5ww            1/1     Running   0          37m   10.225.32.7     c1-control-plane-sn5ww   <none>           <none>
kube-system    pod/kube-controller-manager-c1-control-plane-sn5ww   1/1     Running   0          37m   10.225.32.7     c1-control-plane-sn5ww   <none>           <none>
kube-system    pod/kube-proxy-f6cq7                                 1/1     Running   0          37m   10.225.32.7     c1-control-plane-sn5ww   <none>           <none>
kube-system    pod/kube-proxy-hkpl9                                 1/1     Running   0          36m   10.225.32.246   c1-md-0-zj8bf-cbwnc      <none>           <none>
kube-system    pod/kube-scheduler-c1-control-plane-sn5ww            1/1     Running   0          37m   10.225.32.7     c1-control-plane-sn5ww   <none>           <none>

NAMESPACE   NAME                          STATUS   ROLES           AGE   VERSION   INTERNAL-IP     EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION   CONTAINER-RUNTIME
            node/c1-control-plane-sn5ww   Ready    control-plane   37m   v1.32.0   10.225.32.7     <none>        Ubuntu 24.04.1 LTS   6.8.0-1021-aws   containerd://1.7.24
            node/c1-md-0-zj8bf-cbwnc      Ready    <none>          36m   v1.32.0   10.225.32.246   <none>        Ubuntu 24.04.1 LTS   6.8.0-1021-aws   containerd://1.7.24
```

After deploying kube-flannel, check that the MachineDeployment is now healthy:

```bash
clusterctl describe cluster c1
```

The output should now look like this:

```bash
NAME                                                   READY  SEVERITY  REASON  SINCE  MESSAGE
Cluster/c1                                             True                     40m
├─ClusterInfrastructure - LXCCluster/c1                True                     41m
├─ControlPlane - KubeadmControlPlane/c1-control-plane  True                     40m
│ └─Machine/c1-control-plane-sn5ww                     True                     40m
└─Workers
  └─MachineDeployment/c1-md-0                          True                     86s
    └─Machine/c1-md-0-zj8bf-cbwnc                      True                     40m
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
