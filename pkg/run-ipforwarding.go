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
var runBGP = `
#export KUBECONFIG=/home/kka/.kube/lab0.yaml
#tmux setenv KUBECONFIG /home/kka/.kube/lab0.yaml

tmux new-window -n Nodes; tmux split-window -h -t Nodes; tmux split-window -h -t Nodes; tmux select-layout -t Nodes even-vertical; 
docker exec -it clab-vlab-h00 /bin/bash
docker exec -it clab-vlab-r01 /bin/bash
sudo tcpdump -i sw1 -nnn host 10.10.0.10 -e
docker exec -it clab-vlab-r11 /bin/bash

sudo tcpdump -i sw1 -nnn port 5555 -eodocker exec -it clab-vlab-r11 /bin/bash
tmux send-keys -t Nodes.0 "podman-remote -c lab0  exec -it red-in /bin/bash" C-m
tmux send-keys -t Nodes.0 "ip route add 203.100.100.0/24 via 12.12.12.119" C-m
tmux send-keys -t Nodes.0 "ping -c 1 203.100.100.100"
tmux send-keys -t Nodes.0 "nc -u 5.5.5.5 8888 -p 2424"
tmux send-keys -t Nodes.1 "oc debug node/w0 --image quay.io/karampok/snife:latest" C-m
tmux send-keys -t Nodes.1 "chroot /host" C-m C-m C-m
tmux send-keys -t Nodes.1 "watch -d iptables -nvL FORWARD" C-m C-m C-m
tmux send-keys -t Nodes.2 "oc debug node/w0 --image quay.io/karampok/snife:latest" C-m
tmux send-keys -t Nodes.2 "mount -t debugfs none /sys/kernel/debug" C-m
tmux send-keys -t Nodes.2 "tcpdump -i any -nnn icmp" C-m
`

var runIPForwarding = `
# Observe the node that acts as router & keep metallb svc testing in a loop
export KUBECONFIG=/home/kka/.kube/lab0.yaml
tmux setenv KUBECONFIG /home/kka/.kube/lab0.yaml

tmux new-window -n Good; tmux split-window -h -t Goot; tmux split-window -h -t Good; tmux select-layout -t Good even-vertical; 
tmux send-keys -t Good.0 "podman-remote -c lab0 exec -it green /bin/bash" C-m
tmux send-keys -t Good.1 "oc debug node/w0 --image quay.io/karampok/snife:latest" C-m
tmux send-keys -t Good.1 "tcpdump -i any -nnn port 2311" C-m
tmux send-keys -t Good.0 "curl http://5.5.5.1:5555/hostname --local-port 2311" C-m

tmux new-window -n Nodes; tmux split-window -h -t Nodes; tmux split-window -h -t Nodes; tmux select-layout -t Nodes even-vertical; 
tmux send-keys -t Nodes.0 "podman-remote -c lab0  exec -it red-in /bin/bash" C-m
tmux send-keys -t Nodes.0 "ip route add 203.100.100.0/24 via 12.12.12.119" C-m
tmux send-keys -t Nodes.0 "ping -c 1 203.100.100.100"
tmux send-keys -t Nodes.0 "nc -u 5.5.5.5 8888 -p 2424"
tmux send-keys -t Nodes.1 "oc debug node/w0 --image quay.io/karampok/snife:latest" C-m
tmux send-keys -t Nodes.1 "chroot /host" C-m C-m C-m
tmux send-keys -t Nodes.1 "watch -d iptables -nvL FORWARD" C-m C-m C-m
tmux send-keys -t Nodes.2 "oc debug node/w0 --image quay.io/karampok/snife:latest" C-m
tmux send-keys -t Nodes.2 "mount -t debugfs none /sys/kernel/debug" C-m
tmux send-keys -t Nodes.2 "tcpdump -i any -nnn icmp" C-m
`

var cleanup06 = `podman stop green-in red-in`

func RunIPForwardingDemo() *Run {
	r := NewRun("Run runIPForwarding issue")
	r.Step(S("Green-in"), S(greenin))
	r.Step(S("Red-in"), S(redin))
	return r
}
