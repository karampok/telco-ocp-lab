package pkg

import (
	. "github.com/saschagrunert/demo"
)

func RunBGPGracefulRestart() *Run {
	r := NewRun("Run BPG graceful restart demo")

	c := `kubectl apply -f graceful/blue-peering.yaml
	 kubectl apply -f graceful/green-peering.yaml
	 kubectl apply -f graceful/red-peering.yaml`
	r.Step(S("Setup peering"), S(c))

	d := `kubectl apply -f graceful/blue-pod-one.yaml
	 kubectl apply -f graceful/green-pod-one.yaml
	 kubectl apply -f graceful/red-pod-one.yaml`
	r.Step(S("Deploy workloads"), S(d))

	c = `kubectl get pods -o wide; kubectl get svc`
	r.Step(S("Verify workload"), S(c))

	c = `tmux new-window -n clients
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
`
	r.Step(nil, S(c))

	c = `tmux new-window -n gateway
tmux send-keys -t gateway.0 "docker exec clab-vlab-gw1 vtysh -c 'show bgp vrf all summary'" C-m C-m C-m
tmux send-keys -t gateway.0 "watch -d -c -n 1 docker exec clab-vlab-gw1 vtysh -c \\\"show ip bgp vrf all\\\""
# tmux send-keys -t gateway.0 "bash" C-m C-m
# tmux send-keys -t gateway.0 "while true; do docker exec clab-vlab-gw1 vtysh -c 'show ip bgp vrf all'|grep 4.4.4.1 ; echo \"== \$(date --utc)\" ; sleep 1; done" C-m
tmux split-window -h -t gateway
tmux send-keys -t gateway.1 "docker exec clab-vlab-gw1 tail -f /tmp/frr.log" C-m
tmux split-window -v -t gateway
tmux send-keys -t gateway.2 "kubectl -n metallb-system logs -c reloader (kubectl -n metallb-system get pods -l component=speaker -o name) -f" C-m
tmux split-window -v -t gateway
tmux send-keys -t gateway.3 "kubectl set image daemonset/speaker frr=quay.io/frrouting/frr:9.1.0 -n metallb-system; kubectl -n metallb-system get pods -o wide -w"
`

	r.Step(nil, S(c))
	return r
}

func RunBGPGracefulRestartWithBFD() *Run {
	r := NewRun("Run BPG graceful restart with BFD demo")

	c := `kubectl apply -f graceful/red-peering.yaml`
	r.Step(S("Setup peering"), S(c))

	d := `kubectl apply -f graceful/red-pod-two.yaml`
	r.Step(S("Deploy workloads"), S(d))

	c = `tmux new-window -n clients
tmux send-keys -t clients.0 "docker exec -it clab-vlab-sidecar-gw1 /bin/bash" C-m C-m C-m
tmux send-keys -t clients.0 "ip vrf exec red /bin/bash" C-m C-m C-m
tmux send-keys -t clients.0 "curl -sf http://6.6.6.1:5555/hostname" C-m C-m C-m
tmux send-keys -t clients.0 "while true;do curl -sf http://6.6.6.1:5555/hostname --connect-timeout 1 || printf \"%s \" \$(date +%s) ;sleep 1;echo; done" C-m
tmux split-window -v -t clients
tmux send-keys -t clients.1 "kubectl get pods -o wide" C-m C-m C-m
tmux split-window -v -t clients
tmux send-keys -t client.2 "watch -d -c -n 1 docker exec clab-vlab-gw1 vtysh -c \\\"show ip bgp vrf red\\\"" C-m
tmux split-window -v -t clients
tmux send-keys -t client.3 "# kubectl set image daemonset/speaker frr=quay.io/frrouting/frr:9.1.0 -n metallb-system; kubectl -n metallb-system get pods -o wide -w" C-m
tmux send-keys -t client.3 "docker exec -it k00-worker bash -c \"ip link set dev red down\""
tmux select-layout -t clients even-vertical
`
	r.Step(nil, S(c))

	c = `tmux new-window -n gateway
tmux send-keys -t gateway.0 "docker exec clab-vlab-gw1 tail -f /tmp/frr.log" C-m
tmux split-window -v -t gateway
tmux send-keys -t gateway.1 "sudo tcpdump -i sw1 -nnn tcp port 179 -w - | wireshark -k -i -"
`
	r.Step(nil, S(c))

	return r
}
