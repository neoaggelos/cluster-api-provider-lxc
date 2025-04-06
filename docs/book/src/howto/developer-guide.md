# Developer Guide

This document describes the necessary steps tools to get started with developing and testing CAPL on a local environment.

## Table Of Contents

<!-- toc -->

## Setup environment

### Install pre-requisites

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

### Create kind management cluster

Create a kind cluster:

```bash
sudo kind create cluster --kubeconfig ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config
```

### Initialize ClusterAPI

We deploy core ClusterAPI providers and enable ClusterTopology feature gate:

```bash
export CLUSTER_TOPOLOGY=true
clusterctl init
```

### Initialize repository

Clone the [cluster-api-provider-lxc](https://github.com/neoaggelos/cluster-api-provider-lxc) repository with:

```bash
git clone https://github.com/neoaggelos/cluster-api-provider-lxc
cd cluster-api-provider-lxc
```

### Initialize infrastructure

{{#tabs name:"infrastructure" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

If Incus is not already installed on your machine, install latest stable version and initialize using [setup-incus.sh](https://github.com/neoaggelos/cluster-api-provider-lxc/blob/main/hack/scripts/ci/setup-incus.sh):

```bash
./hack/scripts/ci/setup-incus.sh
```

The script will perform the following steps:

- Install latest stable incus version
- Initialize incus daemon using default options
- Configure incus daemon to listen on `https://$hostip:8443`
- Configure client certificate for local incus deamon
- Create a secret `lxc-secret.yaml` on the local directory with infrastructure credentials for the local incus daemon.

{{#/tab }}

{{#tab Canonical LXD }}

If LXD is not already installed on your machine, install and initialize using [setup-lxd.sh](https://github.com/neoaggelos/cluster-api-provider-lxc/blob/main/hack/scripts/ci/setup-lxd.sh):

```bash
./hack/scripts/ci/setup-lxd.sh
```

The script will perform the following steps:

- Install Canonical LXD 5.21 snap
- Initialize LXD with default options
- Configure LXD daemon to listen on `https://$hostip:8443`
- Configure client certificate for local LXD deamon
- Create a secret `lxc-secret.yaml` on the local directory with infrastructure credentials for the local LXD daemon.
{{#/tab }}

{{#/tabs }}

Then, apply the `lxc-secret.yaml` on the cluster to create the infrastructure crendentials secret:

```bash
kubectl apply -f lxc-secret.yaml
```

## Running locally

First, deploy the CRDs with:

```bash
make install
```

Then, run the controller manager with:

```bash
make run V=4
```

### Deploy a test cluster

On a separate window, generate a cluster manifest and deploy:

```bash
export LOAD_BALANCER="lxc: {}"
export LXC_SECRET_NAME="lxc-secret"
export KUBERNETES_VERSION="v1.32.3"
export CONTROL_PLANE_MACHINE_COUNT=1
export WORKER_MACHINE_COUNT=1

clusterctl generate cluster c1 --from ./templates/cluster-template.yaml > c1.yaml
```

Deploy the cluster with:

```bash
kubectl apply -f ./templates/clusterclass-lxc-default.yaml
kubectl apply -f c1.yaml
```

## Running unit tests

```bash
make test
```

## Running e2e tests

First, build the e2e image with:

```bash
make e2e-image
```

Then, run the e2e tests with:

```bash
# run the e2e tests
make test-e2e

# run on existing cluster (NOTE: expects providers to be installed)
make test-e2e E2E_ARGS='-use-existing-cluster' KUBE_CONTEXT=kind-kind

# run in parallel
make test-e2e E2E_GINKGO_PARALLEL=2

# run specific tests
make test-e2e E2E_GINKGO_FOCUS='QuickStart OCI'
```

Unless specified, the e2e tests will use the default `local-https` remote from the client configuration.

## Running conformance tests

First, build the e2e image with:

```bash
make e2e-image
```

Then, run the conformance tests with:

```bash
# run upstream k8s conformance tests (full suite)
make test-conformance

# run upstream k8s conformance tests (fast)
make test-conformance-fast
```
