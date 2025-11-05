# MetalLB Configuration

Namespace: `metallb-system`

## Prerequisites

Verify the MetalLB operator is running:

```bash
# Check operator pods are running
oc get pods -n metallb-system -l control-plane=controller-manager
oc get pods -n metallb-system -l control-plane=webhook-server
```

## Deploy MetalLB Instance

After the operator is installed, deploy a MetalLB instance to create the controller and speaker pods.

```yaml
---
apiVersion: metallb.io/v1beta1
kind: MetalLB
metadata:
  name: metallb
  namespace: metallb-system
spec:
  logLevel: debug
  nodeSelector:
    node-role.kubernetes.io/worker: ''
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: env-overrides
  namespace: openshift-frr-k8s
data:
  frrk8s-loglevel: "--log-level=debug"
```

This will create:
- MetalLB Controller (deployment) - manages MetalLB resources and IP allocation
- MetalLB Speaker (daemonset) - announces IPs via BGP or L2, runs on each worker node
- FRRK8s pods (daemonset) - FRR routing daemon for BGP, deployed by CNO in `openshift-frr-k8s` namespace

## Verify MetalLB Instance

```bash
# Check MetalLB instance status
oc get metallb -n metallb-system
oc get metallb metallb -n metallb-system -o yaml

# Check instance conditions
oc get metallb metallb -n metallb-system -o jsonpath='{.status.conditions[*].type}'

# Check controller pod (1 replica)
oc get pods -n metallb-system -l component=controller

# Check speaker pods (runs on each worker node)
oc get pods -n metallb-system -l component=speaker

# Wait for all pods to be ready
oc wait --for=condition=Ready pods -l component=controller -n metallb-system --timeout=120s
oc wait --for=condition=Ready pods -l component=speaker -n metallb-system --timeout=120s
```

Expected MetalLB instance status conditions:
- Available: True
- Upgradeable: True
- Progressing: False
- Degraded: False

## Verify FRRK8S

FRRK8s (FRRouting for Kubernetes) is deployed automatically by CNO (Cluster Network Operator)
when a MetalLB instance is created. It runs in the `openshift-frr-k8s` namespace.

**Note**: For detailed FRRk8s operations see the **frrk8s skill**.

**Note**: For FRR commands, BGP session verification, vtysh usage, and advanced
FRR configuration, refer to the **frr skill**.

## Available API Resources

MetalLB provides custom resources in two API groups:

### metallb.io/v1beta1 and v1beta2
```bash
oc api-resources --api-group=metallb.io
```

Resources:
- `bfdprofiles` - BFD (Bidirectional Forwarding Detection) profiles for fast failure detection
- `bgpadvertisements` - BGP advertisement configuration
- `bgppeers` (v1beta2) - BGP peer configuration
- `communities` - BGP community tags
- `ipaddresspools` - IP address pool definitions
- `l2advertisements` - Layer 2 advertisement configuration
- `metallbs` - MetalLB instance CR
- `servicebgpstatuses` - BGP status for services
- `servicel2statuses` - L2 status for services


## Configure MetalLB Resources

After the MetalLB instance is deployed and running, configure IP address pools and advertisements:

```bash
# Check MetalLB instance
oc get metallb -n metallb-system

# Configure IP address pools
oc get ipaddresspool -n metallb-system

# Configure BGP advertisement
oc get bgpadvertisement -n metallb-system

# Configure L2 advertisement
oc get l2advertisement -n metallb-system

# Check BGP peers
oc get bgppeers -n metallb-system

# Check service status
oc get servicebgpstatuses -n metallb-system
oc get servicel2statuses -n metallb-system

# Check FRRK8s resources (see frr skill for details)
oc get frrconfigurations -n openshift-frr-k8s
oc get bgpsessionstates -n openshift-frr-k8s
oc get frrnodestates
```

Example IPAddressPool:
```yaml
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: example-pool
  namespace: metallb-system
spec:
  addresses:
  - 192.168.1.100-192.168.1.200
```

Example BGPAdvertisement:
```yaml
apiVersion: metallb.io/v1beta1
kind: BGPAdvertisement
metadata:
  name: example-bgp
  namespace: metallb-system
spec:
  ipAddressPools:
  - example-pool
```

Example BGPPeer:
```yaml
apiVersion: metallb.io/v1beta2
kind: BGPPeer
metadata:
  name: example-peer
  namespace: metallb-system
spec:
  myASN: 64500
  peerASN: 64501
  peerAddress: 192.168.1.1
```

## Deployment Summary

Get a complete overview of the MetalLB deployment with this script:

```bash
#!/bin/bash

echo "MetalLB Instance Summary"
echo "  Cluster: $(oc whoami --show-server)"
echo "  Version: $(oc get clusterversion version -o jsonpath='{.status.desired.version}')"
echo "  Operator CSV: $(oc get csv -n metallb-system -o jsonpath='{.items[0].metadata.name}')"
echo ""

echo "  Operator Pods (metallb-system):"
oc get pods -n metallb-system -o json | jq -r '
  .items[] |
  select(.metadata.labels["control-plane"] // "" | length > 0) |
  "    " + .metadata.name + "    " + .status.phase + "  (" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + ")\n" +
  (.spec.containers | map("      " + .image) | join("\n"))
' | while IFS= read -r line; do
  echo "$line"
  if [[ "$line" =~ ^[[:space:]]{4}[^[:space:]] ]]; then
    POD_NAME=$(echo "$line" | awk '{print $1}')
    COMMIT=$(oc logs -n metallb-system "$POD_NAME" 2>/dev/null | grep -i "git commit" | head -1 | grep -oE '[a-f0-9]{40}' | head -1)
    if [ -n "$COMMIT" ]; then
      echo "      Git Commit (reported in logs): $COMMIT"
    fi
  fi
done
echo ""

echo "  MetalLB Instance Pods (metallb-system):"
FIRST_SPEAKER=""
oc get pods -n metallb-system -o json | jq -r '
  .items[] |
  select(.metadata.labels.component // "" | length > 0) |
  .metadata.name + "|" + .status.phase + "|" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + "|" + .metadata.labels.component + "|" +
  (.spec.containers | map(.image) | join("|||"))
' | while IFS='|' read -r POD_NAME STATUS READY COMPONENT IMAGES; do
  if [ "$COMPONENT" = "controller" ]; then
    echo "    $POD_NAME    $STATUS  ($READY)"
    echo "$IMAGES" | tr '|||' '\n' | while read -r img; do echo "      $img"; done
    COMMIT=$(oc logs -n metallb-system "$POD_NAME" -c controller 2>/dev/null | grep -i "git commit\|version" | grep -oE '[a-f0-9]{40}' | head -1)
    if [ -n "$COMMIT" ]; then
      echo "      Git Commit (reported in logs): $COMMIT"
    fi
    echo ""
  elif [ "$COMPONENT" = "speaker" ]; then
    if [ -z "$FIRST_SPEAKER" ]; then
      echo "    $POD_NAME    $STATUS  ($READY)"
      echo "$IMAGES" | tr '|||' '\n' | while read -r img; do echo "      $img"; done
      FIRST_SPEAKER="$POD_NAME"
    else
      echo "    $POD_NAME    $STATUS  ($READY)"
    fi
  fi
done
echo ""

echo "  FRRK8s Pods (openshift-frr-k8s):"
FIRST_POD=$(oc get pods -n openshift-frr-k8s -o jsonpath='{.items[0].metadata.name}')
oc get pods -n openshift-frr-k8s -o json | jq -r '
  .items[0:1][] |
  "    " + .metadata.name + "    " + .status.phase + "  (" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + ")\n" +
  ([.spec.containers[].image] | unique | map("      " + .) | join("\n"))
'
COMMIT=$(oc logs -n openshift-frr-k8s "$FIRST_POD" -c controller 2>/dev/null | grep -i "git commit\|version" | grep -oE '[a-f0-9]{40}' | head -1)
if [ -n "$COMMIT" ]; then
  echo "      Git Commit (reported in logs): $COMMIT"
fi
FRR_RELEASE_COMMIT=$(oc adm release info --commits 2>/dev/null | grep "metallb-frr" | awk '{print $3}')
if [ -n "$FRR_RELEASE_COMMIT" ]; then
  echo "      Git Commit (from release): $FRR_RELEASE_COMMIT"
fi
FRR_RPM_VERSION=$(oc exec -n openshift-frr-k8s "$FIRST_POD" -c frr -- rpm -q frr 2>/dev/null)
if [ -n "$FRR_RPM_VERSION" ]; then
  echo "      FRR RPM: $FRR_RPM_VERSION"
fi
oc get pods -n openshift-frr-k8s -o json | jq -r '
  .items[1:][] |
  "    " + .metadata.name + "    " + .status.phase + "  (" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + ")"
'
```

Example output:
```
MetalLB Instance Summary
  Cluster: https://api.5gc.telco.vlab:6443
  Version: 4.21.0-ec.2
  Operator CSV: metallb-operator.v4.21.0-202511040653

  Operator Pods (metallb-system):
    metallb-operator-controller-manager-7b5586d6c6-t7j2l    Running  (1/1)
      registry.redhat.io/openshift4/metallb-rhel9-operator@sha256:1136a4ca...53e118af
      Git Commit (reported in logs): 60e4342ae66ba5a4e1ec1691fc60a39a43d7c078

    metallb-operator-webhook-server-57d9fddf5b-rl5ph        Running  (1/1)
      registry.redhat.io/openshift4/metallb-rhel9@sha256:b0bb169e...0effbd8b

  MetalLB Instance Pods (metallb-system):
    controller-5db9d85fff-jxfvr                             Running  (2/2)
      registry.redhat.io/openshift4/metallb-rhel9@sha256:b0bb169e...0effbd8b
      registry.redhat.io/openshift4/ose-kube-rbac-proxy-rhel9@sha256:aac09485...44844d11
      Git Commit (reported in logs): b0bb169ee40e0fdd8c445ea435aa209abcff16e44eb764e85ca69efe0effbd8b

    speaker-nqht4                                           Running  (2/2)
      registry.redhat.io/openshift4/metallb-rhel9@sha256:b0bb169e...0effbd8b
      registry.redhat.io/openshift4/ose-kube-rbac-proxy-rhel9@sha256:aac09485...44844d11
    speaker-wfxd4                                           Running  (2/2)

  FRRK8s Pods (openshift-frr-k8s):
    frr-k8s-6s424                                           Running  (7/7)
      quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:e76f0c51...3ecee053
      quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:34eb149d...5142ccd9
      Git Commit (reported in logs): a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0
      Git Commit (from release): fc0fe74f94b415b28d772dbc61f6323171a11b50
      FRR RPM: frr-8.5.2-1.el9.x86_64
    frr-k8s-8tqw6                                           Running  (7/7)
    frr-k8s-hp9b9                                           Running  (7/7)
    frr-k8s-vdjt4                                           Running  (7/7)
    frr-k8s-vsqtw                                           Running  (7/7)
```

## Troubleshooting

For troubleshooting MetalLB issues, refer to:
- Official troubleshooting guide: https://metallb.io/troubleshooting/
- **frr skill** for BGP/BFD session troubleshooting and FRR-specific debugging

Check pod logs:
```bash
# Operator logs
oc logs -n metallb-system -l control-plane=controller-manager
oc logs -n metallb-system -l control-plane=webhook-server

# MetalLB instance logs
oc logs -n metallb-system -l component=controller
oc logs -n metallb-system -l component=speaker

# FRRK8s logs (see frr skill for detailed FRR troubleshooting)
oc logs -n openshift-frr-k8s -l app=frr-k8s -c controller
oc logs -n openshift-frr-k8s -l app=frr-k8s -c frr
```

Check resource states:
```bash
# BGP session states (see frr skill for detailed BGP troubleshooting)
oc get bgpsessionstates -n openshift-frr-k8s

# Service status
oc get servicebgpstatuses -n metallb-system
oc get servicel2statuses -n metallb-system
```

## Repository

https://github.com/openshift/metallb-operator
