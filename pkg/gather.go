package pkg

import . "github.com/saschagrunert/demo"

var cleanup = []string{
	"ip link delete sw1",
	"ip link delete dataplane",
	"ip link delete ixp-net",
	"ip link delete bmc",
	"virsh net-destroy sw1",
	"virsh net-destroy dataplane",
}

func SetupInfra() *Run {
	r := NewRun("Setup Virtual Infra")
	r.Step(S("Build L2 fabric"), S(bridges))

	r.Step(S("Enable bridges in libvirt"), nil)
	r.Step(nil, S(cmd03))

	c := "containerlab deploy"
	r.Step(S("Containerlab"), S(c))
	cleanup = append(cleanup, "containerlab destroy")

	vbmh := `kcli create plan -f vbmh-kcli-plan.yaml vbmh`
	r.Step(S("Create baremetal with kcli"), S(vbmh))
	cleanup = append(cleanup, "kcli delete -y plan vbmh")
	//
	// r.BreakPoint()
	// r.Step(S("Setup green client on green net "), S(green))
	// r.Step(S("Setup red client on red net "), S(red))
	// r.Step(S("Setup macnet host on baremetal net "), S(macnet))
	// r.Step(S("Setup proxy"), S(proxy))
	// r.Step(S("Configure routing for proxy"), S(proxy01))
	// r.Step(S("Setup DHCPv4"), S(dhcpv4))
	// r.Step(S("Setup DHCPv6"), S(dhcpv6))
	//
	//
	// r.BreakPoint()
	// r.Step(S("Create bmc with Sushy"), S(sushy))
	// r.Step(S("Configure BMC networking (access ACM Hub) "), S(sushyNetconfig))
	//
	// r.BreakPoint()
	// r.Step(S("Setup kernel client on baremetal two interfaces"), S(kernel))
	// r.Step(S(helpkernel), nil)
	//
	// r.BreakPoint()
	// r.Step(S("Setup nmstate client on baremetal two interfaces"), S(nmstate))
	// r.Step(S(helpnmstate), nil)

	//r.BreakPoint()
	//r.Step(S("workstation:"), S("podman logs workstation"))

	return r
}
