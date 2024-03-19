package pkg

import (
	. "github.com/saschagrunert/demo"
)

var gw0 = `podman run --name frr-zero --rm -d --hostname frr-zero --privileged \
--sysctl net.ipv6.conf.all.forwarding=1 --sysctl net.ipv4.ip_forward=1 \
-v /lib/modules:/lib/modules \
--net=access:interface_name=access --net=bmc:interface_name=bmc \
-v ./opt/frr-zero:/etc/frr:Z quay.io/frrouting/frr:8.5.1`

var gw00 = `ns=$(podman inspect frr-zero | jq -r '.[0]["NetworkSettings"].SandboxKey')
 ip link add upstream netns "${ns##*/}" type veth peer name frr-zero ; ip a s frr-zero
 ip addr add 169.254.100.9/30 dev frr-zero
 ip link set dev frr-zero up
 ip netns exec "${ns##*/}" ip addr add 192.168.100.254/24 dev bmc
 ip netns exec "${ns##*/}" ip link set dev bmc up
 ip netns exec "${ns##*/}" ip addr add 169.254.100.10/30 dev upstream
 ip netns exec "${ns##*/}" ip link set dev upstream up
 ip netns exec "${ns##*/}" ip addr add 10.10.20.254/24 dev access
 ip netns exec "${ns##*/}" ip link set dev access up
 ip netns exec "${ns##*/}" ping -c3 169.254.100.9
 ip netns exec "${ns##*/}" ip route add default via 169.254.100.9 src 10.10.20.254
 ip netns exec "${ns##*/}" ip route add 10.10.10.0/24 via 10.10.20.1
 ip netns exec "${ns##*/}" ip route add 10.10.10.0/24 via 10.10.20.2
 ip route add 10.10.20.0/24 via 169.254.100.10 #route access net back
 ip route add 10.10.10.0/24 via 169.254.100.10 #route baremetal net back
dev=$(ip -json route get 8.8.8.8 |jq -r .[0].dev)
iptables -t nat -I POSTROUTING -o $dev --source 10.10.10.0/24 -j MASQUERADE
iptables -t nat -I POSTROUTING -o $dev --source 10.10.20.0/24 -j MASQUERADE
ip netns exec "${ns##*/}" ping -c3 8.8.8.8 #check firewalld`

var gw1 = `podman run --name frr-one --rm -d --hostname frr-one --privileged \
  --sysctl net.ipv6.conf.all.forwarding=1 --sysctl net.ipv4.ip_forward=1 \
  --net=baremetal:mac=aa:aa:aa:aa:aa:10,interface_name=baremetal \
  --net=access:mac=aa:aa:aa:aa:aa:11,interface_name=access \
  -v ./opt/frr-one:/etc/frr:Z quay.io/frrouting/frr:8.5.1`

var gw10 = `ns=$(podman inspect frr-one | jq -r '.[0]["NetworkSettings"].SandboxKey')
 ip netns exec "${ns##*/}" ip link add link baremetal name baremetal.10 type vlan id 10
 ip netns exec "${ns##*/}" ip link set dev baremetal.10 up
 ip netns exec "${ns##*/}" ip addr add 10.10.20.1/24 dev access
 ip netns exec "${ns##*/}" ip route add default via 10.10.20.254
 ip netns exec "${ns##*/}" ping -c3 8.8.8.8`

var gw2 = `podman run --name frr-two --rm -d --hostname frr-two --privileged \
--sysctl net.ipv6.conf.all.forwarding=1 --sysctl net.ipv4.ip_forward=1 \
--net=baremetal:mac=aa:aa:aa:aa:aa:20,interface_name=baremetal \
--net=access:mac=aa:aa:aa:aa:aa:22,interface_name=access \
--net=green-net:interface_name=green-eth \
--net=red-net:interface_name=red-eth \
-v ./opt/frr-two:/etc/frr:Z quay.io/frrouting/frr:9.0.2`

var gw20 = `ns=$(podman inspect frr-two | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip netns exec "${ns##*/}" ip link add link baremetal name baremetal.10 type vlan id 10
ip netns exec "${ns##*/}" ip link set dev baremetal.10 up
ip netns exec "${ns##*/}" ip addr add 10.10.20.2/24 dev access
ip netns exec "${ns##*/}" ip route add default via 10.10.20.254
ip netns exec "${ns##*/}" ping -c3 8.8.8.8`

var gw21 = `ns=$(podman inspect frr-two | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip netns exec "${ns##*/}" ip link add green type vrf table 110
ip netns exec "${ns##*/}" ip link add link baremetal name baremetal.11 type vlan id 11
ip netns exec "${ns##*/}" ip link set baremetal.11 master green
ip netns exec "${ns##*/}" ip link set green-eth master green
ip netns exec "${ns##*/}" ip add add 203.100.100.2/24 dev green-eth
ip netns exec "${ns##*/}" ip add add 11.11.11.254/24 dev baremetal.11
ip netns exec "${ns##*/}" ip link set dev green up #this is the VRF
ip netns exec "${ns##*/}" ip link set dev baremetal.11 up
ip netns exec "${ns##*/}" ip link set dev green-eth up`

var gw22 = `ns=$(podman inspect frr-two | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip netns exec "${ns##*/}" ip link add red type vrf table 120
ip netns exec "${ns##*/}" ip link add link baremetal name baremetal.12 type vlan id 12
ip netns exec "${ns##*/}" ip link set baremetal.12 master red
ip netns exec "${ns##*/}" ip link set red-eth master red
ip netns exec "${ns##*/}" ip add add 12.100.100.2/24 dev red-eth
ip netns exec "${ns##*/}" ip add add 12.12.12.254/24 dev baremetal.12
ip netns exec "${ns##*/}" ip link set dev red up #this is the VRF
ip netns exec "${ns##*/}" ip link set dev baremetal.12 up
ip netns exec "${ns##*/}" ip link set dev red-eth up`

var cleanupL3 = []string{
	"podman stop frr-zero",
	"podman stop frr-one",
	"podman stop frr-two",
	"podman stop green",
	"podman stop red",
	"ip route delete 10.10.20.0/24 via 169.254.100.10 #route access net back",
	"ip route delete 10.10.10.0/24 via 169.254.100.10 #route baremetal net back",
	"export dev=$(ip -json route get 8.8.8.8 |jq -r .[0].dev) && iptables -t nat -D POSTROUTING -o $dev --source 10.10.20.0/24 -j MASQUERADE",
	"export dev=$(ip -json route get 8.8.8.8 |jq -r .[0].dev) && iptables -t nat -D POSTROUTING -o $dev --source 10.10.10.0/24 -j MASQUERADE",
}

func SetupIPv6WithSLAAC() *Run {
	r := NewRun("Setup IPv6 SLAAC")
	return r
}

func SetupIPv6WithDHCPv6() *Run {
	r := NewRun("Setup IPv6 with DHCPv6")
	return r
}
