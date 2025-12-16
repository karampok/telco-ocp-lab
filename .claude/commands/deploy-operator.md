---
description: Deploy an OpenShift operator using Konflux-built images and verify deployment
---

Deploy an OpenShift operator (nmstate, sriov, metallb, or pfstatus) using the
bin/deploy-operator.sh script and verify the deployment is successful.

# Operator-Specific Infos

Before starting deployment, read the operator-specific configuration file:
- Read `.claude/commands/skills/<operator-name>.md` for operator-specific details
- This file contains: namespace, subscription name, CatalogSource name, and verification steps
- Use this information throughout the deployment process

# Pre-deployment checks

1. Verify cluster access and version:
   - Get API server URL: `oc whoami --show-server`
   - Get cluster version: `oc get clusterversion version -o jsonpath='{.status.desired.version}'`

2. Check if operator is already installed:
   - Check namespace exists: `oc get namespace <operator-namespace> 2>/dev/null`
   - Check for existing subscription: `oc get subscription -n <operator-namespace> 2>/dev/null`
   - Check CSV status: `oc get csv -n <operator-namespace> -o jsonpath='{.items[*].status.phase}' 2>/dev/null`
   - Determine state:
     * Not installed: namespace doesn't exist or no subscription found
     * Failed: CSV phase is "Failed" or subscription has CatalogSourcesUnhealthy=True
     * Installed: CSV phase is "Succeeded" and pods are running
   - If already installed print the summary and stop and do the Operator-Specific Infos
   - If in failed state, ask user whether to:
     * Debug why is failing
     * Clean up and reinstall
     * Abort

3. List existing IDMS configurations:
   - List all IDMS: `oc get imagedigestmirrorset`
   - Check if operator-specific IDMS exists: `oc get imagedigestmirrorset <operator>-art-idms 2>/dev/null`
   - If exists, show mirrors: `oc get imagedigestmirrorset <operator>-art-idms -o yaml | grep -A 10 imageDigestMirrors`
   - Warn if IDMS already exists (will be replaced)

# Steps

0. Disable `oc patch OperatorHub cluster --type json -p '[{"op": "add", "path": "/spec/disableAllDefaultSources","value": true}]'`

1. Ask the user which operator to deploy if not specified (nmstate, sriov, metallb, pfstatus)

2. Run the deploy-operator.sh script to generate YAML manifests

3. Apply the generated YAML files in this order:
   - IDMS (ImageDigestMirrorSet)
   - CatalogSource
   - Operator deployment YAML (namespace, operatorgroup, subscription)

4. Perform verification checks in sequence:

   a. IDMS verification:
      - Check IDMS is created: `oc get imagedigestmirrorset <operator>-art-idms`
      - Verify status if available

   b. MachineConfigPool check (optional, may not trigger):
      - Check if MCPs are updating: `oc get mcp`
      - If updating, wait for completion with timeout

   c. CatalogSource verification:
      - Check CatalogSource exists: `oc get catalogsource <operator>-konflux -n openshift-marketplace`
      - Wait for READY state: `oc get catalogsource <operator>-konflux -n openshift-marketplace -o jsonpath='{.status.connectionState.lastObservedState}'`
      - Should show "READY"
      - Check for any errors: `oc get catalogsource <operator>-konflux -n openshift-marketplace -o yaml`

   d. Namespace verification:
      - Confirm namespace created: `oc get namespace <operator-namespace>`

   e. OperatorGroup verification:
      - Confirm created: `oc get operatorgroup -n <operator-namespace>`

   f. Subscription verification:
      - Check subscription exists: `oc get subscription <operator-name> -n <operator-namespace>`
      - Check subscription status: `oc get subscription <operator-name> -n <operator-namespace> -o yaml`
      - Look for conditions, especially CatalogSourcesUnhealthy=False

   g. InstallPlan verification:
      - Find InstallPlan: `oc get installplan -n <operator-namespace>`
      - Check phase is "Complete": `oc get installplan -n <operator-namespace> -o jsonpath='{.items[0].status.phase}'`

   h. ClusterServiceVersion (CSV) verification:
      - Find CSV: `oc get csv -n <operator-namespace>`
      - Check phase is "Succeeded": `oc get csv -n <operator-namespace> -o jsonpath='{.items[0].status.phase}'`
      - Verify CSV conditions

   i. Operator pods verification:
      - List all pods: `oc get pods -n <operator-namespace>`
      - Check all pods are Running and Ready
      - Wait with timeout: `oc wait --for=condition=Ready pods --all -n <operator-namespace> --timeout=300s`

5. Provide a summary report with:
   - Operator name and namespace
   - Version deployed (from CSV)
   - All related images with full digests
   - IDMS mirror configuration (source and mirror registries)
   - Status of each verification step
   - Any warnings or errors encountered
   - Overall deployment status (SUCCESS/FAILED/PARTIAL)

   Summary template:
   ```
   Version: <cluster-version>
   Cluster: <api-url>

   Operator: <operator-name>
   Status: ✓ ALREADY INSTALLED (Age: <age>) | ✓ SUCCESSFULLY DEPLOYED | ✗ FAILED

   Deployment Components:
     ✓ Namespace: <operator-namespace> (Active)
     ✓ CSV: <csv-name> (Succeeded)
     ✓ CatalogSource: <catalogsource-name> (READY)
     ✓ Subscription: <subscription-name>
         Package: <package-name>
         Source: <catalogsource-name>
         Channel: <channel-name>

   Pods and Images:
     <pod-name> (Running, Ready: <ready>/<total>)
       <container-name>: <source-image>@sha256:<digest>
                         <mirror-registry> (IDMS mirror)
       Git Commit: <commit-hash> (extracted from logs if available)
     podman pull --authfile=/tmp/t.c \
       <source-image>@sha256:<full-digest>

   Repository: <github-repo-url>

   Next Steps (Optional Post-Deployment):
     - <operator-specific post-deployment steps>
   ```

# Error handling

- If any step fails, report the error but continue with remaining checks
- Provide troubleshooting suggestions for common issues:
  - CatalogSource not ready: Check image pull, network connectivity
  - Subscription unhealthy: Check CatalogSource availability
  - InstallPlan failed: Check CSV compatibility, image pull issues
  - Pods not ready: Check logs with `oc logs -n <namespace> <pod-name>`
  - Image pull failures: Analyze pull secrets (see remediation below)

## Pull Secret Analysis (Only on Image Pull Failures)

IMPORTANT: Only perform pull secret analysis when pods show ImagePullBackOff or ErrImagePull errors.
Do NOT run pull secret checks proactively for successful deployments.

When image pull failures occur, analyze pull secrets to verify registry credentials:

1. Get pull secret registries (simple check, can be done without user approval):
   ```bash
   oc get secret/pull-secret -n openshift-config -o json | \
     jq -r '.data.".dockerconfigjson"' | base64 -d | jq -r '.auths | keys[]'
   ```

2. Check if IDMS mirror registry has credentials (requires user approval due to sensitive data):
   ```bash
   oc get secret/pull-secret -n openshift-config -o json | \
     jq -r '.data.".dockerconfigjson"' | base64 -d | \
     jq -r '.auths | to_entries[] | select(.value.auth != null) | "\(.key): \(.value.auth)"' | \
     while read -r line; do
       registry=$(echo "$line" | cut -d: -f1)
       auth=$(echo "$line" | cut -d: -f2- | xargs)
       username=$(echo "$auth" | base64 -d 2>/dev/null | cut -d: -f1)
       if [ ${#username} -gt 60 ]; then
         echo "$registry: ${username:0:10}...${username: -10}"
       else
         echo "$registry: $username"
       fi
     done
   ```

3. Display in summary (only when pull secret analysis was performed):
   ```
   Pull Secret Registries:
     ✓ <registry-1> (<username> or <prefix>...<suffix> if long)
     ✓ <registry-2> (<username> or <prefix>...<suffix> if long)
     ✓/✗ <idms-mirror-registry> (<username> or <prefix>...<suffix> if long)
     Note: Truncate usernames longer than 60 chars to first 10 and last 10 chars with ... in between
   ```

## Remediation for Missing Pull Secrets

If IDMS mirror registry shows ✗ (missing pull secret), provide user with manual steps:

**User must execute manually:**

```bash
# Extract current pull-secret
oc get secret/pull-secret -n openshift-config -o json | \
  jq -r '.data.".dockerconfigjson"' | base64 -d >authfile

# Login to the registry (user provides credentials interactively)
if ! podman login --authfile authfile <registry-url>; then
  rm authfile
  exit 1
fi

# Update the pull-secret
if ! oc set data secret/pull-secret -n openshift-config \
  --from-file=.dockerconfigjson=authfile; then
  rm authfile
  exit 1
fi

rm authfile
```

# Output format

Provide clear, structured output showing:
- Each verification step with status (✓ or ✗)
- Timing information where relevant
- Commands executed for transparency
- Final summary with actionable next steps if issues found

# OLM v1 Alternative (OpenShift 4.17+)

This command uses OLM v0 (CatalogSource, Subscription, CSV). OpenShift 4.17+ supports OLM v1 with
improved security and simplified resources.

## Key Differences

| OLM v0 (Current Script) | OLM v1 (Modern) |
|-------------------------|-----------------|
| CatalogSource (namespaced) | ClusterCatalog (cluster-scoped) |
| OperatorGroup | ServiceAccount with RBAC |
| Subscription | ClusterExtension |
| InstallPlan, CSV | Managed automatically |
| API: operators.coreos.com/v1alpha1 | API: olm.operatorframework.io/v1 |

## OLM v1 Security Model

OLM v1 uses a **least privilege model** requiring explicit ServiceAccount with necessary
permissions. You control what the operator installation can do via RBAC, unlike OLM v0 where
OLM itself had broad permissions.

Benefits:
- Reduced attack surface
- Explicit permission control
- Installation fails if ServiceAccount lacks permissions
- Upgrade fails if new version requires additional permissions (prevents silent escalation)

## OLM v1 Example

```yaml
# 1. ClusterCatalog (replaces CatalogSource)
apiVersion: olm.operatorframework.io/v1
kind: ClusterCatalog
metadata:
  name: metallb-konflux
spec:
  source:
    type: Image
    image:
      ref: quay.io/redhat-user-workloads/ocp-art-tenant/art-fbc:ocp__4.21__metallb-rhel9-operator
---
# 2. Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: metallb-system
---
# 3. ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: metallb-operator-installer
  namespace: metallb-system
---
# 4. RBAC (cluster-admin for simplicity, use custom ClusterRole in production)
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: metallb-operator-installer-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: metallb-operator-installer
  namespace: metallb-system
---
# 5. ClusterExtension (replaces Subscription)
apiVersion: olm.operatorframework.io/v1
kind: ClusterExtension
metadata:
  name: metallb-operator
spec:
  namespace: metallb-system
  serviceAccount:
    name: metallb-operator-installer
  source:
    sourceType: Catalog
    catalog:
      packageName: metallb-operator
      selector:
        matchLabels:
          olm.operatorframework.io/metadata.name: metallb-konflux
      channel: stable
```

## References

- [Manage operators as ClusterExtensions with OLM v1](https://developers.redhat.com/articles/2025/06/02/manage-operators-clusterextensions-olm-v1)
- [Announcing OLM v1](https://www.redhat.com/en/blog/announcing-olm-v1-next-generation-operator-lifecycle-management)
- [ClusterExtension API](https://docs.redhat.com/en/documentation/openshift_container_platform/4.19/html/operatorhub_apis/clusterextension-olm-operatorframework-io-v1)
- [OLM v1 Multi-tenant Considerations](https://github.com/operator-framework/operator-controller/discussions/269)
- [operator-framework-catalogd](https://github.com/openshift/operator-framework-catalogd)
- [operator-framework-operator-controller](https://github.com/openshift/operator-framework-operator-controller)
