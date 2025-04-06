# Unprivileged containers

When using `instanceType: container`, CAPL will launch an LXC container for each cluster node. In order for Kubernetes and the container runtime to work, CAPL launches `privileged` containers by default.

However, privileged containers can pose security risks, especially in multi-tenant deployments. In such scenarios, if an adversary workload takes control of the kubelet, it can use the `privileged` capabilities to escape the container boundaries and affect workloads of other tenants or even fully take over the hypervisor.

In order to address these security risks, it is possible to use unprivileged containers instead.

## Using unprivileged containers

To use unprivileged containers, use the [default cluster template](../reference/templates/default.md) and set `PRIVILEGED=false`.

Unprivileged containers require extra configuration on the container runtime. This configuration is available in the kubeadm images starting from version  **v1.32.3**.

## Running Kubernetes in unprivileged containers

In order for Kubernetes to work inside an unprivileged containers, configuration of containerd, kubelet and kube-proxy is adjusted, in accordance with [the upstream project documentation](https://kubernetes.io/docs/tasks/administer-cluster/kubelet-in-userns/).

In particular, the following configuration adjustments are performed:

### kubelet

- use `cgroupDriver: cgroupfs`
- add feature gate `KubeletInUserNamespace: true`

When using the default cluster template, these are applied on the nodes through a KubeletConfiguration patch.

### containerd

- set `disable_apparmor = true`
- set `restrict_oom_score_adj = true`
- set `disable_hugetlb_controller = true`
- set `SystemdCgroup = false`

When using the default images, the `/opt/configure-containerd-unprivileged-mode.sh` script can be used to reconfigure containerd with the options above.

The default cluster template will run this script as `preKubeadmCommands` as needed.

## Support in pre-built kubeadm images

Unprivileged containers are supported with the pre-built kubeadm images starting from version **v1.32.3**.

## Limitations in unprivileged containers

Known limitations apply when using unprivileged containers, e.g. consuming NFS volumes. See [Caveats](https://kubernetes.io/docs/tasks/administer-cluster/kubelet-in-userns/#caveats) and [Caveats and Future work](https://rootlesscontaine.rs/caveats/) for more details.

Similar limitations might apply for the CNI of the cluster. `kube-flannel` with the vxlan backend is known to work.

## Testing

The above have been tested with Incus 6.10+ on Kernel 6.8 or newer.
