# Kubeadm profile

## Privileged containers

In order for Kubernetes to work properly on LXC, the following profile is applied:

```yaml
# incus profile create kubeadm
# curl https://neoaggelos.github.io/cluster-api-provider-lxc/static/v0.1/profile.yaml | incus profile edit kubeadm

{{#include ../../static/v0.1/profile.yaml }}
```

## Unprivileged containers

When using unprivileged containers, the following profile is applied instead:

```yaml
# incus profile create kubeadm-unprivileged
# curl https://neoaggelos.github.io/cluster-api-provider-lxc/static/v0.1/profile.yaml | incus profile edit kubeadm-unprivileged

{{#include ../../static/v0.1/unprivileged.yaml }}
```
