#!/bin/bash
set -Eeuo pipefail

OPERATOR="${1:-"nmstate"}"
VERSION=$(oc get clusterversion version -o jsonpath='{.status.desired.version}' 2>/dev/null | cut -d. -f1-2)

case "$OPERATOR" in
    sriov)
        OPERATOR_NAME="sriov-network-operator"
        OPERATOR_NAMESPACE="openshift-sriov-network-operator"
        FBC_TAG="ocp__${VERSION}__ose-sriov-network-rhel9-operator"
        ;;
    metallb)
        OPERATOR_NAME="metallb-operator"
        OPERATOR_NAMESPACE="metallb-system"
        FBC_TAG="ocp__${VERSION}__metallb-rhel9-operator"
        ;;
    nmstate)
        OPERATOR_NAME="kubernetes-nmstate-operator"
        OPERATOR_NAMESPACE="openshift-nmstate"
        FBC_TAG="ocp__${VERSION}__kubernetes-nmstate-rhel9-operator"
        ;;
    pfstatus)
        OPERATOR_NAME="pf-status-relay-operator"
        OPERATOR_NAMESPACE="openshift-pf-status-relay"
        FBC_TAG="ocp__${VERSION}__pf-status-relay-rhel9-operator"
        ;;
    *)
        echo "[ERROR] Unknown operator: $OPERATOR"
        exit 1
        ;;
esac
FBC_SOURCE_IMAGE="quay.io/redhat-user-workloads/ocp-art-tenant/art-fbc:${FBC_TAG}"

echo "[INFO] Detected OpenShift version: $VERSION"
echo "[INFO] Operator Name: $OPERATOR_NAME"
echo "[INFO] Operator Namespace: $OPERATOR_NAMESPACE"
echo "[INFO] FBC Image: $FBC_SOURCE_IMAGE"

TEMP_DIR=$(mktemp -d -t "${OPERATOR}-${VERSION}-XXXXXX")
echo "[INFO] Using temp directory: $TEMP_DIR"

cd "$TEMP_DIR"

DIGEST=$(oras discover --format json "$FBC_SOURCE_IMAGE" | \
    jq -r '.referrers[] | select(.annotations.attachedMediaType == "application/vnd.konflux-ci.attached-artifact.related-images+json") | .digest')
oras pull "quay.io/redhat-user-workloads/ocp-art-tenant/art-fbc@${DIGEST}"
RELATED_REPOS=$(cat related-images.json | jq -r '.[]' | sed 's/@sha256:.*//' | sort -u)
echo "$RELATED_REPOS" | nl -w2 -s'. '

cd -

cat > "$TEMP_DIR/idms.yaml" <<EOF
apiVersion: config.openshift.io/v1
kind: ImageDigestMirrorSet
metadata:
  name: ${OPERATOR}-art-idms
spec:
  imageDigestMirrors:
EOF

while IFS= read -r repo; do
    [[ -z "$repo" ]] && continue
    cat >> "$TEMP_DIR/idms.yaml" <<EOF
  - mirrors:
    - quay.io/redhat-user-workloads/ocp-art-tenant/art-images-share
    source: $repo
EOF
done <<< "$RELATED_REPOS"
echo "IDMS written to $TEMP_DIR/idms.yaml"

# echo "[INFO] Waiting for Machine Config Pool update after IDMS creation..."
# oc wait --for=condition=Updating mcp --all --timeout=60s >/dev/null 2>&1 || true
# oc wait --for=condition=Updating=false mcp --all --timeout=600s >/dev/null 2>&1 || true
# echo "[SUCCESS]Machine Config Pool update completed"

# Disable default catalogs
# echo "[INFO] Configuring cluster settings"
# echo "[INFO] Disabling default catalogs..."
# oc patch operatorhub cluster -p '{"spec": {"disableAllDefaultSources": true}}' --type=merge
# echo "[SUCCESS]Default catalogs disabled"

cat > "$TEMP_DIR/catalogsource.yaml" <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: ${OPERATOR}-konflux
  namespace: openshift-marketplace
spec:
  displayName: ${OPERATOR}-konflux
  image: ${FBC_SOURCE_IMAGE}
  sourceType: grpc
  updateStrategy:
    registryPoll:
      interval: 10m
EOF
echo "CatalogSource written to $TEMP_DIR/catalogsource.yaml"

# oc wait --for=jsonpath='{.status.connectionState.lastObservedState}'=READY \
#     catalogsource "$CATALOG_NAME" -n openshift-marketplace --timeout=300s 2>/dev/null || true

cat > "$TEMP_DIR/namespace.yaml" <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: $OPERATOR_NAMESPACE
  labels:
    pod-security.kubernetes.io/enforce: privileged
EOF
echo "Namespace written to $TEMP_DIR/namespace.yaml"

cat > "$TEMP_DIR/operatorgroup.yaml" <<EOF
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: ${OPERATOR_NAME}-og
  namespace: ${OPERATOR_NAMESPACE}
spec:
  targetNamespaces:
  - ${OPERATOR_NAMESPACE}
EOF
echo "OperatorGroup written to $TEMP_DIR/operatorgroup.yaml"

cat > "$TEMP_DIR/subscription.yaml" <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: ${OPERATOR_NAME}
  namespace: ${OPERATOR_NAMESPACE}
spec:
  channel: stable
  installPlanApproval: Automatic
  name: ${OPERATOR_NAME}
  source: ${OPERATOR}-konflux
  sourceNamespace: openshift-marketplace
EOF
echo "Subscription written to $TEMP_DIR/subscription.yaml"

# oc wait --for=condition=CatalogSourcesUnhealthy=False subscription "${OPERATOR_NAME}" \
#     -n "${OPERATOR_NAMESPACE}" --timeout=120s 2>/dev/null || true

# Wait for CSV to be created
# oc wait --for=condition=Ready pods --all -n "$OPERATOR_NAMESPACE" --timeout=180s 2>/dev/null || \
#     { echo "[WARNING]Operator pods did not reach Ready state within timeout"; }
# echo "[SUCCESS] All operator pods are ready"
echo "  oc apply -f $TEMP_DIR"
