package pkg

import (
	. "github.com/saschagrunert/demo"
)

// var x = `
// tmux setenv KUBECONFIG /home/kka/.kube/lab0.yaml
// tmux setenv DOCKER_HOST tcp://10.1.104.10:2375
// `

func RunBGPGracefulRestart() *Run {
	r := NewRun("Run BPG graceful restart demo")

	c := `
kubectl apply -f day2/green-peering.yaml
kubectl apply -f day2/red-peering.yaml`
	r.Step(S("Setup peering"), S(c))

	c = `tmux new-window -n gateway
tmux send-keys -t gateway.0 "docker exec clab-vlab-gw1 vtysh -c 'show bgp vrf green summary'" C-m C-m C-m
tmux send-keys -t gateway.0 "docker exec clab-vlab-gw1 vtysh -c 'show bgp vrf red summary'" C-m C-m C-m
tmux send-keys -t gateway.0 "docker exec clab-vlab-gw1 vtysh -c 'show bfd peers brief'" C-m C-m C-m
tmux send-keys -t gateway.0 "docker exec clab-vlab-gw1 vtysh -c 'show bgp vrf green neighbors 11.11.11.103' 2>&1 >/tmp/green" C-m C-m C-m
tmux send-keys -t gateway.0 "docker exec clab-vlab-gw1 vtysh -c 'show bgp vrf red neighbors 12.12.12.103' 2>&1 >/tmp/red" C-m C-m C-m
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
	r.Step(S("Verify peering"), S(c))

	d := `
kubectl apply -f day2/green-pod-one.yaml
kubectl apply -f day2/red-pod-one.yaml`
	r.Step(S("Deploy workloads"), S(d))

	c = `kubectl get pods -o wide; kubectl get svc`
	r.Step(S("Verify workload"), S(c))

	c = `tmux new-window -n clients
tmux send-keys -t clients.0 "docker exec -it clab-vlab-gw1-sidecar /bin/bash" C-m C-m C-m
tmux send-keys -t clients.0 "ip vrf exec green /bin/bash" C-m C-m C-m
tmux send-keys -t clients.0 "curl -sf http://5.5.5.1:5555/hostname" C-m C-m C-m
tmux send-keys -t clients.0 "curl -sf \"http://5.5.5.1:5555/shell?cmd=\"env%7Cgrep%20-i%20node%0A\"\"" C-m C-m C-m
tmux send-keys -t clients.0 "while true;do curl -sf http://5.5.5.1:5555/hostname --connect-timeout 1  -o /dev/null || printf \"%s \" \$(date +%s) ;sleep 1;done" C-m
tmux split-window -v -t clients
tmux send-keys -t clients.1 "docker exec -it clab-vlab-gw1-sidecar /bin/bash" C-m C-m C-m
tmux send-keys -t clients.1 "ip vrf exec green /bin/bash" C-m C-m C-m
tmux send-keys -t clients.1 "curl -sf http://5.5.5.2:5555/hostname" C-m C-m C-m
tmux send-keys -t clients.1 "curl -sf \"http://5.5.5.2:5555/shell?cmd=\"env%7Cgrep%20-i%20node%0A\"\"" C-m C-m C-m
tmux send-keys -t clients.1 "while true;do curl -sf http://5.5.5.2:5555/hostname --connect-timeout 1  -o /dev/null || printf \"%s \" \$(date +%s) ;sleep 1;done" C-m
tmux split-window -v -t clients
tmux send-keys -t clients.2 "docker exec -it clab-vlab-gw1-sidecar /bin/bash" C-m C-m C-m
tmux send-keys -t clients.2 "ip vrf exec red /bin/bash" C-m C-m C-m
tmux send-keys -t clients.2 "curl -sf http://6.6.6.1:5555/hostname" C-m C-m C-m
tmux send-keys -t clients.2 "curl -sf \"http://6.6.6.1:5555/shell?cmd=\"env%7Cgrep%20-i%20node%0A\"\"" C-m C-m C-m
tmux send-keys -t clients.2 "while true;do curl -sf http://6.6.6.1:5555/hostname --connect-timeout 1  -o /dev/null || printf \"%s \" \$(date +%s) ;sleep 1;done" C-m
tmux split-window -v -t clients
tmux send-keys -t clients.3 "docker exec -it clab-vlab-gw1-sidecar /bin/bash" C-m C-m C-m
tmux send-keys -t clients.3 "ip vrf exec red /bin/bash" C-m C-m C-m
tmux send-keys -t clients.3 "curl -sf http://6.6.6.2:5555/hostname" C-m C-m C-m
tmux send-keys -t clients.3 "curl -sf \"http://6.6.6.2:5555/shell?cmd=\"env%7Cgrep%20-i%20node%0A\"\"" C-m C-m C-m
tmux send-keys -t clients.3 "while true;do curl -sf http://6.6.6.2:5555/hostname --connect-timeout 1  -o /dev/null || printf \"%s \" \$(date +%s) ;sleep 1;done" C-m
tmux select-layout -t clients even-vertical
tmux select-window -t clients; tmux setw synchronize-panes on
`
	r.Step(nil, S(c))

	return r
}
