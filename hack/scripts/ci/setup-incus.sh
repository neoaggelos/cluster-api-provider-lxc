#!/bin/bash -xeu

DIR="$(dirname "$(realpath "$0")")"

if ! which incus; then
  curl https://pkgs.zabbly.com/get/incus-stable | sudo bash -x
fi

# get node IP address
ip_address="$(ip -o route get to 1.1.1.1 | sed -n 's/.*src \([0-9.]\+\).*/\1/p')"

# Initialize LXD and configure clustering
sudo incus admin init --auto --network-address "$ip_address"
sudo incus network set incusbr0 ipv6.address=none
sudo incus cluster enable "$ip_address"

# Generate client certificate and key, trust certificate
if ! incus remote switch local-https; then
  incus remote generate-certificate
  sudo incus config trust add-certificate ~/.config/incus/client.crt
  incus remote add local-https "https://$(sudo incus config get core.https_address)" --accept-certificate
  incus remote switch local-https
fi

# Do not drop instance traffic
sudo iptables -P FORWARD ACCEPT
