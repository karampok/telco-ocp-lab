package pkg

import (
	. "github.com/saschagrunert/demo"
)

func Clean() *Run {
	r := NewRun("Clean")
	r.StepCanFail(S("Clean Workstation"), S(cleanup05))
	r.StepCanFail(S("Clean VMS"), S(cleanup04))
	r.StepCanFail(S("Clean clients"), S(cleanup03))
	r.StepCanFail(S("Clean SVC"), S(cleanup02))
	r.Step(S("Clean L3"), nil)
	for _, cmd := range cleanupL3 {
		r.StepCanFail(nil, S(cmd))
	}
	r.Step(S("Clean L2"), nil)
	for _, cmd := range cleanupL2 {
		r.StepCanFail(nil, S(cmd))
	}

	return r
}
