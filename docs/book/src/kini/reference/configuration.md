# Configuration Reference

`kini` uses configuration from environment variables.

## Table Of Contents

<!-- toc -->

## Runtime configuration

| Environment variable     | Options          | Default                        | Description                                                                                                                 |
| ------------------------ | ---------------- | ------------------------------ | --------------------------------------------------------------------------------------------------------------------------- |
| `KINI_MODE`              | `oci` <br> `lxc` | `oci` (Incus) <br> `lxc` (LXD) | See [Compatibility Notes > LXC and OCI instances](../explanation/compatibility.md)                                          |
| `KINI_MOUNT_UNIX_SOCKET` | `true`/`false`   | `false`                        | Mount default UNIX socket for interacting with the Incus API. This is mainly used for cluster-api-provider-incus e2e tests. |
| `KINI_UNPRIVILEGED`      | `true`/`false`   | `false`                        | See [Privileged vs unprivileged containers](../explanation/compatibility.md#privileged-vs-unprivileged-containers))         |
| `KINI_CACHE`             | `<dir>`          | `~/.cache/kini`                | Directory to use for caching OCI image tarballs, used by the `docker save/load` commands.                                   |

## Config file

| Environment variable | Options     | Default   | Description                                                                                       |
| -------------------- | ----------- | --------- | ------------------------------------------------------------------------------------------------- |
| `KINI_CONFIG`        | `<file>`    | `<unset>` | Path to config file to use for Incus server configuration. If unset, use the default config file. |
| `KINI_REMOTE`        | `<name>`    | `<unset>` | Name of remote server to connect to. If unset, use the default remote.                            |
| `KINI_PROJECT`       | `<project>` | `<unset>` | Override the project name where instances will be launched. If unset, use the default project.    |

## Troubleshooting

| Environment variable | Options  | Default   | Description                                                            |
| -------------------- | -------- | --------- | ---------------------------------------------------------------------- |
| `KINI_DOCKER_LOG`    | `<file>` | `<unset>` | Write debug logs for `kini docker` invocations to `a.log`              |
| `KINI_DOCKER_LOGV`   | `5`      | `<unset>` | Adjust verbosity level for debugging logs of `kini docker` invocations |
