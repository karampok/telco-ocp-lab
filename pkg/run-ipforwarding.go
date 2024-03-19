package pkg

import (
	. "github.com/saschagrunert/demo"
)

// MTU ISSUE
var runIPForwarding = `
export KUBECONFIG=/home/kka/.kube/ztp5gc.yaml
`

func RunIPForwardingDemo() *Run {
	r := NewRun("Run runIPForwarding issue")
	r.Step(S("Commands"), S(runIPForwarding))
	return r
}
