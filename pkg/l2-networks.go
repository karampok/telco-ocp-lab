package pkg

func sudo(s string) string {
	return "sudo " + s
}

func orTrue(s string) string {
	return s + " || true"
}

var cmds01 = []string{
	`ip link add name access type bridge
ip link add name baremetal type bridge
ip link add name green-net type bridge
ip link add name red-net type bridge
ip link add name bmc type bridge
ip link add name dataplane type bridge
ip link set mtu 9000 dev dataplane
ip link set mtu 9000 dev baremetal
ip link set dev baremetal up
ip link set dev access up
ip link set dev green-net up
ip link set dev red-net up
ip link set dev bmc up
ip link set dev dataplane up`,
}

var cmd02 = `mkdir -p /etc/cni/net.d
cp ./opt/cni.d/{access,baremetal,green-net,red-net,bmc}.conflist /etc/cni/net.d/
# podman network ls (minimal CNI, no ipam, gateway or anything)`

var cmd03 = `cat > /tmp/baremetal.xml <<EOM
<network>
  <name>baremetal</name>
  <forward mode="bridge"/>
  <bridge name="baremetal"/>
</network>
EOM
virsh net-create /tmp/baremetal.xml
cat > /tmp/bmc.xml <<EOM
<network>
  <name>bmc</name>
  <forward mode="bridge"/>
  <bridge name="bmc"/>
</network>
EOM
virsh net-create /tmp/bmc.xml
cat > /tmp/dataplane.xml <<EOM
<network>
  <name>dataplane</name>
  <forward mode="bridge"/>
  <bridge name="dataplane"/>
</network>
EOM
virsh net-create /tmp/dataplane.xml
cat > /tmp/access.xml <<EOM
<network>
  <name>access</name>
  <forward mode="bridge"/>
  <bridge name="access"/>
</network>
EOM
virsh net-create /tmp/access.xml
#virsh net-list`

var cleanupL2 = []string{
	"ip link delete access",
	`iptables -F
iptables -X
iptables -t nat -F
iptables -t mangle -F
iptables -X -t nat
ip link del dev cni-podman0
ip link del dev virbr0`,
	"ip link delete baremetal",
	"ip link delete green-net",
	"ip link delete red-net",
	"ip link delete bmc",
	"ip link delete dataplane",
	"rm /etc/cni/net.d/*",
	"virsh net-destroy baremetal",
	"virsh net-destroy access",
	"virsh net-destroy dataplane",
	"rm /tmp/*.xml",
}
