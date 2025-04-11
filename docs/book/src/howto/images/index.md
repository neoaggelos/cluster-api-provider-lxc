# Build base images

The cluster-api-provider-incus project builds and pushes base images on the [default simplestreams server](../../reference/default-simplestreams-server.md).

Images on the default server do not support all Kubernetes versions, and availability might vary. Follow the links below for instructions to build base images for:

- [`kubeadm`](./kubeadm.md): used to launch the Kubernetes control plane and worker node machines
- [`haproxy`](./haproxy.md): used to launch the load balancer container in development clusters

> **NOTE**: The images on the default simplesteams server are meant for evaluation and development purposes only. Administrators should build and maintain their own images for production clusters.
