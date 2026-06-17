---
name: deploy-operator
description: Deploy an OpenShift operator using Konflux-built images and verify deployment
user-invocable: true
trigger: deploy operator, install operator, konflux operator, deploy nmstate, deploy sriov, deploy metallb, deploy ptp, deploy pfstatus
---

Deploy a Konflux-built operator on a connected OCP cluster.

# Konflux FBC details (not derivable — must use these exact values)

FBC image pattern:
```
quay.io/redhat-user-workloads/ocp-art-tenant/art-fbc:ocp__<VERSION>__<FBC_TAG>
```

IDMS mirror registry:
```
quay.io/redhat-user-workloads/ocp-art-tenant/art-images-share
```

Operator -> FBC tag mapping (VERSION = `oc get clusterversion -o jsonpath='{.status.desired.version}' | cut -d. -f1-2`):

| Operator | FBC_TAG | Package name | Namespace |
|----------|---------|--------------|-----------|
| sriov | ose-sriov-network-rhel9-operator | sriov-network-operator | openshift-sriov-network-operator |
| metallb | metallb-rhel9-operator | metallb-operator | metallb-system |
| nmstate | kubernetes-nmstate-rhel9-operator | kubernetes-nmstate-operator | openshift-nmstate |
| pfstatus | pf-status-relay-rhel9-operator | pf-status-relay-operator | openshift-pf-status-relay |
| ptp | ose-ptp-rhel9-operator | ptp-operator | openshift-ptp |

VERSION can be overridden to install a different stream (e.g., 4.23 operator on 4.22 cluster). Only the CatalogSource image changes — IDMS repos are the same across versions.

CatalogSource naming: `<OPERATOR>-konflux-v<VERSION>` (e.g., `nmstate-konflux-v4-23`). Include version so multiple streams can coexist and intent is clear from `oc get catalogsource`.

# Pre-flight

1. Check if operator already installed: `oc get csv -n <NS> 2>/dev/null`
   - If Succeeded: report and stop (unless user wants upgrade/reinstall)
   - If Failed: ask user — debug, clean up, or abort
2. Check if IDMS already exists: `oc get imagedigestmirrorset <OP>-art-idms 2>/dev/null`
   - If exists: skip IDMS apply (no MCP reboot needed)
3. Verify FBC image exists: `skopeo inspect --no-tags docker://<FBC_IMAGE> | jq .Digest`

# Extract related images for IDMS (requires `oras`)

Konflux attaches `related-images.json` as an OCI referrer to the FBC image:

```bash
DIGEST=$(oras discover --format json <FBC_IMAGE> | \
  jq -r '.referrers[] | select(.annotations.attachedMediaType == "application/vnd.konflux-ci.attached-artifact.related-images+json") | .digest')
cd $(mktemp -d) && oras pull "quay.io/redhat-user-workloads/ocp-art-tenant/art-fbc@${DIGEST}"
cat related-images.json | jq -r '.[]' | sed 's/@sha256:.*//' | sort -u
```

Fallback if `oras` not available:
```bash
podman pull <FBC_IMAGE>
podman create --name fbc-tmp <FBC_IMAGE>
podman cp fbc-tmp:/configs/ /tmp/fbc-extract/
# parse catalog.yaml for relatedImages
podman rm fbc-tmp
```

IDMS maps repos, not digests — content is stable across builds. Generate once per operator, reuse. Save to `.ailocal/<operator>-art-idms.yaml`.

Channel is `stable` for all operators in the table.

# Deploy order

When deploying multiple operators, batch all IDMS into a single `oc apply` to get one MCP reboot:
```bash
oc apply -f .ailocal/nmstate-art-idms.yaml -f .ailocal/pfstatus-art-idms.yaml
```

1. Disable default catalogs: `oc patch OperatorHub cluster --type json -p '[{"op":"add","path":"/spec/disableAllDefaultSources","value":true}]'`
2. Apply IDMS (batch all operators, skip if already exists)
3. Wait MCP if updating: `oc wait --for=condition=Updating=false mcp --all --timeout=600s`
4. Apply CatalogSource, wait READY: `oc wait --for=jsonpath='{.status.connectionState.lastObservedState}'=READY catalogsource/<NAME> -n openshift-marketplace --timeout=300s`
5. Apply Namespace + OperatorGroup + Subscription
6. Wait CSV Succeeded: `oc wait --for=jsonpath='{.status.phase}'=Succeeded csv -l operators.coreos.com/<PACKAGE> -n <NS> --timeout=300s`
7. Wait pods ready: `oc wait --for=condition=Ready pods --all -n <NS> --timeout=300s`

# Verify

```bash
oc get csv -n <NS>                    # phase=Succeeded
oc get pods -n <NS>                   # all Running/Ready
oc get catalogsource -n openshift-marketplace  # READY
oc get imagedigestmirrorset           # IDMS present
```

# On ImagePullBackOff

Check pull-secret has credentials for IDMS mirror:
```bash
oc get secret/pull-secret -n openshift-config -o json | jq -r '.data.".dockerconfigjson"' | base64 -d | jq '.auths | keys[]'
```

If `quay.io` or `quay.io/redhat-user-workloads` missing, user must add auth via:
```bash
oc get secret/pull-secret -n openshift-config -o json | jq -r '.data.".dockerconfigjson"' | base64 -d > authfile
podman login --authfile authfile quay.io
oc set data secret/pull-secret -n openshift-config --from-file=.dockerconfigjson=authfile
rm authfile
```
