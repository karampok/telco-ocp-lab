package pkg

var bridges = `ip link add name dataplane type bridge
ip link set dev dataplane up
ip link add name sw1 type bridge
ip link set dev sw1 up
ip link add name ixp-net type bridge
ip link set dev ixp-net up`

var cmd03 = `cat > /tmp/sw1.xml <<EOM
<network>
  <name>sw1</name>
  <forward mode="bridge"/>
  <bridge name="sw1"/>
</network>
EOM
virsh net-create /tmp/sw1.xml
rm /tmp/sw1.xml
<network>
  <name>dataplane</name>
  <forward mode="bridge"/>
  <bridge name="dataplane"/>
</network>
EOM
virsh net-create /tmp/dataplane.xml
rm /tmp/sw1.xml

#virsh net-list`

var cleanupL2 = []string{
	"ip link delete sw1",
	"ip link delete dataplane",
	"ip link delete ixp-net",
	"virsh net-destroy sw1",
	"virsh net-destroy dataplane",
}
