package pkg

import (
	. "github.com/saschagrunert/demo"
)

// MTU ISSUE
var runMTU = `

# https://blog.cloudflare.com/path-mtu-discovery-in-practice/

export KUBECONFIG=/home/kka/.kube/ztp5gc.yaml
tmux setenv KUBECONFIG /home/kka/.kube/ztp5gc.yaml

oc apply -f ./dayX/icmp-frag-case.yaml

# on frr-two
#show bgp vrf all summary
#show ip route vrf all

# tcpdump on all nodes
tmux new-window -n Nodes; tmux split-window -h -t Nodes; tmux split-window -h -t Nodes; tmux select-layout -t Nodes even-vertical; 
tmux send-keys -t Nodes.0 "oc debug node/m0.5gc.eric.vlab --image quay.io/karampok/snife:latest" C-m
tmux send-keys -t Nodes.1 "oc debug node/m1.5gc.eric.vlab --image quay.io/karampok/snife:latest" C-m
tmux send-keys -t Nodes.2 "oc debug node/m2.5gc.eric.vlab --image quay.io/karampok/snife:latest" C-m
tmux select-window -t Nodes; tmux setw synchronize-panes on
tmux send-keys -t Nodes.0 "chroot /host" C-m C-m C-m
tmux send-keys -t Nodes.0 "PS1=\$(hostname)\>' '" C-m C-m C-m
tmux send-keys -t Nodes.0 "podman run -it --privileged --net=host  quay.io/karampok/snife:latest tcpdump -nnn -i bond0.11 '(icmp and icmp[0] == 3 and icmp[1] == 4) or (host 203.100.100.100 and ip and ip[20+13] & tcp-syn != 0)'"

# tcpdump on all pods
kubectl tmux-exec -l app=bigfile -c server --select-layout=even-vertical -- /bin/bash
tmux rename-window Pods
tmux send-keys -t Pods "PS1=\$NODE_NAME\>' ';clear" C-m  # Prefix {
tmux send-keys -t Pods "tcpdump -i any -nn 'icmp and icmp[0] == 3 and icmp[1] == 4' or '(ip and ip[20+13] & tcp-syn != 0)' "

# ssh into green
oc get svc greensvc-icmp-mtu-local --output jsonpath='{.status.loadBalancer.ingress[0].ip}'
ssh lab1 podman exec -it green /bin/bash
cd /tmp/ && wget 5.8.6.1:9000/big.iso

# case 1 - no changes, it works
# 1. no ICMP/ip fragmentation same target node because X, why options [mss 1460] from client
#  ip link set mtu 1420 dev eth0
# 2. always traffic reaches Node X and pod on X handles the requestion (externalTrafficPolicy: Local in SVC)
# no cross node traffic (on geneve tunnels)
# 3. always same target node because ECMP (Equal-Cost-Mult-Path) routing on the router
# ip route show vrf green
# and the hash policy returns same value
# By default, Linux kernel uses the Layer 3 hash policy to load balance the IPv4 traffic. 
# Layer 3 hashing uses the following information: f(Source IP address,Destination IP address)= node X
# /proc/sys/net/ipv4/fib_multipath_hash_policy


# case 2: force smaller MTU in the path and fragmentation
ssh lab1 podman exec -it frr-two /bin/bash
ip link set dev green-eth mtu 1028
green: wget 5.8.6.9:9000/big.iso
# still works because route cache ?
ip route get 203.100.100.100
ip route flush cache
# and if clean cache ICMP frag always reaches node/pod due to L3 hash
# IP 11.11.11.254 > 10.129.0.132: ICMP 203.100.100.100 unreachable - need to frag (mtu 1028), length 556

# case 3: broken
# on the router we enforce L4 hashing
(Source IP address Destination IP address Source port number Destination port number Protocol)
echo 1 > /proc/sys/net/ipv4/fib_multipath_hash_policy
.green: wget 5.8.6.9:9000/big.iso

# case 4: workaround
# in all pods
cat /proc/sys/net/ipv4/tcp_mtu_probing
# 0
echo 1 > /proc/sys/net/ipv4/tcp_mtu_probing
`

func RunMTUDemo() *Run {
	r := NewRun("Run MTU frag issue")
	r.Step(S("Commands"), S(runMTU))
	return r
}
