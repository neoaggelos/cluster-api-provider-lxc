#!/bin/bash -xeu

DIR="$(dirname "$(realpath "$0")")"

if ! snap list lxd; then
  sudo apt update
  sudo apt install snapd -y
  sudo snap install lxd --channel 5.21/stable
fi

sudo snap refresh lxd --channel 5.21/stable

# get node IP address
ip_address="$(ip -o route get to 1.1.1.1 | sed -n 's/.*src \([0-9.]\+\).*/\1/p')"

# Initialize LXD and configure clustering
sudo lxd init --auto --network-address "$ip_address"
sudo lxc network set lxdbr0 ipv6.address=none
sudo lxc cluster enable "$ip_address"

# Generate client certificate and key, trust certificate
if ! lxc remote switch local-https; then
  token="$(sudo lxc config trust add --name client | tail -1)"
  lxc remote add local-https "https://$(sudo lxc config get core.https_address)" --accept-certificate --token "$token"
  lxc remote set-default local-https
fi

# Do not drop instance traffic
sudo iptables -P FORWARD ACCEPT
