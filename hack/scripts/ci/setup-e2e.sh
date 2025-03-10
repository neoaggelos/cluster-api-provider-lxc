#!/bin/bash -xeu

########################################################################
### Usage:
###   $ CLI=incus ./common-e2e.sh
###   $ CLI=lxc ./common-e2e.sh

# instance profiles
LXC_PROFILE_NAME=default

# local bridge network (10.200.1.0/24)
LXC_NETWORK_NAME="testbr0"
LXC_NETWORK_IPV6="none"
LXC_NETWORK_IPV4="10.200.1.1/24"
LXC_NETWORK_IPV4_DHCP="10.200.1.10-10.200.1.100"
LXC_NETWORK_IPV4_OVN="10.200.1.101-10.200.1.150"
LXC_NETWORK_IPV4_KUBE_VIP="10.200.1.151"

# local OVN network (192.168.200.0/24)
LXC_OVN_NETWORK_NAME="testovn0"
LXC_OVN_NETWORK_IPV6="none"
LXC_OVN_NETWORK_IPV4="192.168.200.1/24"
LXC_OVN_NETWORK_IPV4_DHCP="192.168.200.10-192.168.200.100"
LXC_OVN_NETWORK_IPV4_LB="10.200.1.201"

########################################################################

# install and configure OVN
sudo apt install ovn-host ovn-central -y
sudo ovs-vsctl set open_vswitch . external_ids:ovn-encap-ip="127.0.0.1"
sudo ovs-vsctl set open_vswitch . external_ids:ovn-encap-type="geneve"
sudo ovs-vsctl set open_vswitch . external_ids:ovn-remote="unix:/run/ovn/ovnsb_db.sock"

########################################################################

# configure default network. the user.capl.e2e.kube-vip-address annotation is used by "QuickStart KubeVIP"
if ! "${CLI}" network show "${LXC_NETWORK_NAME}" 2> /dev/null; then
  "${CLI}" network create "${LXC_NETWORK_NAME}" --type=bridge \
    ipv4.address="${LXC_NETWORK_IPV4}" ipv4.nat=true \
    ipv6.address="${LXC_NETWORK_IPV6}" ipv6.nat=true \
    ipv4.dhcp.ranges="${LXC_NETWORK_IPV4_DHCP}" \
    ipv4.ovn.ranges="${LXC_NETWORK_IPV4_OVN}" \
    user.capl.e2e.kube-vip-address="${LXC_NETWORK_IPV4_KUBE_VIP}"
fi

# configure ovn network. the user.capl.e2e.ovn-lb-address annotation is used by "QuickStart OVN"
if ! "${CLI}" network show "${LXC_OVN_NETWORK_NAME}" 2> /dev/null; then
  "${CLI}" network create "${LXC_OVN_NETWORK_NAME}" --type=ovn \
    network="${LXC_NETWORK_NAME}" \
    ipv4.address="${LXC_OVN_NETWORK_IPV4}" ipv4.nat=true \
    ipv6.address="${LXC_OVN_NETWORK_IPV6}" ipv6.nat=true \
    ipv4.dhcp.ranges="${LXC_OVN_NETWORK_IPV4_DHCP}" \
    user.capl.e2e.ovn-lb-address="${LXC_OVN_NETWORK_IPV4_LB}"
fi

# configure default profile
echo "
devices:
  eth0:
    type: nic
    network: ${LXC_NETWORK_NAME}
  root:
    type: disk
    path: /
    pool: local
" | "${CLI}" profile edit "${LXC_PROFILE_NAME}"
