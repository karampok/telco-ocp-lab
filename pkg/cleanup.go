package pkg

import (
	. "github.com/saschagrunert/demo" //nolint:stylecheck // dot imports are intended here
)

func Clean() *Run {
	r := NewRun("Clean")
	//r.StepCanFail(S("Clean VMS"), S(cleanup04))
	r.Step(S("Clean L2"), nil)
	for _, cmd := range cleanupL2 {
		r.StepCanFail(nil, S(cmd))
	}

	return r
}
