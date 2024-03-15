package pkg

var workstation = `
ID=$(hostnamectl --static)
PUBLICIP=$(ip -json route get 8.8.8.8 |jq -r .[0].prefsrc)
podman run --name workstation --rm -d --privileged --hostname w$ID \
--dns-search "telco.vlab" --pull=always --dns 10.10.20.10 --pull=always \
--net=access:interface_name=access  -e PUBLICIP=$PUBLICIP \
-v $(pwd):/workdir --workdir=/workdir \
-v $HOME/.ssh/authorized_keys:/root/.ssh/authorized_keys:ro \
-v /var/run/libvirt:/var/run/libvirt:rw \
--sysctl net.ipv6.conf.all.forwarding=1 --sysctl net.ipv4.ip_forward=1 \
--sysctl net.ipv4.conf.all.src_valid_mark=1 \
-v /lib/modules:/lib/modules quay.io/karampok/infra:latest`

//-v /run/n/podman.sock:/var/podman:rw \

var workstationConfig = `ns=$(podman inspect workstation | jq -r '.[0]["NetworkSettings"].SandboxKey')
ip netns exec "${ns##*/}" ip addr add 10.10.20.200/24 dev access && ip link set dev access up
ip netns exec "${ns##*/}" ip route add 192.168.100.0/24 via 10.10.20.254
ip netns exec "${ns##*/}" ip route add 10.10.10.0/24 via 10.10.20.1
ip netns exec "${ns##*/}" ip route add default via 10.10.20.254
iptables -t nat -A PREROUTING -p udp --dport 51820 -j DNAT --to 10.10.20.200
conntrack -D conntrack dport 51820 || true
`

var cleanup05 = `podman stop workstation`
