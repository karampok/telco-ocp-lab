package pkg

import . "github.com/saschagrunert/demo"

func SetupInfra() *Run {
	r := NewRun("Setup Virtual Infra")
	r.BreakPoint()
	r.Step(S("Build L2 fabric"), nil)
	for _, cmd := range cmds01 {
		r.Step(nil, S(cmd))
	}

	r.BreakPoint()
	r.Step(S("Enable podman to attach containers"), nil)
	r.StepCanFail(nil, S(cmd02))

	r.BreakPoint()
	r.Step(S("Enable libvirt to attach vms"), nil)
	r.StepCanFail(nil, S(cmd03))

	r.BreakPoint()
	r.Step(S("Setup GW-zero (L3 Gateway) on access net"), S(gw0))
	r.Step(S("Configure GW-zero with upstream"), S(gw00))

	r.BreakPoint()
	r.Step(S("Setup workstation"), S(workstation))
	r.Step(S("Config workstation"), S(workstationConfig))
	r.Step(S("podman logs workstation"), nil)

	r.BreakPoint()
	r.Step(S("Setup GW-one (L3 Gateway) on baremetal,access net"), S(gw1))
	r.Step(S("Configure GW-one with vlan"), S(gw10))

	r.BreakPoint()
	r.Step(S("Setup GW-two (L3 Gateway) on baremetal,access,green net"), S(gw2))
	r.Step(S("Configure GW-two with vlan"), S(gw20))
	r.Step(S("Setup green VRF in router"), S(gw21))
	r.Step(S("Setup red VRF in router"), S(gw22))

	r.BreakPoint()
	r.Step(S("Setup green client on green net "), S(green))
	r.Step(S("Setup red client on red net "), S(red))
	r.Step(S("Setup macnet host on baremetal net "), S(macnet))

	r.BreakPoint()
	r.Step(S("Setup DNS (CoreDNS) service"), S(dns))
	r.Step(S("Configure routing for DNS"), S(dns01))

	r.BreakPoint()
	r.Step(S("Setup proxy"), S(proxy))
	r.Step(S("Configure routing for proxy"), S(proxy01))

	r.Step(S("proxy/dns needs connectivity"), nil)
	// not using my image, I can do
	// r.Step(S("podman run --net=container:dns --rm --privileged -it quay.io/karampok/snife /bin/bash"), nil)

	// r.BreakPoint()
	// r.Step(S("Setup DHCPv4"), S(dhcpv4))
	// r.Step(S("Setup DHCPv6"), S(dhcpv6))

	r.BreakPoint()
	r.Step(S("Create baremetal with kcli"), S(vbmh))

	r.BreakPoint()
	r.Step(S("Create bmc with Sushy"), S(sushy))
	r.Step(S("Configure BMC networking (access ACM Hub) "), S(sushyNetconfig))

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
