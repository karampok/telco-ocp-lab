---
description: Verify IP forwarding configuration on OpenShift nodes
---

Verify IP forwarding settings on OpenShift cluster nodes, comparing global and
per-interface forwarding configurations. Check both sysctl settings and nmstate
configurations.

# Task

As a user of a multi-interface node (OCP), I want to configure a secondary
interface to allow forwarding via sysctl (`net.ipv4.conf.<interface>.forwarding=1`)
in order for MetalLB load balancers to work.

**Scenario:**
- Given: Global forwarding is disabled and per-interface forwarding for <interface> is not set
- When: I apply the nmstate config which includes the `forwarding: true` field in the ipv4 section
- Then: I observe that `net.ipv4.conf.<interface>.forwarding = 1` is set

# nmstate Forwarding API

nmstate supports configuring IP forwarding per interface using the `forwarding` field in the ipv4/ipv6 configuration section.

**Syntax:**
```yaml
interfaces:
  - name: <interface-name>
    type: <interface-type>
    state: up
    ipv4:
      enabled: true
      dhcp: false
      forwarding: true  # Enable IPv4 forwarding for this interface
    ipv6:
      enabled: true
      dhcp: false
      forwarding: true  # Enable IPv6 forwarding for this interface
```

**Requirements:**
- nmstate version >= 2.2.54 in handler pods
- OCP 4.21+ typically includes compatible versions

**Effect:**
Setting `ipv4.forwarding: true` configures the sysctl parameter:
- `net.ipv4.conf.<interface>.forwarding = 1`

This enables the interface to forward IPv4 packets, allowing traffic to be processed by iptables forwarding chains.

# Verification Workflow

1. Verify cluster connectivity and health
2. Check nmstate operator is running
3. Verify nmstate handler version >= 2.2.54
4. Ask user for interface name (must be secondary, not primary with default route)
5. Verify interface exists in nodenetworkstate
6. Check global forwarding is disabled in network.operator
7. Check global forwarding is disabled on nodes
8. Check per-interface forwarding state
9. Observe packet drops due to disabled forwarding
10. Check existing NNCP configurations
11. Generate report with current state and next steps

# Pre-checks

## 1. Verify cluster connectivity

```bash
oc whoami --show-server
oc get nodes -o wide
```

All nodes should be Ready and schedulable.

## 2. Check nmstate operator

```bash
oc get pods -n openshift-nmstate -l component=kubernetes-nmstate-handler
```

All handler pods should be Running.

## 3. Verify nmstate handler version

```bash
HANDLER_POD=$(oc get pods -n openshift-nmstate \
  -l component=kubernetes-nmstate-handler -o jsonpath='{.items[0].metadata.name}')

oc exec -n openshift-nmstate $HANDLER_POD -c nmstate-handler -- nmstatectl version
```

Version must be >= 2.2.54 for `forwarding` field support.

## 4. Ask for interface and verify it exists

Ask user for interface name (e.g., bond0.11, bond0.12).
Cannot use the primary interface (where br-ex is attached and default route exists).

Verify interface exists:
```bash
oc get nodenetworkstate w1 -o jsonpath='{.status.currentState.interfaces[?(@.name=="bond0.11")]}' | jq .
```

## 5. Verify global forwarding is disabled

Check cluster-level configuration:
```bash
oc get network.operator -o yaml | grep -A 5 "routingViaHost"
```

Expected: `routingViaHost: false`

## 6. Check global IP forwarding on nodes

```bash
# Check on worker nodes
oc debug node/w0 -- sysctl net.ipv4.ip_forward 2>&1 | grep -E "^net\."
oc debug node/w1 -- sysctl net.ipv4.ip_forward 2>&1 | grep -E "^net\."
```

Expected: `net.ipv4.ip_forward = 0`

## 7. Check per-interface forwarding state

```bash
# Replace / with . for sysctl (e.g., bond0.11 -> bond0/11)
oc debug node/w0 -- sysctl net.ipv4.conf.bond0/11.forwarding 2>&1 | grep -E "^net\."
oc debug node/w1 -- sysctl net.ipv4.conf.bond0/11.forwarding 2>&1 | grep -E "^net\."
```

Expected: `net.ipv4.conf.bond0/11.forwarding = 0`

## 8. Observe packet drops due to disabled forwarding

When per-interface forwarding is disabled, packets destined for non-local addresses
arriving on that interface will be dropped at the IP layer. You can observe this
using SNMP counters.

### Check IP forwarding statistics

Use `nstat` to view IP forwarding statistics. `nstat` is the modern iproute2 replacement
for `netstat -s`, providing cleaner output with raw counter names directly from the kernel.

```bash
# Show all IP counters with absolute values using nstat
oc debug node/w1 -- nstat -asz 2>&1 | grep -E '^Ip'
```

**Key metrics to check:**

```
IpInReceives                    21938391           0.0
IpInAddrErrors                  27                 0.0  ← Packets for non-local addresses
IpForwDatagrams                 0                  0.0  ← Should be 0 if forwarding disabled
IpInDiscards                    0                  0.0
IpInDelivers                    21938346           0.0
IpOutRequests                   21687825           0.0
```

### Check interface-level statistics

```bash
# Interface statistics (should show 0 dropped at NIC level)
oc debug node/w1 -- ip -s link show bond0.11 2>&1 | grep -A10 "bond0.11"
```

**Expected output:**
```
RX:  bytes packets errors dropped  missed   mcast
       XXX    YYYY      0       0       0     ZZ
TX:  bytes packets errors dropped carrier collsns
       XXX    YYYY      0       0       0       0
```

### Understanding the counters

According to Linux kernel documentation:

- **IpForwDatagrams**: Number of packets forwarded through this node
  - Should be 0 when per-interface forwarding is disabled
- **IpInAddrErrors**: Packets with invalid addresses
  - Increments when packets arrive for non-local addresses and forwarding is disabled
- **RX dropped**: Packets dropped at interface/NIC level
  - Should be 0 (drops happen at IP layer, not interface layer)

**Tools:**
- `nstat`: Modern iproute2 tool for viewing network statistics (recommended)
- Alternative: Read raw counters from `/proc/net/snmp`

**Reference:** https://docs.kernel.org/networking/snmp_counter.html

### Evidence of forwarding issue

If you see:
- `IpForwDatagrams: 0` (no packets forwarded)
- `IpInAddrErrors: N` where N > 0 (packets for non-local addresses dropped)
- Interface RX dropped: 0 (not dropped at NIC level)

This confirms packets are reaching the interface but cannot be forwarded because
`net.ipv4.conf.<interface>.forwarding = 0`.

## 10. Check NodeNetworkConfigurationPolicy

```bash
oc get nodenetworkconfigurationpolicy
```

If NNCPs exist, check for forwarding settings:
```bash
oc get nodenetworkconfigurationpolicy <policy-name> -o yaml | grep -A 5 "forwarding"
```

# Enabling Forwarding

To enable IP forwarding on a secondary interface, create an NNCP:

```yaml
apiVersion: nmstate.io/v1
kind: NodeNetworkConfigurationPolicy
metadata:
  name: bond0-11-forwarding
spec:
  nodeSelector:
    node-role.kubernetes.io/worker: ""
  desiredState:
    interfaces:
      - name: bond0.11
        type: vlan
        state: up
        ipv4:
          enabled: true
          dhcp: false
          forwarding: true
```

Apply and verify:
```bash
oc apply -f nncp-forwarding.yaml
oc get nodenetworkconfigurationenactment
oc debug node/w1 -- sysctl net.ipv4.conf.bond0/11.forwarding
```

Expected result: `net.ipv4.conf.bond0/11.forwarding = 1`

**IMPORTANT**  Never suggest workaround this is the only way to work.


# Common Issues

## Issue: NNCP forwarding field not supported

**Symptom:**
```
NmstateError: InvalidArgument: unknown field `forwarding`
```

**Cause:** nmstate handler version < 2.2.54

**Solution:**
1. Verify handler version: `oc exec -n openshift-nmstate <handler-pod> -c nmstate-handler -- nmstatectl version`
2. Upgrade OCP cluster to 4.21+ to get newer handler pods
3. Alternative: Use MachineConfig to set sysctl directly (workaround)

## Issue: Interface doesn't exist

**Symptom:** `error: key path "status.currentState.interfaces[?(@.name=="<interface>")]" not found`

**Cause:** Interface name is incorrect or doesn't exist on the node

**Solution:**
1. List all interfaces: `oc get nodenetworkstate <node> -o jsonpath='{.status.currentState.interfaces[*].name}'`
2. Verify interface naming (e.g., bond0.11 uses VLAN ID 11)
3. Check if interface is configured in node network configuration
