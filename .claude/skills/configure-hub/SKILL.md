---
name: configure-hub
description: Configure SNO as ACM+ZTP hub — LVMS, GitOps, ACM, TALM, Metal3, AgentServiceConfig, Observability
user-invocable: true
trigger: configure hub, setup hub, hub day2, install hub operators, setup ztp hub, configure acm
---

Turn a freshly deployed SNO into a ZTP hub cluster.
Source: telco-hub/configuration/reference-crs/required/acm/

# Non-derivable facts

## Operator table

| Component | Package | Namespace | Channel | Source |
|-----------|---------|-----------|---------|--------|
| LVMS | lvms-operator | openshift-storage | stable-`<OCP_MINOR>` | redhat-operators |
| GitOps | openshift-gitops-operator | openshift-gitops-operator | gitops-1.`<OCP_MINOR-4.1>` | redhat-operators |
| ACM | advanced-cluster-management | open-cluster-management | release-2.17 | redhat-operators |
| TALM | topology-aware-lifecycle-manager | openshift-operators | stable | redhat-operators |

`OCP_MINOR`: `oc get clusterversion -o jsonpath='{.status.desired.version}' | cut -d. -f1-2`
ACM channel: `oc get packagemanifest advanced-cluster-management -o jsonpath='{.status.defaultChannel}'`

## Known pitfalls

- **LVMS Konflux FBC** (`lvm-operator-catalog:v5.0`) references production `registry.redhat.io` bundle digests not mirrored to quay.io — IDMS cannot fix. Use `redhat-operators`.
- **TALM**: no Konflux FBC catalog image. Use `redhat-operators`.
- **ACM subscription name**: `open-cluster-management-subscription` (not the package name).
- **MCH annotations**: required for connected cluster — sets `redhat-operators` as source for mce and oadp sub-operators. Without them MCH pulls from `redhat-operators-disconnected` and fails.
- **AgentServiceConfig**: no namespace in CR (cluster-scoped or auto-namespaced by MCH). Sizes: 20Gi db/fs, 100Gi image. `storageClassName` must be explicit (`lvms-vg1`). ISO filename is `rhcos-live-iso.x86_64.iso` (not `rhcos-live.x86_64.iso` — returns 404). RHCOS version string: `oc get node <node> -o jsonpath='{.status.nodeInfo.osImage}'` → strip prefix, e.g. `9.8.20260617-0`.
- **GitOps channel + namespace**: channel follows OCP minor — `gitops-1.(X-1)` for OCP `4.X` (e.g. 4.22 → `gitops-1.21`). Namespace must be `openshift-gitops-operator`, not `openshift-operators`.
- **MCE cluster-proxy-addon**: MCH auto-creates MCE with `cluster-proxy-addon: enabled`. Must patch to `enabled: false` post-MCH to match telco-hub reference (telco uses direct connectivity).
- **Observability OBC**: requires NooBaa (`openshift-storage.noobaa.io`) — NOT available with LVMS only. On LVMS-only SNO: skip OBC + thanos-secret-policy, create `thanos-object-storage` secret manually from external S3.
- **SNO VM** ships with only `vda` (OS). Must add `vdb` before LVMCluster.
- **SNO memory**: 32 GiB needed for full hub stack. Default 16 GiB is insufficient (GitOps pods will be Pending).

## Lab-specific

```
DOCKER_HOST=ssh://lab0
Container: clab-vlab-bmh1 (SNO), clab-vlab-bmh3 (hub)
VM name inside container: vm1
Redfish BMC SNO: http://10.10.10.11:8000/redfish/v1/Systems/11111111-1111-1111-1111-111111111111
StorageClass from LVMS: lvms-vg1
```

Memory resize (if VM has no hotplug slots):
```bash
# Power off via redfish, update XML, power on via redfish
curl -s -X POST http://10.10.10.11:8000/redfish/v1/Systems/11111111-1111-1111-1111-111111111111/Actions/ComputerSystem.Reset \
  -H 'Content-Type: application/json' -d '{"ResetType":"ForceOff"}'
DOCKER_HOST=ssh://lab0 docker exec clab-vlab-bmh1 virsh setmaxmem vm1 33554432 --config
DOCKER_HOST=ssh://lab0 docker exec clab-vlab-bmh1 virsh setmem vm1 33554432 --config
curl -s -X POST http://10.10.10.11:8000/redfish/v1/Systems/11111111-1111-1111-1111-111111111111/Actions/ComputerSystem.Reset \
  -H 'Content-Type: application/json' -d '{"ResetType":"On"}'
```

# Deploy order

1. Add `vdb` to `vm1` in `clab-vlab-bmh1` (blocker for LVMCluster)
2. LVMS operator → LVMCluster → wait `lvms-vg1` SC
3. GitOps operator (ArgoCD — bootstrap for remaining installs)
4. ACM operator → MultiClusterHub (10-15 min)
5. TALM operator
6. Metal3 Provisioning CR
7. Mirror registry ConfigMap (connected = empty)
8. AgentServiceConfig (needs `lvms-vg1` + MCH Running)
9. Verify MultiClusterEngine Available
10. Observability NS → pull-secret policy → OBC → thanos secret → MCO (needs NooBaa or external S3)
11. Search perf tuning (after everything Running)

# Step 1 — Add data disk to SNO VM

```bash
DOCKER_HOST=ssh://lab0 docker exec clab-vlab-bmh1 \
  qemu-img create -f qcow2 /var/lib/libvirt/images/vm1-data.qcow2 50G
DOCKER_HOST=ssh://lab0 docker exec clab-vlab-bmh1 \
  virsh attach-disk vm1 /var/lib/libvirt/images/vm1-data.qcow2 vdb \
  --persistent --subdriver qcow2 --driver qemu
```

# Step 2 — LVMS

```bash
# Add cluster-monitoring label — required, LVMS operator does not set it
oc label namespace openshift-storage openshift.io/cluster-monitoring=true
oc apply -f .claude/skills/configure-hub/templates/lvms.yaml
oc wait --for=jsonpath='{.status.phase}'=Succeeded \
  csv -l operators.coreos.com/lvms-operator.openshift-storage \
  -n openshift-storage --timeout=300s
oc apply -f .claude/skills/configure-hub/templates/lvmcluster.yaml
oc wait --for=jsonpath='{.status.ready}'=true \
  lvmcluster/lvmcluster -n openshift-storage --timeout=300s
oc get sc  # expect lvms-vg1
```

# Step 3 — GitOps

```bash
oc apply -f .claude/skills/configure-hub/templates/gitops.yaml
oc wait --for=jsonpath='{.status.phase}'=Succeeded \
  csv -l operators.coreos.com/openshift-gitops-operator.openshift-gitops-operator \
  -n openshift-gitops-operator --timeout=300s
```

# Step 4 — ACM

```bash
oc apply -f .claude/skills/configure-hub/templates/acm.yaml
# Wait: subscription state = AtLatestKnown
oc -n open-cluster-management get subscription open-cluster-management-subscription \
  -o jsonpath='{.status.state}'
oc apply -f .claude/skills/configure-hub/templates/multiclusterhub.yaml
oc wait --for=jsonpath='{.status.phase}'=Running \
  multiclusterhub/multiclusterhub -n open-cluster-management --timeout=900s
```

# Step 5 — TALM

```bash
oc apply -f .claude/skills/configure-hub/templates/talm.yaml
```

# Step 6 — Metal3 Provisioning

```bash
oc apply -f .claude/skills/configure-hub/templates/provisioning.yaml
```

# Step 7 — Mirror registry ConfigMap (connected cluster)

```bash
oc apply -f .claude/skills/configure-hub/templates/mirror-registry-cm.yaml
```

# Step 8 — AgentServiceConfig

```bash
# Get RHCOS version string from a running node (spoke target version)
oc get node <node> -o jsonpath='{.status.nodeInfo.osImage}' | grep -oP '[0-9]+\.[0-9]+\.[0-9]+-[0-9]+'
# ISO filename pattern: rhcos-live-iso.x86_64.iso (NOT rhcos-live.x86_64.iso — 404)
# rootFS pattern:      rhcos-live-rootfs.x86_64.img
# Base URL: https://mirror.openshift.com/pub/openshift-v4/x86_64/dependencies/rhcos/<MINOR>/<MINOR>.0/
oc apply -f .claude/skills/configure-hub/templates/agent-service-config.yaml
```

# Step 9 — Patch MCE + Verify

MCH auto-creates MCE with defaults that differ from telco-hub reference. Two patches needed:

```bash
# 1. Disable cluster-proxy-addon (MCH default: on, telco-hub reference: off)
# 2. Enable image-based-install-operator (MCH default: off, telco-hub reference: on)
# Use jq to avoid fragile index-based patches:
oc get multiclusterengine multiclusterengine -o json | \
  jq '(.spec.overrides.components[] | select(.name == "cluster-proxy-addon")).enabled = false |
      (.spec.overrides.components[] | select(.name == "image-based-install-operator")).enabled = true' | \
  oc apply -f -

oc get multiclusterengine multiclusterengine -o jsonpath='{.status.phase}'
# expect: Available
```

# Step 10 — Observability (requires NooBaa or external S3)

```bash
oc apply -f .claude/skills/configure-hub/templates/observability-ns.yaml
oc apply -f .claude/skills/configure-hub/templates/pull-secret-policy.yaml
# If NooBaa available:
oc apply -f .claude/skills/configure-hub/templates/observability-obc.yaml
oc apply -f .claude/skills/configure-hub/templates/thanos-secret-policy.yaml
# Edit storageClass in observability-mco.yaml if needed, then:
oc apply -f .claude/skills/configure-hub/templates/observability-mco.yaml
```

# Step 11 — Search perf tuning

```bash
oc apply -f .claude/skills/configure-hub/templates/search-perf.yaml
```
