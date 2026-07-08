---
name: how-to-ztp
description: ZTP spoke deployment via ClusterInstance + ArgoCD on the clab SNO hub
user-invocable: true
trigger: ztp, deploy spoke, clusterinstance, argocd ztp, add cluster, mno deploy, sno spoke
---

Deploy spoke clusters via ZTP using ClusterInstance + ArgoCD on the clab hub.
Hub cluster: `sno.telco.vlab` (cluster named `sno`, functioning as ZTP hub).

# Non-derivable facts

## ArgoCD

- URL: `https://openshift-gitops-server-openshift-gitops.apps.sno.telco.vlab`
- Login flags required: `--insecure --grpc-web` (reencrypt route, gRPC-web needed)
- Password: `oc get secret openshift-gitops-cluster -n openshift-gitops -o jsonpath='{.data.admin\.password}' | base64 -d`
- Server version check: `curl -sk https://openshift-gitops-server-openshift-gitops.apps.sno.telco.vlab/api/version | jq .Version`

```bash
argocd login openshift-gitops-server-openshift-gitops.apps.sno.telco.vlab \
  --username admin --insecure --grpc-web
```

## RBAC — ArgoCD SA needs ClusterInstance permission

```bash
oc apply -f .claude/skills/how-to-ztp/templates/argocd-rbac.yaml
```

Two objects:
- `hub-rds-cluster-role` — ClusterInstance write, aggregates into ACM `cluster-manager-admin` via label
- `hub-rds-gitops` — binds ArgoCD SA to `cluster-manager-admin` (covers ClusterInstance + ACM secrets/lifecycle in one binding)

**ClusterInstance spec is immutable** once provisioning starts (admission webhook). To fix imageSet or other spec fields: `oc delete clusterinstance <name> -n <ns>` — ArgoCD recreates from git.

## ArgoCD Application — critical fields

| Field | Value | Why |
|---|---|---|
| `destination.namespace` | cluster name (e.g. `mno`) | ArgoCD respects namespace set in manifest; destination is fallback only |
| `project` | `ztp-app-project` | AppProject must exist first |
| `ignoreDifferences` ManagedCluster | required | ACM controller adds labels ArgoCD would fight against |
| `PrunePropagationPolicy` | `background` | clean deprovisioning |
| `RespectIgnoreDifferences` | `true` | makes ignoreDifferences work with automated sync |
| `directory.recurse` | `false` | flat directory, no kustomize |

## Repo structure

```
ztp/clusters/
├── pre-req/                        ← apply manually before ArgoCD sync
│   ├── ns.yaml                     (cluster namespace)
│   ├── bmc-credentials.yaml        (single secret for all nodes)
│   └── kustomization.yaml
├── mno-clusterinstance.yaml        ← add cluster = add one file here
├── mno2-clusterinstance.yaml
└── sno1-clusterinstance.yaml
```

No kustomization.yaml at top level. ArgoCD directory mode picks up all YAML files.
Pre-req NOT synced by ArgoCD — applied manually once per cluster.

## Pre-req per cluster

```bash
# Apply namespace + BMC secret
oc apply -k ztp/clusters/pre-req/

# Pull secret — never in git, apply manually
oc create secret generic assisted-deployment-pull-secret -n mno \
  --from-file=.dockerconfigjson=/home/kka/.pull-secret.json \
  --type=kubernetes.io/dockerconfigjson
```

## BMC secret — single secret for all nodes (sushy-emulator)

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mno-bmc-secret
  namespace: mno
type: Opaque
stringData:
  username: admin
  password: password
```

All nodes reference `bmcCredentialsName.name: mno-bmc-secret` — one secret per cluster, not per node.

## ClusterInstance — non-derivable fields

- `apiVersion: siteconfig.open-cluster-management.io/v1alpha1`
- `metadata.namespace`: same as `clusterName` (e.g. `mno`) — ArgoCD respects namespace in manifest
- `templateRefs` cluster: `ai-cluster-templates-v1` in `open-cluster-management`
- `templateRefs` node: `ai-node-templates-v1` in `open-cluster-management`
- `clusterImageSetNameRef`: naming pattern `img<VERSION>-x86-64-appsub` — e.g. `img4.22.2-x86-64-appsub`. Check: `oc get clusterimageset | grep 4.22`
- Per-node for VMs: `ironicInspect: "disabled"`, `automatedCleaningMode: disabled`
- BMC address format: `redfish-virtualmedia+http://10.10.10.XX:8000/redfish/v1/Systems/<UUID>`
- `bootMode: UEFI` (not UEFISecureBoot for KVM VMs)

- **`sshPublicKey`**: must be a real key — placeholder `ssh-ed25519 AAAA...` causes `AgentClusterInstall` error: `SSH key does not match any supported type`. Get from `cat ~/.ssh/id_ed25519.pub`.

## Secrets — must exist in cluster namespace before sync

```bash
# Apply namespace + BMC secret
oc apply -k ztp/clusters/pre-req/

# Pull secret — never in git, apply manually
oc create secret generic assisted-deployment-pull-secret -n mno \
  --from-file=.dockerconfigjson=/home/kka/.pull-secret.json \
  --type=kubernetes.io/dockerconfigjson
```

## clab git daemon (laptop → hub connectivity)

Laptop WireGuard IP: `10.0.0.2`
Git URL for ArgoCD: `git://10.0.0.2:9418/repo.git`

Start git daemon (run in separate terminal, foreground):
```bash
docker run --rm --network host \
  -v /home/kka/github.com/karampok/telco-ocp-lab/.git:/srv/repo.git:ro \
  bitnami/git:latest bash -c \
  'git config --global --add safe.directory /srv/repo.git && \
   git daemon --reuseaddr --base-path=/srv --export-all --verbose --port=9418 /srv/repo.git'
```

Routing (persistent in topo.yaml, auto-applied on clab deploy):
- Gateway: `ip route add 10.0.0.0/24 via 10.10.20.200`
- Infra: `iptables -t nat -A POSTROUTING -s 10.10.0.0/16 -d 10.0.0.0/24 -j MASQUERADE`

If clab was redeployed and routes are missing, re-apply manually:
```bash
DOCKER_HOST=ssh://lab0 docker exec clab-vlab-gateway ip route add 10.0.0.0/24 via 10.10.20.200
DOCKER_HOST=ssh://lab0 docker exec clab-vlab-infra iptables -t nat -A POSTROUTING -s 10.10.0.0/16 -d 10.0.0.0/24 -j MASQUERADE
```

Test from hub:
```bash
oc run git-test --image=alpine/git --restart=Never \
  --command -- git ls-remote git://10.0.0.2:9418/repo.git
sleep 5 && oc logs git-test && oc delete pod git-test
```

# Workflow: add a new spoke cluster

1. Create `ztp/clusters/<name>-clusterinstance.yaml` (copy mno, change values)
2. Create secrets in `clusters-sub` namespace (BMC + pull-secret)
3. Commit and push (git daemon serves immediately — no push needed, serves from `.git`)
4. ArgoCD auto-syncs within poll interval

# Bootstrap (one-time)

```bash
# 1. Create AppProject
oc apply -f .claude/skills/how-to-ztp/templates/appproject.yaml

# 2. Create Application
oc apply -f .claude/skills/how-to-ztp/templates/application.yaml

# 3. Create clusters-sub namespace + secrets
oc create ns clusters-sub
# ... create secrets (see above)
```

# Monitor provisioning

Run in a loop — checks all key resources in order of the install flow:

```bash
watch_cluster() {
  NS=${1:-mno}
  echo '=====> INFRAENV:'
  oc get infraenv -n "$NS"
  echo '=====> BMH:'
  oc get bmh -n "$NS"
  echo '=====> Agents:'
  oc get agent -n "$NS" \
    -o custom-columns=HOST:".spec.hostname",ROLE:".spec.role",APPROVED:".spec.approved",STAGE:".status.progress.currentStage",STATEINFO:".status.debugInfo.stateInfo" \
    --sort-by=".spec.hostname" 2>/dev/null || echo "none yet"
  echo '=====> AgentClusterInstall Messages:'
  oc get agentclusterinstall -n "$NS" "$NS" -o jsonpath='{.status.conditions}' | jq '.' | grep message
  echo '=====> ManagedCluster:'
  oc get managedcluster "$NS" 2>/dev/null || echo "not yet"
  echo '=====> ClusterDeployment:'
  oc get clusterdeployment -n "$NS" "$NS"
}

# Run once
watch_cluster mno

# Or loop
while true; do date; watch_cluster mno || true; sleep 60; done
```

## Known issues

- **SSH key placeholder** — `ssh-ed25519 AAAA...` causes AgentClusterInstall error. Use real key.
- **`clusterImageSetNameRef: openshift-4.22`** — does not exist. Use `img4.22.2-x86-64-appsub` pattern.
- **ClusterInstance spec immutable** — admission webhook blocks updates during provisioning. Delete and let ArgoCD recreate.
- **InfraEnv ISO empty** — assisted-image-service downloads RHCOS ISO on first use (~1GB). Wait for `minimal iso created` in image-service logs.
- **ArgoCD RBAC** — ArgoCD SA needs `hub-rds-gitops` ClusterRoleBinding to `cluster-manager-admin` (covers ClusterInstance + ACM secrets).
- **BMH `preparing` for long time** — Metal3 connecting to sushy via redfish. Check sushy container logs in `clab-vlab-bmc<N>`.
