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
- `metadata.namespace`: must be `clusters-sub` (where ArgoCD syncs)
- `templateRefs` cluster: `ai-cluster-templates-v1` in `open-cluster-management`
- `templateRefs` node: `ai-node-templates-v1` in `open-cluster-management`
- `clusterImageSetNameRef`: naming pattern `img<VERSION>-x86-64-appsub` — e.g. `img4.22.2-x86-64-appsub`. Check: `oc get clusterimageset | grep 4.22`
- Per-node for VMs: `ironicInspect: "disabled"`, `automatedCleaningMode: disabled`
- BMC address format: `redfish-virtualmedia+http://10.10.10.XX:8000/redfish/v1/Systems/<UUID>`
- `bootMode: UEFI` (not UEFISecureBoot for KVM VMs)

## Secrets — must exist in `clusters-sub` before sync

```bash
# Pull secret
oc create secret generic assisted-deployment-pull-secret \
  -n clusters-sub --from-file=.dockerconfigjson=~/.pull-secret.json \
  --type=kubernetes.io/dockerconfigjson

# BMC secrets (sushy-emulator accepts any credentials)
for node in mno-master-0 mno-master-1 mno-master-2 mno-worker-0; do
  oc create secret generic ${node}-bmh-secret -n clusters-sub \
    --from-literal=username=admin --from-literal=password=password
done
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

# Check status

```bash
oc get applications.argoproj.io -n openshift-gitops
oc get clusterinstance -n clusters-sub
oc get agentclusterinstall -n <cluster-name>
oc get bmh -n <cluster-name>
```
