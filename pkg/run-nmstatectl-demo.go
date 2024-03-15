package pkg

import (
	. "github.com/saschagrunert/demo"
)

var kernel = `podman run --net=baremetal:interface_name=eth1 --name kernel --dns=none --rm -d --privileged \
quay.io/karampok/snife:latest
ns=$(podman inspect kernel | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip link add eth2 netns "${ns##*/}" type veth peer name veth2brmt
ip link set dev veth2brmt master baremetal
ip link set dev veth2brmt up`

var helpkernel = `
# podman exec -it kernel /bin/bash
ip link add dev bond0 type bond
ip link set dev eth1 down
ip link set dev eth2 down
ip link set dev bond0 down
ip link set dev eth1 master bond0
ip link set dev eth2 master bond0
ip link set dev eth1 up
ip link set dev eth2 up
ip link set dev bond0 up
ip link add link bond0 name bond0.10 type vlan id 10
ip link set dev bond0.10 address CA:FE:C0:FF:EE:50

# ssh kvm2 tcpdump -i baremetal -U -s0 -w - ip6 or ether host CA:FE:C0:FF:EE:50 | wireshark -k -i -
ip link set dev bond0.10 up
# podman logs -f dhcpv4
dhclient -v bond0.10
# podman run -it --privileged --net=host  quay.io/karampok/snife:latest tcpdump -n -i any icmp
ping 8.8.8.8 -c 1
# podman logs -f dns
dig www.ntua.gr
# podman logs -f proxy
curl -kL -x "http://10.10.20.20:3128" https://ntua.gr
`

var nmstate = `podman run --net=baremetal:interface_name=eth1 --dns=none \
--name nmstate --rm -d --security-opt label=disable --cap-add=NET_ADMIN,SYS_ADMIN \
quay.io/karampok/nmstate:latest
ns=$(podman inspect nmstate | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip link add eth2 netns "${ns##*/}" type veth peer name veth2nmstate
ip link set dev veth2nmstate master baremetal
ip link set dev veth2nmstate up`

var helpnmstate = `
# podman exec -it nmstate /bin/bash
nmstatectl apply with-bond.yaml
nmstatectl apply with-vlans.yaml
nmstatectl apply with-vrf.yaml
ping 11.11.11.254
ip vrf exec green ping 11.11.11.254
ping 203.100.100.100
ip vrf exec green ping 203.100.100.100
ip route
ip route show vrf green
ip vrf exec green nc 203.100.100.100 3050
`

var green = `podman run --net=green-net:interface_name=eth0 --name green --dns=none --rm -d --privileged \
quay.io/karampok/snife:latest
ns=$(podman inspect green | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip netns exec "${ns##*/}" ip add add 203.100.100.100/24 dev eth0
ip netns exec "${ns##*/}" ip link set dev eth0 up
ip netns exec "${ns##*/}" ip route add default via 203.100.100.2`

var red = `podman run --net=red-net:interface_name=eth0 --name red --dns=none --rm -d --privileged \
quay.io/karampok/snife:latest
ns=$(podman inspect red | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip netns exec "${ns##*/}" ip add add 12.100.100.100/24 dev eth0
ip netns exec "${ns##*/}" ip link set dev eth0 up
ip netns exec "${ns##*/}" ip route add default via 12.100.100.2`

var macnet = `podman run --net=baremetal:interface_name=eth0 --name macnet --dns=none --rm -d --privileged \
quay.io/karampok/snife:latest
ns=$(podman inspect macnet | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip netns exec "${ns##*/}" ip link set dev eth0 up
ip netns exec "${ns##*/}" ip link add link eth0 name eth0.125 type vlan id 125
ip netns exec "${ns##*/}" ip link add link eth0 name eth0.126 type vlan id 126
ip netns exec "${ns##*/}" ip link add link eth0 name eth0.127 type vlan id 127
ip netns exec "${ns##*/}" ip link set dev eth0.125 up
ip netns exec "${ns##*/}" ip link set dev eth0.126 up
ip netns exec "${ns##*/}" ip link set dev eth0.127 up
ip netns exec "${ns##*/}" ip a a 172.100.125.1/24 dev eth0.125
ip netns exec "${ns##*/}" ip a a 172.100.126.1/24 dev eth0.126
ip netns exec "${ns##*/}" ip a a 172.100.127.1/24 dev eth0.127`

var cleanup03 = `podman stop kernel
podman stop nmstate
podman stop green
podman stop red
podman stop macnet`

func RunNMSTATEDemo() *Run {
	r := NewRun("Setup client")
	r.BreakPoint()
	r.Step(S("Setup kernel client on baremetal two interfaces"), S(kernel))
	r.Step(S(helpkernel), nil)

	r.BreakPoint()
	r.Step(S("Setup nmstate client on baremetal two interfaces"), S(nmstate))
	r.Step(S(helpnmstate), nil)

	return r
}
