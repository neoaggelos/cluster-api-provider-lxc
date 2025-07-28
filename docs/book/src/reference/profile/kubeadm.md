# Kubeadm profile

## Privileged containers

In order for Kubernetes to work properly on LXC, the following profile is applied:

```yaml
# incus profile create kubeadm
# curl https://capn.linuxcontainers.org/static/v0.1/profile.yaml | incus profile edit kubeadm

{{#include ../../static/v0.1/profile.yaml }}
```

## Unprivileged containers

When using unprivileged containers, the following profile is applied instead:

```yaml
# incus profile create kubeadm-unprivileged
# curl https://capn.linuxcontainers.org/static/v0.1/unprivileged.yaml | incus profile edit kubeadm-unprivileged

{{#include ../../static/v0.1/unprivileged.yaml }}
```

## Unprivileged containers (Canonical LXD)

When using unprivileged containers with Canonical LXD, it is also required to enable `security.nesting` and disable apparmor:

```yaml
# lxc profile create kubeadm-unprivileged
# curl https://capn.linuxcontainers.org/static/v0.1/unprivileged-lxd.yaml | lxc profile edit kubeadm-unprivileged

{{#include ../../static/v0.1/unprivileged-lxd.yaml }}
```

## Privileged kind containers

When using privileged kind containers, the following profile is applied:

```yaml
# incus profile create kind
# curl https://capn.linuxcontainers.org/static/v0.1/kind.yaml | lxc profile edit kind

{{#include ../../static/v0.1/kind.yaml }}
```

## Unprivileged kind containers

When using unprivileged kind containers, the following profile is applied:

```yaml
# incus profile create kind-unprivileged
# curl https://capn.linuxcontainers.org/static/v0.1/kind-unprivileged.yaml | lxc profile edit kind-unprivileged

{{#include ../../static/v0.1/kind.yaml }}
```
