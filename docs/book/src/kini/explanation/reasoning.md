# Reasoning

## Why build kini?

`kind` is commonly used in the ClusterAPI world in order to bring up management clusters for development purposes. `kind` requires a Docker runtime to be installed on the machine, in order to launch privileged containers and start up a local Kubernetes cluster.

However, the most common configuration for single node `Docker` and `Incus` configurations comes with a few caveats:
- Incus will usually configure a `incusbr0` local bridge, which is used by instances.
- Docker will configure a separate `dockerbr0` local bridge, and by default aggressively adjust iptables rules.

In this mode, the Docker containers (`kind-control-plane`) and the CAPN machines are in different networks, and the default iptables rules do not allow them to communicate with each other.

Motivated by the fact that recent Incus versions have added support for OCI containers, `kini` implements a shim layer between `kind` and the Docker daemon. As such, instead of creating a Docker container, it will instead launch an Incus container, using the same storage and networking as other machines.
