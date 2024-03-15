package pkg

var proxy = `podman run --net=access:interface_name=access --name proxy --rm -d --privileged \
-v ./opt/proxy/squid.conf:/etc/squid/squid.conf:Z --hostname proxy --dns-search telco.vlab --dns 10.10.20.10 \
--entrypoint='["/bin/bash", "-c", "ip addr add 10.10.20.20/24 dev access && ip link set dev access up; /sbin/entrypoint.sh"]' \
quay.io/karampok/squid:latest`

var proxy01 = `ns=$( podman inspect proxy | jq -r '.[0]["NetworkSettings"].SandboxKey')
 ip netns exec "${ns##*/}" ip route add 10.10.10.0/24 via 10.10.20.1
 ip netns exec "${ns##*/}" ip route add 10.10.10.0/24 via 10.10.20.2
 ip netns exec "${ns##*/}" ip route add default via 10.10.20.254
 ip netns exec "${ns##*/}" ping -c 3 8.8.8.8`

var dns = `podman run --net=access:interface_name=access --name dns --rm -d --read-only --privileged \
--hostname dns -v ./opt/coredns:/etc/coredns:Z \
--entrypoint='["/bin/bash", "-c", "ip addr add 10.10.20.10/24 dev access && ip link set dev access up; /usr/bin/coredns -conf /etc/coredns/Corefile"]' \
  quay.io/openshift/origin-coredns:latest `

var dns01 = `ns=$( podman inspect dns | jq -r '.[0]["NetworkSettings"].SandboxKey')
 ip netns exec "${ns##*/}" ip route add 10.10.10.0/24 via 10.10.20.1
 ip netns exec "${ns##*/}" ip route add 10.10.10.0/24 via 10.10.20.2 metric 100
 ip netns exec "${ns##*/}" ip route add default via 10.10.20.254
 ip netns exec "${ns##*/}" ping -c 3 8.8.8.8`

var dhcpv4 = `podman run --name dhcpv4 --rm -d --privileged --net=baremetal:interface_name=baremetal \
  --sysctl net.ipv6.conf.all.disable_ipv6=1 -v ./opt/dhcpv4/dnsmasq.conf:/etc/dnsmasq.conf:Z -v ./opt/dhcpv4/:/opt/dnsmasq/:Z \
  --entrypoint='["/bin/bash", "-c", "ip link add link baremetal name baremetal.10 type vlan id 10;ip link set baremetal.10 up; ip addr add 10.10.10.10/24 dev baremetal.10; dnsmasq -d"]' \
  quay.io/karampok/tools:dnsmasq`

var dhcpv6 = `podman run --name dhcpv6 --rm -d --privileged --net=baremetal:interface_name=baremetal \
  -v ./opt/dhcpv6/dnsmasq.conf:/etc/dnsmasq.conf:Z -v ./opt/dhcpv6/:/opt/dnsmasq/:Z \
  --entrypoint='["/bin/bash", "-c", "ip link add link baremetal name baremetal.10 type vlan id 10;ip link set baremetal.10 up; dnsmasq -d"]' \
  quay.io/karampok/tools:dnsmasq`

var cleanup02 = `
podman stop proxy || true
podman stop dns || true
podman stop dhcpv4 || true
podman stop dhcpv6 || true`
