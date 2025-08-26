package pkg

import (
	. "github.com/saschagrunert/demo"
)

// var x = `
// tmux setenv KUBECONFIG ~/.kube/lab0.yaml
// tmux setenv DOCKER_HOST tcp://10.1.104.10:2375
// `

func RunBGPUnnumber() *Run {
	r := NewRun("Run BPG unnumbered peering demo")

	// 	c := `kubectl apply -f day2/red-nmstate.yaml`
	// 	r.Step(S("Setup unnumbered interface/p2p on the worker"), S(c))
	//
	// 	c = `tmux new-window -n unnumber
	// 	tmux send-keys -t unnumber.0 "function fish_prompt; echo '(external peer) #2> ';  end; clear" C-m C-m
	// 	tmux send-keys -t unnumber.0 "docker exec clab-vlab-gw1 vtysh -c 'show running-config bgpd'" C-m C-m
	// 	tmux send-keys -t unnumber.0 "docker exec clab-vlab-gw1 /bin/bash -c 'ip a s eth1.red'" C-m C-m
	// 	tmux split-window -v -t unnumber
	// 	tmux send-keys -t unnumber.1 "function fish_prompt; echo '(switch) #2> ';  end; clear" C-m C-m
	// 	tmux send-keys -t unnumber.1 "ssh lab0 -- tcpdump -i sw1 -nn -l -c 5 -e 'vlan 12' " C-m
	// 	tmux split-window -v -t unnumber
	// 	tmux send-keys -t unnumber.2 "function fish_prompt; echo '(internal peer - w1) #2> ';  end; clear" C-m C-m
	// 	tmux send-keys -t unnumber.2 "oc debug node/w1  --image quay.io/karampok/snife:latest -- ip a s bond0.12 " C-m C-m
	// 	tmux select-layout -t unnumber even-vertical
	// 	 `
	// 	r.Step(S("Verify unnumber interfaces in p2p"), S(c))
	//
	// 	c = `
	// tmux send-keys -t unnumber.0 "ip link add dummy0 type dummy; ip link set dev dummy0 master red; ip link set dev dummy0 up" C-m C-m
	// sleep 5
	// tmux send-keys -t unnumber.0 "ip addr add 100.100.100.150/24 dev dummy0" C-m C-m
	// tmux send-keys -t unnumber.0 "ip addr show dummy0" C-m C-m
	// tmux send-keys -t unnumber.2 "clear" C-m
	// 	`
	// 	r.Step(S("Fix source IP"), S(c))
	//
	c := `kubectl apply -f day2/red-peering.yaml`
	r.Step(S("Setup peering"), S(c))

	//oc -n openshift-frr-k8s label pod frr-k8s-ntgqt node=w1
	c = `
	#	tmux send-keys -t unnumber.0 "docker exec clab-vlab-gw1 vtysh -c 'show bgp vrf red summary'" C-m C-m
	#tmux send-keys -t unnumber.2 "oc -n openshift-frr-k8s exec -it -c frr (oc -n openshift-frr-k8s  get pods -l node=w1 -o name) -- /bin/bash" C-m
	#tmux send-keys -t unnumber.2 "vtysh -c 'show bgp summary'" C-m
	#tmux send-keys -t unnumber.2 "vtysh -c 'show bgp neighbor'" C-m
	 `
	r.Step(S("Verify peering"), S(c))

	// c = `
	// kubectl apply -f day2/red-deploy.yaml
	// sleep 5
	// kubectl get pods -o wide; kubectl get svc
	// tmux send-keys -t unnumber.0 "docker exec clab-vlab-gw1 vtysh -c 'show bgp vrf red summary'" C-m C-m
	//  `
	// r.Step(S("Deploy workloads"), S(c))

	// 	c = `
	// tmux send-keys -t unnumber.1 "clear" C-m
	// tmux send-keys -t unnumber.1 "ssh lab0 -- tcpdump -i sw1 -nn -l -c 15 -e 'vlan 12' and host 6.6.6.1" C-m
	// sleep 5
	// tmux send-keys -t unnumber.0 'docker exec -it clab-vlab-gw1-sidecar /bin/bash' C-m C-m
	// tmux send-keys -t unnumber.0 "ip vrf exec red curl --connect-timeout 1 http://6.6.6.1:5555/hostname" C-m C-m
	// sleep 5
	// 	`
	// 	r.Step(S("Observe 0.0.0.0 as source IP"), S(c))

	r.BreakPoint()
	c = `
kubectl apply -f day2/red-frrconfig-allow-prx-in.yaml
# external should network 100.100.100.0/24
sleep 5s
tmux send-keys -t unnumber.2 "ip route get 100.100.100.150" C-m`
	r.Step(S("Fix routing on OCP node - bgp learning only, no static"), S(c))

	c = `
	#tmux send-keys -t unnumber.0 "clear; ip vrf exec red curl --connect-timeout 1 http://6.6.6.1:5555/hostname" C-m C-m
tmux send-keys -t unnumber.2 "ip route get 100.100.100.150" C-m
	`
	r.Step(S("Observe e2e traffic working"), S(c))

	return r
}

// while true
//     curl -sf http://4.4.4.1:4444/hostname --connect-timeout 1 -o /dev/null; or printf "%s " (date +%s)
//     sleep 1
