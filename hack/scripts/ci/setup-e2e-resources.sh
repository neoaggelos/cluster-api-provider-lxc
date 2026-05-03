#!/bin/bash -xeu

########################################################################
### Usage:
###   $ CLI=incus ./setup-e2e.sh
###   $ CLI=lxc ./setup-e2e.sh

# instance profiles
LXC_PROFILE_NAME=default

# zfs storage pool
LXC_STORAGE_POOL_NAME=zfs
LXC_STORAGE_POOL_SIZE=20GiB

# local bridge network (10.200.1.0/24)
LXC_NETWORK_NAME="testbr0"
LXC_NETWORK_IPV6="none"
LXC_NETWORK_IPV4="10.200.1.1/24"
LXC_NETWORK_IPV4_DHCP="10.200.1.10-10.200.1.100"
LXC_NETWORK_IPV4_OVN="10.200.1.101-10.200.1.150"
LXC_NETWORK_IPV4_VIP_RANGES="10.200.1.151-10.200.1.160"

# local OVN network (192.168.200.0/24)
LXC_OVN_NETWORK_NAME="testovn0"
LXC_OVN_NETWORK_IPV6="none"
LXC_OVN_NETWORK_IPV4="192.168.200.1/24"
LXC_OVN_NETWORK_IPV4_VIP_RANGES="10.200.1.161-10.200.1.170"

########################################################################

# configure storage
if ! "${CLI}" storage show "${LXC_STORAGE_POOL_NAME}" 2> /dev/null; then
  if ! "${CLI}" storage create "${LXC_STORAGE_POOL_NAME}" zfs size="${LXC_STORAGE_POOL_SIZE}"; then
    echo "Failed to create ZFS storage pool, will use default storage"
  else
    "${CLI}" profile device set "${LXC_PROFILE_NAME}" root type=disk pool="${LXC_STORAGE_POOL_NAME}"
  fi
fi

########################################################################

# configure default network. the user.capn.vip.ranges annotation is used by "QuickStart KubeVIP"
if ! "${CLI}" network show "${LXC_NETWORK_NAME}" 2> /dev/null; then
  "${CLI}" network create "${LXC_NETWORK_NAME}" --type=bridge \
    ipv4.address="${LXC_NETWORK_IPV4}" ipv4.nat=true \
    ipv6.address="${LXC_NETWORK_IPV6}" ipv6.nat=true \
    ipv4.dhcp.ranges="${LXC_NETWORK_IPV4_DHCP}" \
    ipv4.ovn.ranges="${LXC_NETWORK_IPV4_OVN}" \
    user.capn.vip.ranges="${LXC_NETWORK_IPV4_VIP_RANGES}" \
    user.capn.e2e.kube-vip=true
fi

# configure ovn network. the user.capn.vip.ranges annotation is used by "QuickStart OVN"
if ! "${CLI}" network show "${LXC_OVN_NETWORK_NAME}" 2> /dev/null; then
  "${CLI}" network create "${LXC_OVN_NETWORK_NAME}" --type=ovn \
    network="${LXC_NETWORK_NAME}" \
    ipv4.address="${LXC_OVN_NETWORK_IPV4}" ipv4.nat=true \
    ipv6.address="${LXC_OVN_NETWORK_IPV6}" ipv6.nat=true \
    user.capn.vip.ranges="${LXC_OVN_NETWORK_IPV4_VIP_RANGES}" \
    user.capn.e2e.ovn=true \
  || echo "Failed to create OVN network, will skip OVN tests"
fi

# configure default profile
"${CLI}" profile device set "${LXC_PROFILE_NAME}" eth0 type=nic network="${LXC_NETWORK_NAME}"
"${CLI}" profile show default
