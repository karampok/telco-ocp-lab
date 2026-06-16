# clab-only — Virtual Telco OCP Lab

Containerlab topology simulating bare-metal SNO (Single Node OpenShift)
deployment via Redfish

## Host prerequisites

KVM  `/dev/kvm`
No libvirt on host needed(runs inside `bmh1` container).
Docker  with network disabled

Required files on host: - `~/.pull-secret.json`

## Deploy

Sync (rsync or git clone) repo to host:

```bash
rsync -av --exclude='clab-vlab/' --exclude='v0/' ./ <remote>:~/clab-only/
```

As root:

```bash
cd ~/clab-only && 
export PUBLICIP=$(ip --json route get 8.8.8.8 | jq -r '.[].prefsrc') 
clab deploy --topo topo.yaml
```

## Network

Three segments: provisioning (BMC/BMH), internal (infra/dns), and host-uplink. See `topo.yaml` for addresses. Host route to lab networks injected by clab via gw1.

## Cleanup

WireGuard (local):
```bash
sudo wg-quick down /tmp/lab.conf
```

Containerlab (must run on remote — host-level cleanup won't work via DOCKER_HOST):
```bash
ssh lab0 "cd ~/clab-only && clab destroy --topo topo.yaml"
```

## Verify

DNS (requires WireGuard up):
```bash
resolvectl query api.sno.telco.vlab
# or: dig @10.10.20.10 api.sno.telco.vlab
```

OCP install config is constructed from `sno-template/` (`baseDomain: telco.vlab`, cluster name `sno`). CoreDNS (`cetc/dns/zones/`) must match: `api.<cluster>.<baseDomain>` and `*.apps.<cluster>.<baseDomain>`.

## OCP Deployment

Trigger from inside infra container:
```bash
docker exec -it clab-vlab-infra bash
./deploy-ocp.sh sno
```

Monitor install progress:
```bash
openshift-install agent wait-for install-complete --log-level info --dir /share/sno
```

Fetch kubeconfig after first node reboot (over WireGuard, sno or mno):
```bash
curl -s http://10.0.0.1:9000/sno/auth/kubeconfig > ~/.kube/lab0.yaml
export KUBECONFIG=~/.kube/lab0.yaml
oc get clusteroperators -w
```

SSH to rendezvous node during install (requires WireGuard or direct lab access):
```bash
ssh -i ~/.ssh/f14ssh/github-actions core@10.10.10.225

journalctl -u assisted-service.service -f    # bootstrap phase
journalctl -u assisted-installer.service -f  # install phase
journalctl -u apply-host-config.service -f   # nmstate/network config applied
sudo crictl ps                               # running containers on node
sudo crictl logs <id>                        # container logs
sudo nmcli                                   # network state
```

## Remote Docker access

```fish
set -x DOCKER_HOST ssh://lab1
docker ps
```

## Multi-lab: creating a second lab instance

When user says "do another lab" or "create lab <name>":

1. Create worktree+branch named `lab-<name>` from current branch
2. Apply all substitutions from the table below (Lab 1 is the baseline)
3. Commit with message: `Configure lab <name> (10.<octet>.0.0/16)`

The second octet distinguishes labs: lab1=`10`, lab2=`11`, lab3=`12`, etc.

| Resource | Lab 1 (baseline) | Lab N (substitute) | Files to change |
|---|---|---|---|
| clab name | `vlab` | `vlab-<name>` | `topo.yaml:name` |
| IP supernet | `10.10.0.0/16` | `10.<N>.0.0/16` | `cetc/clab.nft`, topo.yaml gw1 host route |
| BMC subnet | `10.10.0.x` | `10.<N>.0.x` | topo.yaml (bmc1, gw1) |
| Node subnet | `10.10.10.x` | `10.<N>.10.x` | topo.yaml, `sno-template/agent-config.yaml`, `sno-template/install-config.yaml`, DNS zone, `cetc/gw1/frr.conf` |
| Internal subnet | `10.10.20.x` | `10.<N>.20.x` | topo.yaml (gw1, infra, dns-sidecar), FRR, Corefile |
| IPv6 prefix | `2600:10:10:{10,20}::` | `2600:10:<N>:{10,20}::` | topo.yaml, agent-config, FRR, DNS zone |
| Host veth | `vlab-upstream` | `vlab-<name>-upstream` | topo.yaml link + gw1 stages |
| Host link-local | `169.254.0.1/30`, `.2` | `169.254.0.<4N+1>/30`, `.<4N+2>` | topo.yaml gw1 stages + exec |
| Host route | `via 169.254.0.2` | `via 169.254.0.<4N+2>` | topo.yaml gw1 stage |
| nft table | `inet clab` | `inet clab-<name>` | `cetc/clab.nft` |
| WireGuard port | `51820` | `5182<N>` | `cetc/clab.nft` DNAT, infra WG config |
| libvirt socket | `/tmp/bmh1-libvirt` | `/tmp/bmh-<name>-libvirt` | topo.yaml binds |
| DNS zone / baseDomain | `sno.telco.vlab` | `sno.telco.vlab<N>` | `cetc/dns/Corefile`, `cetc/dns/zones/` (rename+update zone file), `sno-template/install-config.yaml` baseDomain |
| Node IP | `10.10.10.225` | `10.<N>.10.225` | agent-config, DNS zone |
| BGP peer | `10.10.10.100` | `10.<N>.10.100` | `cetc/gw1/frr.conf` |
| FRR router-id | `10.10.10.1` | `10.<N>.10.1` | `cetc/gw1/frr.conf` |

Where `<N>` = 10 + lab index (lab1=10, lab2=11, lab3=12, ...).

Not changed (safe to keep): VM UUID, mgmt network (dummy, no IPs), container images.

## Debugging

VPN (WireGuard) status — check infra logs:
```bash
docker logs clab-vlab-infra
```

Redfish/BMC console per host — check bmh1 logs:
```bash
docker logs clab-vlab-bmh1
```
