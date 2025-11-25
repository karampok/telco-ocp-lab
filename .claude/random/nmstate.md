# NMState Configuration

Namespace: `openshift-nmstate`

## Prerequisites

Verify the NMState operator is running:

```bash
# Check operator pods are running
oc get pods -n openshift-nmstate -l app=kubernetes-nmstate-operator
oc get pods -n openshift-nmstate -l app=nmstate-webhook
```

## Deploy NMState Instance

After the operator is installed, deploy a NMState instance to create the handler daemonset.

```yaml
apiVersion: nmstate.io/v1
kind: NMState
metadata:
  name: nmstate
spec:
  nodeSelector:
    node-role.kubernetes.io/worker: ''
```

This will create:
- NMState Handler (daemonset) - manages network configuration on each node
- NMState Webhook (deployment) - validates NodeNetworkConfigurationPolicy resources
- NMState Metrics (deployment) - exposes metrics for monitoring
- NMState Console Plugin (deployment) - provides OpenShift Console integration

## Verify NMState Instance

```bash
# Check NMState instance status
oc get nmstate -n openshift-nmstate
oc get nmstate nmstate -n openshift-nmstate -o yaml

# Check instance conditions
oc get nmstate nmstate -n openshift-nmstate -o jsonpath='{.status.conditions[*].type}'

# Check handler pods (runs on each node)
oc get pods -n openshift-nmstate -l component=kubernetes-nmstate-handler

# Check webhook pods (2 replicas)
oc get pods -n openshift-nmstate -l app=nmstate-webhook

# Check metrics pod (1 replica)
oc get pods -n openshift-nmstate -l app=nmstate-metrics

# Check console plugin pod (1 replica)
oc get pods -n openshift-nmstate -l app.kubernetes.io/name=nmstate-console-plugin

# Wait for all pods to be ready
oc wait --for=condition=Ready pods --all -n openshift-nmstate --timeout=120s
```

Expected NMState instance status conditions:
- Available: True
- Degraded: False

## Available API Resources

NMState provides custom resources in the nmstate.io API group:

```bash
oc api-resources --api-group=nmstate.io
```

Resources:
- `nmstates` - NMState instance CR
- `nodenetworkstates` - Current network state of each node (read-only)
- `nodenetworkconfigurationpolicies` - Desired network configuration to apply to nodes
- `nodenetworkconfigurationenactments` - Status of configuration application per node (read-only)

## View Node Network States

NodeNetworkState is automatically created for each node and reports the current network configuration:

```bash
# List all node network states
oc get nodenetworkstates

# View detailed network state for a specific node
oc get nodenetworkstate <node-name> -o yaml

# View specific interface information
oc get nodenetworkstate <node-name> -o jsonpath='{.status.currentState.interfaces[*].name}'

# View interface IP addresses
oc get nodenetworkstate <node-name> -o jsonpath='{.status.currentState.interfaces[?(@.name=="ens1")].ipv4}'
```

## Configure Node Network

Use NodeNetworkConfigurationPolicy to declaratively configure node networking:

```bash
# List all policies
oc get nodenetworkconfigurationpolicies

# View policy status
oc get nodenetworkconfigurationpolicy <policy-name> -o yaml

# Check enactments (per-node status)
oc get nodenetworkconfigurationenactments

# View enactment for specific node
oc get nodenetworkconfigurationenactment <node-name>.<policy-name> -o yaml
```

Example NodeNetworkConfigurationPolicy (VLAN interface):
```yaml
apiVersion: nmstate.io/v1
kind: NodeNetworkConfigurationPolicy
metadata:
  name: vlan100
spec:
  nodeSelector:
    node-role.kubernetes.io/worker: ''
  desiredState:
    interfaces:
    - name: eth0.100
      type: vlan
      state: up
      vlan:
        base-iface: eth0
        id: 100
      ipv4:
        enabled: true
        dhcp: false
        address:
        - ip: 192.168.100.10
          prefix-length: 24
```

Example NodeNetworkConfigurationPolicy (Bond interface):
```yaml
apiVersion: nmstate.io/v1
kind: NodeNetworkConfigurationPolicy
metadata:
  name: bond0
spec:
  nodeSelector:
    node-role.kubernetes.io/worker: ''
  desiredState:
    interfaces:
    - name: bond0
      type: bond
      state: up
      link-aggregation:
        mode: 802.3ad
        options:
          miimon: '100'
        port:
        - eth1
        - eth2
      ipv4:
        enabled: true
        dhcp: true
```

Example NodeNetworkConfigurationPolicy (Bridge):
```yaml
apiVersion: nmstate.io/v1
kind: NodeNetworkConfigurationPolicy
metadata:
  name: br1
spec:
  nodeSelector:
    node-role.kubernetes.io/worker: ''
  desiredState:
    interfaces:
    - name: br1
      type: linux-bridge
      state: up
      bridge:
        options:
          stp:
            enabled: false
        port:
        - name: eth1
      ipv4:
        enabled: true
        dhcp: false
        address:
        - ip: 192.168.1.10
          prefix-length: 24
```

## Monitor Configuration Status

```bash
# Check policy conditions
oc get nodenetworkconfigurationpolicy <policy-name> \
  -o jsonpath='{.status.conditions[*].type}'

# Expected conditions: Available, Degraded

# Check enactment status for all nodes
oc get nodenetworkconfigurationenactments -l nmstate.io/policy=<policy-name>

# View failed enactments
oc get nodenetworkconfigurationenactments \
  -o json | jq -r '.items[] | select(.status.conditions[] | select(.type=="Failing" and .status=="True")) | .metadata.name'
```

## Deployment Summary

Get a complete overview of the NMState deployment:

```bash
#!/bin/bash

echo "NMState Instance Summary"
echo "  Cluster: $(oc whoami --show-server)"
echo "  Version: $(oc get clusterversion version -o jsonpath='{.status.desired.version}')"
echo "  Operator CSV: $(oc get csv -n openshift-nmstate -o jsonpath='{.items[0].metadata.name}')"
echo ""

echo "  Operator Pod (openshift-nmstate):"
oc get pods -n openshift-nmstate -l app=kubernetes-nmstate-operator -o json | jq -r '
  .items[0] |
  "    " + .metadata.name + "    " + .status.phase + "  (" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + ")\n" +
  (.spec.containers | map("      " + .image) | join("\n"))
'
echo ""

echo "  Webhook Pods (openshift-nmstate):"
FIRST_WEBHOOK=$(oc get pods -n openshift-nmstate -l app=nmstate-webhook -o jsonpath='{.items[0].metadata.name}')
oc get pods -n openshift-nmstate -l app=nmstate-webhook -o json | jq -r '
  .items[0:1][] |
  "    " + .metadata.name + "    " + .status.phase + "  (" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + ")\n" +
  (.spec.containers | map("      " + .image) | join("\n"))
'
oc get pods -n openshift-nmstate -l app=nmstate-webhook -o json | jq -r '
  .items[1:][] |
  "    " + .metadata.name + "    " + .status.phase + "  (" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + ")"
'
echo ""

echo "  Handler Pods (openshift-nmstate):"
FIRST_HANDLER=$(oc get pods -n openshift-nmstate -l component=kubernetes-nmstate-handler -o jsonpath='{.items[0].metadata.name}')
oc get pods -n openshift-nmstate -l component=kubernetes-nmstate-handler -o json | jq -r '
  .items[0:1][] |
  "    " + .metadata.name + "    " + .status.phase + "  (" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + ")\n" +
  (.spec.containers | map("      " + .image) | join("\n"))
'
oc get pods -n openshift-nmstate -l component=kubernetes-nmstate-handler -o json | jq -r '
  .items[1:][] |
  "    " + .metadata.name + "    " + .status.phase + "  (" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + ")"
'
echo ""

echo "  Metrics Pod (openshift-nmstate):"
oc get pods -n openshift-nmstate -l app=nmstate-metrics -o json | jq -r '
  .items[0] |
  "    " + .metadata.name + "    " + .status.phase + "  (" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + ")\n" +
  (.spec.containers | map("      " + .image) | join("\n"))
'
echo ""

echo "  Console Plugin Pod (openshift-nmstate):"
oc get pods -n openshift-nmstate -l app.kubernetes.io/name=nmstate-console-plugin -o json | jq -r '
  .items[0] |
  "    " + .metadata.name + "    " + .status.phase + "  (" +
  (.status.containerStatuses | map(select(.ready)) | length | tostring) + "/" +
  (.spec.containers | length | tostring) + ")\n" +
  (.spec.containers | map("      " + .image) | join("\n"))
'
```

Example output:
```
NMState Instance Summary
  Cluster: https://api.5gc.telco.vlab:6443
  Version: 4.21.0-ec.2
  Operator CSV: kubernetes-nmstate-operator.4.21.0-202511041724

  Operator Pod (openshift-nmstate):
    nmstate-operator-65448c8c7b-sj5hg    Running  (1/1)
      registry.redhat.io/openshift4/kubernetes-nmstate-rhel9-operator@sha256:89a4312f...f4480267

  Webhook Pods (openshift-nmstate):
    nmstate-webhook-5f555c89b6-6jd64    Running  (1/1)
      registry.redhat.io/openshift4/ose-kubernetes-nmstate-handler-rhel9@sha256:226bf6be...e8b93a80
    nmstate-webhook-5f555c89b6-p7zcx    Running  (1/1)

  Handler Pods (openshift-nmstate):
    nmstate-handler-dtmmk    Running  (1/1)
      registry.redhat.io/openshift4/ose-kubernetes-nmstate-handler-rhel9@sha256:226bf6be...e8b93a80
    nmstate-handler-m5kpt    Running  (1/1)
    nmstate-handler-p96sp    Running  (1/1)
    nmstate-handler-wsnnb    Running  (1/1)
    nmstate-handler-wvgt6    Running  (1/1)

  Metrics Pod (openshift-nmstate):
    nmstate-metrics-785d586979-78p6t    Running  (2/2)
      registry.redhat.io/openshift4/ose-kubernetes-nmstate-handler-rhel9@sha256:226bf6be...e8b93a80
      registry.redhat.io/openshift4/ose-kube-rbac-proxy-rhel9@sha256:aac09485...44844d11

  Console Plugin Pod (openshift-nmstate):
    nmstate-console-plugin-56f668f8db-s9tpt    Running  (1/1)
      registry.redhat.io/openshift4/nmstate-console-plugin-rhel9@sha256:a7a482ff...eacd3316
```

## Troubleshooting

Check pod logs:
```bash
# Operator logs
oc logs -n openshift-nmstate -l app=kubernetes-nmstate-operator

# Handler logs (check specific node)
oc logs -n openshift-nmstate <handler-pod-name>

# Webhook logs
oc logs -n openshift-nmstate -l app=nmstate-webhook

# Metrics logs
oc logs -n openshift-nmstate -l app=nmstate-metrics -c nmstate-metrics
```

Check configuration status:
```bash
# View policy status
oc get nodenetworkconfigurationpolicy -o wide

# Check enactment failures
oc get nodenetworkconfigurationenactments -A | grep -v Available

# View failed enactment details
oc get nodenetworkconfigurationenactment <name> -o yaml

# Check handler events
oc get events -n openshift-nmstate --field-selector involvedObject.kind=DaemonSet
```

Common issues:
- Configuration conflicts: Check enactment status and failure reason
- Validation errors: Check webhook logs and policy spec
- Handler not running: Check daemonset and node selectors
- Network interface not found: Verify interface name in NodeNetworkState

## Repository

https://github.com/openshift/kubernetes-nmstate
