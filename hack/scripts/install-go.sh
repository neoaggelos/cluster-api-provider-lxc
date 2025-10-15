#!/bin/bash -e

# Usage:
#    $ install-go.sh              # source Go version from go.mod
#    $ install-go.sh 1.25.3       # install specific Go version
#
# Description:
#    - Download go version and install under /usr/local/go
#    - Create symlinks for go and gofmt in /usr/local/bin

VERSION="${1:-}"
if [ -z "${VERSION}" ]; then
  DIR="$(realpath "$(dirname "${0}")")"
  VERSION="$(cat "${DIR}/../../go.mod" | grep "^go" | sed "s,^go\s,,")"

  echo "Will install Go version from go.mod : ${VERSION}"
fi

# infer ARCH
ARCH="$(uname -m)"
if uname -m | grep -q x86_64; then ARCH=amd64; fi
if uname -m | grep -q aarch64; then ARCH=arm64; fi

# infer OS
OS="$(uname -s)"
if uname -s | grep -q Linux; then OS=linux; fi
if uname -s | grep -q Darwin; then OS=darwin; fi

fname="go${VERSION}.${OS}-${ARCH}.tar.gz"
[ -f "$fname" ] || wget "https://go.dev/dl/go${VERSION}.${OS}-${ARCH}.tar.gz"
sudo tar -C /usr/local -xvzf "${fname}"
rm "${fname}"

sudo ln -sf /usr/local/go/bin/go /usr/local/bin/go
sudo ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
