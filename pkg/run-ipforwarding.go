package pkg

import (
	. "github.com/saschagrunert/demo"
)

var redin = `
podman run --net=baremetal:interface_name=eth0 --name red-in --dns=none --rm -d --privileged \
quay.io/karampok/snife:latest
ns=$(podman inspect red-in | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip netns exec "${ns##*/}" ip link add link eth0 name eth0.12 type vlan id 12
ip netns exec "${ns##*/}" ip link set dev eth0.12 address CA:FE:C0:FF:EE:56
ip netns exec "${ns##*/}" ip addr add 12.12.12.150/24 dev eth0.12
ip netns exec "${ns##*/}" ip link set dev eth0.12 up
ip netns exec "${ns##*/}" ip route add 11.11.11.0/24 via 12.12.12.119
`

var greenin = `
podman run --net=baremetal:interface_name=eth0 --name green-in --dns=none --rm -d --privileged \
quay.io/karampok/snife:latest
ns=$(podman inspect green-in | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip netns exec "${ns##*/}" ip link add link eth0 name eth0.11 type vlan id 11
ip netns exec "${ns##*/}" ip link set dev eth0.11 address CA:FE:C0:FF:EE:55
ip netns exec "${ns##*/}" ip addr add 11.11.11.150/24 dev eth0.11
ip netns exec "${ns##*/}" ip link set dev eth0.11 up
ip netns exec "${ns##*/}" ip route add 12.12.12.0/24 via 11.11.11.119
`

var runIPForwarding = `
1. create connectivity from redin <-> greenin
2. observe the node that acts as router
- keep metallb svc testing in a loop
3. break with restricted forwarding, where is dropped
`

func RunIPForwardingDemo() *Run {
	r := NewRun("Run runIPForwarding issue")
	r.Step(S("Green-in"), S(greenin))
	r.Step(S("Red-in"), S(redin))
	return r
}
