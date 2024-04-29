package pkg

import (
	. "github.com/saschagrunert/demo"
)

var runBGP = `
tmux new-window -n clients
tmux send-keys -t clients.0 "docker exec -it clab-vlab-sidecar-gw1 /bin/bash" C-m C-m C-m
tmux send-keys -t clients.0 "curl -sf http://4.4.4.1:5555/hostname" C-m C-m C-m
tmux send-keys -t clients.0 "while true;do curl -sf http://4.4.4.1:5555/hostname --connect-timeout 1  -o /dev/null || printf \"%s \" \$(date +%s) ;sleep 1;done" C-m

tmux split-window -v -t clients
tmux send-keys -t clients.1 "docker exec -it clab-vlab-sidecar-gw1 /bin/bash" C-m C-m C-m
tmux send-keys -t clients.1 "ip vrf exec green /bin/bash" C-m C-m C-m
tmux send-keys -t clients.1 "curl -sf http://5.5.5.1:5555/hostname" C-m C-m C-m
tmux send-keys -t clients.1 "while true;do curl -sf http://5.5.5.1:5555/hostname --connect-timeout 1  -o /dev/null || printf \"%s \" \$(date +%s) ;sleep 1;done" C-m

tmux split-window -v -t clients
tmux send-keys -t clients.2 "docker exec -it clab-vlab-sidecar-gw1 /bin/bash" C-m C-m C-m
tmux send-keys -t clients.2 "ip vrf exec red /bin/bash" C-m C-m C-m
tmux send-keys -t clients.2 "curl -sf http://6.6.6.1:5555/hostname" C-m C-m C-m
tmux send-keys -t clients.2 "while true;do curl -sf http://6.6.6.1:5555/hostname --connect-timeout 1  -o /dev/null || printf \"%s \" \$(date +%s) ;sleep 1;done" C-m
tmux select-layout -t clients even-vertical
tmux select-window -t clients; tmux setw synchronize-panes on

tmux new-window -n gateway
tmux send-keys -t gateway.0 "docker exec clab-vlab-gw1 vtysh -c \"show bgp vrf all summary\"" C-m C-m C-m
tmux send-keys -t gateway.0 "watch -d -c -n 1 docker exec clab-vlab-gw1 vtysh -c \\\"show ip bgp vrf all\\\""
tmux split-window -h -t gateway
tmux send-keys -t gateway.1 "docker exec clab-vlab-gw1 tail -f /tmp/frr.log" C-m
tmux split-window -v -t gateway
tmux send-keys -t gateway.2 "docker exec -it k00-worker /bin/bash" C-m
tmux send-keys -t gateway.2 "while true;do curl -sf http://127.0.0.1:7473/livez --connect-timeout 1  -o /dev/null || printf \"%s \" \$(date +%s) ;sleep 1;done" C-m
tmux split-window -v -t gateway
tmux send-keys -t gateway.3 "kubectl -n metallb-system logs -c reloader (kubectl -n metallb-system get pods -l component=speaker -o name) -f" C-m
tmux split-window -v -t gateway
tmux send-keys -t gateway.4 "kubectl set image daemonset/speaker frr=quay.io/frrouting/frr:9.1.0 -n metallb-system; kubectl -n metallb-system get pods -o wide -w"
`

var _ = `
exec -it clab-vlab-h00 /bin/bash" C-m C-m C-m
tmux split-window -h -t Nodes
 ip vrf exec green /bin/bash

 while true;do curl -sf http://6.6.6.1:5555/hostname --connect-timeout 1  -o /dev/null || printf "%s " $(date +%s) ;done

 while true;do curl -sf http://127.0.0.1:7473/livez --connect-timeout 1  -o /dev/null || printf "%3s " $(date +%s) ;done
 //
 //
  watch -d -c -n  1 docker exec -it clab-vlab-gw1 vtysh -c \"show ip bgp vrf all\"
#export KUBECONFIG=/home/kka/.kube/lab0.yaml
#tmux setenv KUBECONFIG /home/kka/.kube/lab0.yaml
tmux new-window -n Nodes
tmux send-keys -t Nodes.0 "docker exec -it clab-vlab-h00 /bin/bash" C-m C-m C-m
tmux split-window -h -t Nodes
tmux send-keys -t Nodes.1 "sudo tcpdump -i sw0 -nnn host 10.10.0.10 -e"
tmux split-window -h -t Nodes
tmux send-keys -t Nodes.2 "docker exec -it clab-vlab-r01 /bin/bash" C-m C-m C-m
tmux split-window -h -t Nodes
tmux send-keys -t Nodes.3 "docker exec -it clab-vlab-r11 /bin/bash" C-m C-m C-m
tmux split-window -h -t Nodes
tmux send-keys -t Nodes.4 "sudo tcpdump -i sw1 -nnn host 10.10.0.10 -e"
tmux split-window -h -t Nodes
tmux send-keys -t Nodes.5 "kubectl -n metallb-system exec -it -c frr (kubectl -n metallb-system get pods -l component=speaker -o name |tail -1) -- /bin/bash" C-m

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

func RunBGPGracefulRestart() *Run {
	r := NewRun("Run BPG graceful restart demo")
	// Infra is ready
	peers := `kubectl apply -f graceful/blue-peering.yaml
kubectl apply -f graceful/green-peering.yaml
kubectl apply -f graceful/red-peering.yaml`
	r.Step(S("Setup peering"), S(peers))

	c := `docker exec clab-vlab-gw1 vtysh -c "show bgp vrf all summary"`
	r.Step(S("Verify peering"), S(c))

	d := `kubectl apply -f graceful/blue-pod-one.yaml
kubectl apply -f graceful/green-pod-one.yaml
kubectl apply -f graceful/red-pod-one.yaml`
	r.Step(S("Deploy workloads"), S(d))

	c = `kubectl get pods -o wide; kubectl get svc`
	r.Step(S("Verify workload"), S(c))
	r.Step(nil, S(runBGP))
	return r
}
