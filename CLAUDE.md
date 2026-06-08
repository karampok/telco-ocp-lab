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

## Debugging

VPN (WireGuard) status — check infra logs:
```bash
docker logs clab-vlab-infra
```

Redfish/BMC console per host — check bmh1 logs:
```bash
docker logs clab-vlab-bmh1
```
