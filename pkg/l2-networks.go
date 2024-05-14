package pkg

var bridges = `ip link add name dataplane type bridge
ip link set dev dataplane up
ip link add name sw1 type bridge
ip link set dev sw1 up
ip link add name bmc type bridge
ip link set dev bmc up
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
cat > /tmp/dataplane.xml <<EOM
<network>
  <name>dataplane</name>
  <forward mode="bridge"/>
  <bridge name="dataplane"/>
</network>
EOM
virsh net-create /tmp/dataplane.xml
rm /tmp/dataplane.xml

#virsh net-list`
