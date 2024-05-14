package pkg

var sushy = `podman run --name sushy --rm -d --privileged --hostname bmc-sushy \
-v ./opt/sushy:/etc/sushy:Z -v /var/run/libvirt:/var/run/libvirt:rw --net=bmc:interface_name=eth0 \
 --dns 10.10.20.10 \
--entrypoint='["/bin/bash", "-c", "ip addr add 192.168.100.100/24 dev eth0 && /usr/local/bin/sushy-emulator --config /etc/sushy/emulator.conf --debug"]' \
quay.io/karampok/sushy-emulator:latest`

var sushyNetconfig = `ns=$( podman inspect sushy | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip netns exec "${ns##*/}" ip route add default via 192.168.100.254`

var cleanup04 = `podman stop sushy
kcli delete -y plan vbmh`
