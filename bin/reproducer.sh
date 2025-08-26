#!/bin/sh
 
set -x
 
killall -9 tcpdump
killall -9  radvd

# cat radvd.conf
# interface vetha
# {
#    AdvSendAdvert on;
#    MaxRtrAdvInterval 10;
#    MinDelayBetweenRAs  20;
#
# };
echo 1 > /proc/sys/net/ipv6/conf/all/router_solicitations
echo 1 > /proc/sys/net/ipv6/conf/default/router_solicitations

for ns in a b; do
        ip netns del $ns
        ip netns add $ns
        ip -n $ns link set dev lo up
        ip link add name veth$ns type veth peer netns $ns name veth$ns
        ip -n $ns link set dev veth$ns address aa:aa:aa:aa:a$ns:10
        ip -n $ns link set dev veth$ns up
        ip link set dev veth$ns up
done
 
ip link del br0
ip link add name br0 type bridge
ip -6 addr add dev br0 2001::1/64
#sysctl -w net.ipv6.conf.br0.disable_ipv6=1
echo 0 > /proc/sys/net/ipv6/conf/br0/accept_ra
ip link set dev br0 up

ip link set dev vetha master br0
ip link set dev vethb master br0

# docker run --privileged --rm -t --pid=host -v /sys/kernel/debug/:/sys/kernel/debug/ cilium/pwru pwru --output-tuple 'ether host aa:aa:aa:aa:aa:10 and ip6[40] = 134'
 
# tcpdump -nei vetha -vvv icmp6 and ip6[40] == 134  # yes
# tcpdump -nei br0 -vvv icmp6 and ip6[40] == 134 # yes
# tcpdump -nei vethb -vvv icmp6 and ip6[40] == 134 # no

# ip -6 -n a addr add dev veth0 2001::1/64 # this spefically should not happen
# my use case in only link-local ipv6 adresses

sleep 0.1
ip netns exec a radvd -n -d 3 -m stderr -C radvd.conf
 
