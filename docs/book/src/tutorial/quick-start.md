# Getting started

In this tutorial, we will deploy a single-node Incus (or Canonical LXD) server, use a local kind as a management cluster, deploy cluster-api-provider-lxc and create a secret with credentials. Finally, we will provision a development workload cluster and interact with it.

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
kind create cluster
```

Initialize kind cluster as a ClusterAPI management cluster with:

```bash
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

Generate a Kubernetes secret `lxc-secret` with credentials to access the Incus HTTPS endpoint:

```bash
kubectl create secret generic lxc-secret \
  --from-literal=server="https://$(sudo incus config get core.https_address)" \
  --from-literal=server-crt="$(sudo cat /var/lib/incus/cluster.crt)" \
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

lxc remote add lxd "https://$(sudo lxc config get core.https_address)" \
    --accept-certificate --token "$token"
```

Generate a Kubernetes secret `lxc-secret` with credentials to access the LXD HTTPS endpoint:

```bash
kubectl create secret generic lxc-secret \
  --from-literal=server="https://$(sudo lxc config get core.https_address)" \
  --from-literal=server-crt="$(sudo cat /var/snap/lxd/common/lxd/cluster.crt)" \
  --from-literal=client-crt="$(cat ~/snap/lxd/common/config/client.crt)" \
  --from-literal=client-key="$(cat ~/snap/lxd/common/config/client.key)" \
  --from-literal=project="default"
```

{{#/tab }}
{{#/tabs }}

After this step, you should now have your infrastructure ready and a Kubernetes secret with client credentials to access it.

## Deploy cluster-api-provider-lxc

> **TODO**: this should be replaced with `clusterctl init -i lxc` once we have a public release

```bash
kubectl apply -f https://neoaggelos.github.io/cluster-api-provider-lxc/static/v0.1/release/infrastructure-components.yaml
```

Then, wait for `capl-controller-manager` to become healthy

```bash
kubectl get pod -n capl-system
```

The output should look similar to this:

```bash
NAME                                      READY   STATUS    RESTARTS   AGE
capl-controller-manager-b6f789559-vtdvw   1/1     Running   0          4m20s
```

## Generate cluster manifest

> **TODO**: this should be replaced with `clusterctl generate cluster c1 --flavor development` once we have a public release

List the cluster template variables:

```bash
clusterctl generate cluster c1 --list-variables --from https://neoaggelos.github.io/cluster-api-provider-lxc/static/v0.1/release/cluster-template-development.yaml
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
  - CONTROL_PLANE_MACHINE_PROFILES  (defaults to "default")
  - LXC_LOAD_BALANCER_FLAVOR        (defaults to "")
  - LXC_LOAD_BALANCER_PROFILES      (defaults to "default")
  - LXC_LOAD_BALANCER_TYPE          (defaults to "lxc")
  - POD_CIDR                        (defaults to "['10.244.0.0/16']")
  - SERVICE_CIDR                    (defaults to "['10.96.0.0/12']")
  - WORKER_MACHINE_COUNT            (defaults to 0)
  - WORKER_MACHINE_PROFILES         (defaults to "default")
```

Set configuration values:

```bash
{{#include ../../../../templates/development.rc }}
```

Then generate the cluster manifest using:

```bash
clusterctl generate cluster c1 \
  --from https://neoaggelos.github.io/cluster-api-provider-lxc/static/v0.1/release/cluster-template-development.yaml \
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
# describe cluster and infrastructure resources
clusterctl describe cluster c1

# get overview of running machines
kubectl get cluster,lxccluster,machine,lxcmachine
```

Once the cluster is deployed successfully, the output should look similar to:

```bash
NAME                          CLUSTERCLASS   PHASE         AGE     VERSION
cluster.cluster.x-k8s.io/c1                  Provisioned   6m20s

NAME                                            CLUSTER   LOAD BALANCER    READY   AGE
lxccluster.infrastructure.cluster.x-k8s.io/c1   c1        10.141.186.196   true    6m20s

NAME                                              CLUSTER   NODENAME                 PROVIDERID                      PHASE     AGE     VERSION
machine.cluster.x-k8s.io/c1-control-plane-zcz86   c1        c1-control-plane-zcz86   lxc:///c1-control-plane-zcz86   Running   6m4s    v1.32.0
machine.cluster.x-k8s.io/c1-md-0-7jls6-p268m      c1        c1-md-0-7jls6-p268m      lxc:///c1-md-0-7jls6-p268m      Running   5m50s   v1.32.0

NAME                                                                CLUSTER   MACHINE                  PROVIDERID                      READY   AGE
lxcmachine.infrastructure.cluster.x-k8s.io/c1-control-plane-zcz86   c1        c1-control-plane-zcz86   lxc:///c1-control-plane-zcz86   true    6m4s
lxcmachine.infrastructure.cluster.x-k8s.io/c1-md-0-7jls6-p268m      c1        c1-md-0-7jls6-p268m      lxc:///c1-md-0-7jls6-p268m      true    5m50s
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
NAMESPACE      NAME                                                 READY   STATUS    RESTARTS   AGE     IP               NODE                     NOMINATED NODE   READINESS GATES
kube-flannel   pod/kube-flannel-ds-cnfnb                            1/1     Running   0          4m44s   10.141.186.93    c1-control-plane-zcz86   <none>           <none>
kube-flannel   pod/kube-flannel-ds-r5g2c                            1/1     Running   0          4m44s   10.141.186.115   c1-md-0-7jls6-p268m      <none>           <none>
kube-system    pod/coredns-668d6bf9bc-fnlpx                         1/1     Running   0          6m14s   10.244.0.2       c1-control-plane-zcz86   <none>           <none>
kube-system    pod/coredns-668d6bf9bc-z9ljw                         1/1     Running   0          6m14s   10.244.0.4       c1-control-plane-zcz86   <none>           <none>
kube-system    pod/etcd-c1-control-plane-zcz86                      1/1     Running   0          6m18s   10.141.186.93    c1-control-plane-zcz86   <none>           <none>
kube-system    pod/kube-apiserver-c1-control-plane-zcz86            1/1     Running   0          6m21s   10.141.186.93    c1-control-plane-zcz86   <none>           <none>
kube-system    pod/kube-controller-manager-c1-control-plane-zcz86   1/1     Running   0          6m19s   10.141.186.93    c1-control-plane-zcz86   <none>           <none>
kube-system    pod/kube-proxy-68qhk                                 1/1     Running   0          5m37s   10.141.186.115   c1-md-0-7jls6-p268m      <none>           <none>
kube-system    pod/kube-proxy-6w9zm                                 1/1     Running   0          6m14s   10.141.186.93    c1-control-plane-zcz86   <none>           <none>
kube-system    pod/kube-scheduler-c1-control-plane-zcz86            1/1     Running   0          6m20s   10.141.186.93    c1-control-plane-zcz86   <none>           <none>

NAMESPACE   NAME                          STATUS   ROLES           AGE     VERSION   INTERNAL-IP      EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION       CONTAINER-RUNTIME
            node/c1-control-plane-zcz86   Ready    control-plane   6m23s   v1.32.0   10.141.186.93    <none>        Ubuntu 24.04.1 LTS   5.15.0-130-generic   containerd://1.7.24
            node/c1-md-0-7jls6-p268m      Ready    <none>          5m37s   v1.32.0   10.141.186.115   <none>        Ubuntu 24.04.1 LTS   5.15.0-130-generic   containerd://1.7.24
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

- Expore the v1alpha1 [CRDs](../reference/api/v1alpha1/api.md)
- See list of example [Cluster Templates](../reference/templates/index.md)
- Read about the [Defaul Simplestreams Server](../reference/default-simplestreams-server.md)
