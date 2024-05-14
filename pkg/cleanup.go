package pkg

import (
	. "github.com/saschagrunert/demo" //nolint:stylecheck // dot imports are intended here
)

func Clean() *Run {
	r := NewRun("Clean")
	for _, cmd := range cleanup {
		r.StepCanFail(nil, S(cmd))
	}

	return r
}
