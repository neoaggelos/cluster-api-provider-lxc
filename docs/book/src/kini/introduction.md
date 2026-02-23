# kini

`kini` is a tool for running local Kubernetes clusters using LXC container "nodes". `kini` was
primarily designed for testing `cluster-api-provider-incus`, but may also be used for local
development or CI.

`kini` is implemented as a wrapper around [`kind`](https://kind.sigs.k8s.io/). It replaces the
"docker" CLI with a shim executable, so that it creates LXC containers instead.

`kini` can be used as a standalone binary, and does not require `docker` or `kind` to be installed.
The `kini kind` sub-command allows running kind commands. For example, to create a single-node development
cluster on a local machine (where Incus is already installed), you can use:

```bash
kini kind create cluster
```

## Mission statement

The stated mission of `kini` project is to:
- Enable launching development Kubernetes clusters running in Incus containers, through the same well-known `kind` CLI commands.
- Support both Incus and Canonical LXD
- Minimize UX differences between platforms, and allowing 90% of the use cases to just work
- Primary driving use case is to enable development of CAPN provider, and simplify end-to-end testing scenarios

## Next Steps

Refer to [Kini > Quick Start](./quick-start.md) for instructions to install and use kini.

Refer to [Kini > Reasoning](./explanation/reasoning.md) for the reasoning behind building kini.

Refer to [Kini > Compatibility Notes](./explanation/compatibility.md) for compatibility notes for kini vs kind.
