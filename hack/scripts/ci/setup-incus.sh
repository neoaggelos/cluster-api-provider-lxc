#!/bin/bash -xeu

DIR="$(dirname "$(realpath "$0")")"

if ! which zpool; then
  sudo apt update
  sudo apt install zfsutils-linux --yes --no-install-recommends
fi

if ! which incus; then
  curl https://pkgs.zabbly.com/get/incus-stable | sudo bash -x
fi

# get node IP address
ip_address="$(ip -o route get to 1.1.1.1 | sed -n 's/.*src \([0-9.]\+\).*/\1/p')"

# Initialize Incus
sudo incus admin init --auto --network-address "$ip_address"

# NOTE(neoaggelos/2026-08-31): Incus local configuration directory has incorrect permissions
#   $ ls -la /home/runner/.config/incus:
#   total 16
#   drwxr-x--- 3 root   root   4096 Aug 31 14:47 .
#   drwxr-xr-x 5 runner runner 4096 Aug 31 14:47 ..
#   -rw-r--r-- 1 root   root    301 Aug 31 14:47 config.yml
#   drwxr-xr-x 2 root   root   4096 Aug 31 14:47 oidctokens
#
#   $ incus list
#   : Failed to load configuration: Unable to read the configuration file: open /home/runner/.config/incus/config.yml: permission denied
if [ "${GITHUB_ACTIONS:=}" == "true" ]; then
  sudo chown runner:runner ~/.config/incus -R
fi

# Generate client certificate and key, trust certificate
if ! incus remote switch local-https; then
  incus remote generate-certificate
  sudo incus config trust add-certificate ~/.config/incus/client.crt
  incus remote add local-https "https://$(sudo incus config get core.https_address)" --accept-certificate
  incus remote switch local-https
fi

# Do not drop instance traffic
sudo iptables -P FORWARD ACCEPT

# Write Kubernetes secret to top-level dir
echo "
apiVersion: v1
kind: Secret
metadata:
  name: lxc-secret
data:
  project: '$(echo -n default | base64 -w0)'
  server: '$(echo -n "https://$(incus config get core.https_address)" | base64 -w0)'
  server-crt: '$(cat ~/.config/incus/servercerts/local-https.crt | base64 -w0)'
  client-crt: '$(cat ~/.config/incus/client.crt | base64 -w0)'
  client-key: '$(cat ~/.config/incus/client.key | base64 -w0)'
" | tee "${DIR}/../../../lxc-secret.yaml"

# Setup local Incus daemon for e2e tests
CLI=incus "${DIR}/setup-e2e-resources.sh"

# NOTE(neoaggelos/2025-03-27): Incus KVMs are not starting in GitHub actions with ubuntu-24.04
# https://github.com/lxc/cluster-api-provider-incus/issues/44
if [ "${GITHUB_ACTIONS:=}" == "true" ]; then
  echo "** WARNING: Change default firmware, as KVMs do not work in GitHub actions out of the box"
  sudo apt update
  sudo apt install ovmf --no-install-recommends --yes
  sudo cp /usr/share/OVMF/OVMF_CODE_4M.fd /opt/incus/share/qemu/fixup-ovmf-code.fd
  incus profile set default raw.qemu.conf='[drive]
file = "/opt/incus/share/qemu/fixup-ovmf-code.fd"
'
  incus profile show default
fi
