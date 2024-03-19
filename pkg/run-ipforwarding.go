package pkg

import (
	. "github.com/saschagrunert/demo"
)

// MTU ISSUE
var runIPForwarding = `
export KUBECONFIG=/home/kka/.kube/ztp5gc.yaml
green-in at green vlan with ip 11.11.11.50 and default route on 11.11.11.118, nc listen
red-in  at red vlan with ip 12.12.12.50 and default route on a node 12.12.12.118
`

func RunIPForwardingDemo() *Run {
	r := NewRun("Run runIPForwarding issue")
	r.Step(S("Commands"), S(runIPForwarding))
	return r
}
