# Agent-based/ZTP on virtual infrastructure

POC on OCP v4.1X.Z with complex network setup

## Network

![net-diagram](net-diagram.png)

## Experiments

- Dual stack / Single stack (static, DHCPv4, IPv6 SLAAC, DHCPv6)
- W/o proxy for the connected installation
- RoutingViaHost (=local gateway) (instead of default shared gateway)
- Network tuning e.g MTU 9000k, bond on the primary interface, VLANs on top of bond
- Secondary networks
- More than one Gateway setup (VRRP)
- Operators Metallb, NMState
- MNO with ZTP (spokes)
- SNO with Agent-based & ACM on top (hub)

## Run me

```
ssh root@lab0
dnf -y install libvirt libvirt-daemon-driver-qemu qemu-kvm jq conntrack tcpdump bind-utils wireguard-tools
TODO docker
systemctl enable --now libvirtd
# https://www.howtogeek.com/devops/how-and-why-to-use-a-remote-docker-host/
systemctl disable firewalld && systemctl stop firewalld
hostnamectl set-hostname lab0

echo ip_tables > /etc/modules-load.d/ip_tables.conf
curl https://raw.githubusercontent.com/karmab/kcli/main/install.sh | sudo bash
kcli create pool -p /var/lib/libvirt/images default

cd /tmp \
  && curl https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/latest/openshift-client-linux.tar.gz -o openshift-client-linux.tar.gz && tar xvfz openshift-client-linux.tar.gz \
  && mv oc kubectl /usr/bin/ && chmod +x /usr/bin/{oc,kubectl} && rm -f README.md openshift-client-linux.tar.gz


git clone https://github.com/karampok/telco-ocp-lab.git
cd telco-ocp-lab
#scp ~/.pull-secret.json ~/.id-rsa.pub root@lab0:/root/
grep -E '\s{10,}' .github/workflows/ztp-compact.yaml | sed 's/^          //'
```

## Notes

```
# in ci runner
dnf install  podman-2:4.6.0-3.el9 #if not DNS fails
#disable ipv6
# nmcli connection modify enp1s0 ipv6.method "disabled"
# cat /proc/sys/net/ipv6/conf/enp1s0/disable_ipv6
```
