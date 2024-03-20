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

// MTU ISSUE
var runIPForwarding = `
export KUBECONFIG=/home/kka/.kube/ztp5gc.yaml
green-in at green vlan with ip 11.11.11.50 and default route on 11.11.11.118, nc listen
red-in  at red vlan with ip 12.12.12.50 and default route on a node 12.12.12.118
`

func RunIPForwardingDemo() *Run {
	r := NewRun("Run runIPForwarding issue")
	r.Step(S("Green-in"), S(greenin))
	r.Step(S("Red-in"), S(redin))
	return r
}
