# Virtual Telco OCP Lab using containerlab

Containerlab topology which includes bare-metal VMs
that can be used by assisted installer API or ZTP for deploying OpenShift

## Host prerequisites

* KVM  `/dev/kvm` but no libvirt on host needed(runs inside `bmh1` container).
* Docker with network disabled
* oc, containerlab, nmstatectl binary
* PULL_SECRET=${PULL_SECRET:-~/.pull-secret.json}
* OCP_RELEASE=${OCP_RELEASE:-"quay.io/openshift-release-dev/ocp-release:4.22.0-rc.5-x86_64"}

## Deploy

```bash
# 0. Create release tarball (local machine)
git archive --format=tar.gz --prefix=telco-vlab/ -o telco-vlab.tar.gz HEAD \
  topo.yaml deploy-ocp.sh sushy.sh cetc/ sno-template/ mno-template/ Makefile
scp telco-vlab.tar.gz lab1:~/

# 1. Untar (on lab host)
tar xzf ~/telco-vlab.tar.gz -C ~/
cd ~/telco-vlab

# 2. Destroy existing lab if running
sudo clab destroy --topo topo.yaml 2>/dev/null || true

# 3. Deploy containerlab infrastructure
PUBLICIP=$(ip --json route get 8.8.8.8 | jq -r '.[].prefsrc') sudo -E clab deploy --topo topo.yaml

# 4. Run OCP deployment from infra container
DOCKER_HOST=ssh://lab1 docker exec clab-vlab-infra ./deploy-ocp.sh
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



## Check if lab is up with OCP

```bash
# Quick check: containers running + OCP nodes ready (from local machine)
DOCKER_HOST=ssh://lab1 docker exec clab-vlab-infra oc get nodes
```

Expected output when healthy: all nodes `Ready`, roles assigned.

## Connect local to cluster

From local machine:
- access to remote docker containers `DOCKER_HOST ssh://lab1 docker ps`
- setup wg follow `docker logs clab-vlab-infra`, verify with `ssh -p 2022 root@10.0.0.1`
- access dns `resolvectl query api.sno.telco.vlab` once wg up
- ping access ipv4 ipv6 to all IPs from topo.yaml file
- Fetch kubeconfig `scp lab:/sno/autho/kubeconfig ~/.kube/lab0.yaml` if ABI
- Fetch kubeconfig `curl -s http://10.0.0.1:9000/sno/auth/kubeconfig > ~/.kube/lab0.yaml` if ZTP


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
