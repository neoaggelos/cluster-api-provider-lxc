#!/bin/bash -e

# Usage:
#
# 1). Install latest version
#     $ curl https://capn.linuxcontainers.org/get-kini.sh | bash -x
#
# 2). Install in local directory (without root)
#     $ curl https://capn.linuxcontainers.org/get-kini.sh | SUDO= KINI_INSTALL_DIR=~/.local/bin bash
#
# 3). Install specific version
#     $ curl https://capn.linuxcontainers.org/get-kini.sh | KINI_DOWNLOAD_URL=https://github.com/lxc/cluster-api-provider/releases/v0.8.2/download bash

set -eu

KINI_INSTALL_DIR="${KINI_INSTALL_DIR:-/usr/local/bin}"
KINI_DOWNLOAD_URL="${KINI_DOWNLOAD_URL:-https://github.com/lxc/cluster-api-provider/releases/latest/download}"

# infer SUDO, if not running as root
SUDO="${SUDO:-sudo}"
if [ "$(id -u)" == "0" ]; then SUDO=; fi

# infer OS (linux, darwin)
case "$(uname)" in
  Linux) OS=linux ;;
  Darwin) OS=darwin ;;
  *) echo "Unsupported OS $(uname)"; exit 1 ;;
esac

# infer ARCH (amd64, arm64)
case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "Unsupported ARCH $(uname -m)"; exit 1 ;;
esac

echo "Downloading ${KINI_DOWNLOAD_URL}/kini-${OS}-${ARCH}"

# download kini binary
(
  set -x
  mkdir -p "/tmp/kini$$"
  curl "${KINI_DOWNLOAD_URL}/kini-${OS}-${ARCH}" -f -o "/tmp/kini$$/kini"

  chmod +x "/tmp/kini$$/kini"

  "/tmp/kini$$/kini" version

  $SUDO mkdir -p "${KINI_INSTALL_DIR}"
  $SUDO mv "/tmp/kini$$/kini" "${KINI_INSTALL_DIR}/kini"
  rm -rf "/tmp/kini$$"
)

kini version
echo "Successfully downloaded kini at ${KINI_INSTALL_DIR}"
