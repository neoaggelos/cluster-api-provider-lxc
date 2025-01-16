# Kubeadm profile

In order for Kubernetes to work properly on LXC, the following profile is applied:

```yaml
# incus profile create kubeadm
# curl https://neoaggelos.github.io/cluster-api-provider-lxc/static/v0.1/profile.yaml | incus profile edit kubeadm

{{#include ../../static/v0.1/profile.yaml }}
```
