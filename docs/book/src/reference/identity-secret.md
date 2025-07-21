# Identity secret

Each LXCCluster must specify a reference to a secret with credentials that can be used to reach the remote Incus or LXD instance:

```yaml
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: LXCCluster
metadata:
  name: example-cluster
spec:
  secretRef:
    name: incus-secret
```

## Identity secret format

The `incus-secret` must exist in the **same** namespace as the **LXCCluster** object. The following configuration fields can be set:

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: incus-secret
  labels:
    # [optional]
    # If set, "clusterctl move" will include this secret when moving Cluster API objects
    # to the target management cluster. This is required if you are planning on moving
    # the cluster after provisioning.
    clusterctl.cluster.x-k8s.io/move: "true"
stringData:
  # [required]
  # 'server' is the HTTPS URL of the Incus or LXD server.
  #
  # Note that Incus and LXD only expose a local UNIX socket by default. In order to
  # expose a running server on the network, you need to do something like:
  #
  #   $ sudo incus config set core.https_address=:8443
  #
  # See https://linuxcontainers.org/incus/docs/main/howto/server_expose/#server-expose
  server: https://10.0.1.1:8443

  # [required]
  # 'server-crt' is the cluster certificate. Can be retrieved from a running instance with:
  #
  #   $ openssl s_client -connect 10.0.1.1:8443 </dev/null 2>/dev/null | openssl x509
  server-crt: |
    -----BEGIN CERTIFICATE-----
    MIIB9DCCAXqgAwIBAgIQa+btN/ftie8EniUcMM7QeTAKBggqhkjOPQQDAzAuMRkw
    FwYDVQQKExBMaW51eCBDb250YWluZXJzMREwDwYDVQQDDAhyb290QHcwMTAeFw0y
    NTAxMDMxODEyNDdaFw0zNTAxMDExODEyNDdaMC4xGTAXBgNVBAoTEExpbnV4IENv
    bnRhaW5lcnMxETAPBgNVBAMMCHJvb3RAdzAxMHYwEAYHKoZIzj0CAQYFK4EEACID
    YgAEj4f7cUnwXaehJI3jXVsvdLLPRmc2s+qMSNhwM1XFrXM7J57R9UkODwGuDrT8
    39w74Cm9kaDptJt7Ze+ESfBMSo+C0M9W1zqsCwbD96lzkWPGnBGz4xCo/akJQJ/X
    /hpYo10wWzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYD
    VR0TAQH/BAIwADAmBgNVHREEHzAdggN3MDGHBH8AAAGHEAAAAAAAAAAAAAAAAAAA
    AAEwCgYIKoZIzj0EAwMDaAAwZQIxANpf3eGxsFElwWNxzBxdMUQEST2tzJxzeslP
    8bZvAJsRF39LOicqKbwozcJgV/39LQIwYHKtI686IoBUxK0qGXn0C5ltSG7Y6Gun
    bZECNaleEKUa+e9bZQuhh13yWcx+EB7C
    -----END CERTIFICATE-----

  # [required]
  # 'client-crt' is the client certificate to use for authentication. Can be generated with:
  #
  #   $ incus remote generate-certificate
  #   $ cat ~/.config/incus/client.crt
  #
  # The certificate must be added as a trusted client certificate on the remote server, e.g. with:
  #
  #   $ cat ~/.config/incus/client.crt | sudo incus config trust add-certificate - --force-local
  client-crt: |
    -----BEGIN CERTIFICATE-----
    MIIB3DCCAWGgAwIBAgIRAJrtUMjnEBuGqDhqr7J99VUwCgYIKoZIzj0EAwMwNTEZ
    MBcGA1UEChMQTGludXggQ29udGFpbmVyczEYMBYGA1UEAwwPdWJ1bnR1QGRhbW9j
    bGVzMB4XDTI0MTIxNTIxNDUwMloXDTM0MTIxMzIxNDUwMlowNTEZMBcGA1UEChMQ
    TGludXggQ29udGFpbmVyczEYMBYGA1UEAwwPdWJ1bnR1QGRhbW9jbGVzMHYwEAYH
    KoZIzj0CAQYFK4EEACIDYgAErErnYTBj2fCHeMiEllgMvpbJcGYMHAvB0l3D0jbb
    q6KP4Y0nxTwsLQqgiEZ3pUuQ7Q4G7yvjV8mn4a0Y4wf2J7bbJxnN9vkopeHqmqil
    TFbDRa/kkdEVRGkgQ16B1lF0ozUwMzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAww
    CgYIKwYBBQUHAwIwDAYDVR0TAQH/BAIwADAKBggqhkjOPQQDAwNpADBmAjEAi4Ml
    2NHVg8hD6UVt+Mp6wkDWIDlegNb8mR8tcEQe4+Xs7htrswLegPVndvQeM6thAjEA
    97SouLFMm8OnZr9kKdMr3N3hx3ngV7Fx9hUm4gCKoOLFU2xEHo/ytwnKAKsRGrss
    -----END CERTIFICATE-----

  # [required]
  # 'client-key' is the private key for the client certificate to use for authentication.
  #
  #   $ cat ~/.config/incus/client.key
  client-key: |
    -----BEGIN EC PRIVATE KEY-----
    MIGkAgEBBDDC7pty/YA+IFDQx4aP2hXpw5S7rwTat5POJsCQMM06kn2qY+PoITY+
    7xTGg1xBeL6gBwYFK4EEACKhZANiAASsSudhMGPZ8Id4yISWWAy+lslwZgwcC8HS
    XcPSNturoo/hjSfFPCwtCqCIRnelS5DtDgbvK+NXyafhrRjjB/YnttsnGc32+Sil
    4eqaqKVMVsNFr+SR0RVEaSBDXoHWUXQ=
    -----END EC PRIVATE KEY-----

  # [optional]
  # 'project' is the name of the project to launch instances in. if not set, "default" is used.
  project: default

  # [optional]
  # 'insecure-skip-verify' will disable checking the server certificate when connecting to the
  # remote server. if not set, "false" is assumed.
  insecure-skip-verify: "false"
