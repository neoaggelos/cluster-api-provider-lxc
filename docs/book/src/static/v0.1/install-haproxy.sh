#!/bin/bash -xeu

# Usage:
#  $ /opt/cluster-api-lxc/50-install-haproxy.sh

set -xeu

apt update
apt install haproxy -y
