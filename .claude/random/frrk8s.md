# FRRK8s (FRRouting for Kubernetes)

Namespace: `openshift-frr-k8s`

## Overview

FRRK8s is deployed automatically by CNO (Cluster Network Operator) when a
MetalLB instance is created. It provides BGP routing capabilities via the FRR
routing daemon running in Kubernetes pods.

For FRR vtysh commands and BGP/BFD configuration details, see the **frr skill**.

## FRRK8s Architecture

Each FRRK8s pod runs as a daemonset on worker nodes with 7 containers:
- `controller` - FRRK8s controller managing FRR configurations
- `frr` - FRRouting daemon for BGP routing
- `reloader` - watches for configuration changes
- `frr-metrics` - exposes FRR metrics on port 7573
- `frr-status` - monitors FRR status
- `kube-rbac-proxy` - secures metrics endpoints
- `kube-rbac-proxy-frr` - secures FRR metrics endpoints

## Check FRRK8s Status

```bash
# List FRRK8s pods
oc get pods -n openshift-frr-k8s

# Check daemonset
oc get daemonset -n openshift-frr-k8s

# Detailed pod information
oc get pods -n openshift-frr-k8s -o wide

# Wait for pods to be ready
oc wait --for=condition=Ready pods --all -n openshift-frr-k8s --timeout=120s

# Get node mapping
oc get pods -n openshift-frr-k8s -l app=frr-k8s \
  -o custom-columns=POD:.metadata.name,NODE:.spec.nodeName
```

## Version Information

```bash
# Get FRR commit from OpenShift release
oc adm release info --commits | grep frr

# Check FRR RPM version
FRR_POD=$(oc get pods -n openshift-frr-k8s -o jsonpath='{.items[0].metadata.name}')
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- rpm -q frr

# Detailed RPM info
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- rpm -qi frr
```

## frrk8s.metallb.io/v1beta1
```bash
oc api-resources --api-group=frrk8s.metallb.io
```

Resources:
- `bgpsessionstates` - BGP session state information
- `frrconfigurations` - FRR routing daemon configurations
- `frrnodestates` (cluster-scoped) - FRR state per node



## Running FRR Commands in FRRK8s

### Access FRR Container

```bash
# Get first FRR pod
FRR_POD=$(oc get pods -n openshift-frr-k8s -o jsonpath='{.items[0].metadata.name}')

# Get FRR pod for specific node
FRR_POD=$(oc get pods -n openshift-frr-k8s -l frrk8s.metallb.io/node=<node-name> \
  -o jsonpath='{.items[0].metadata.name}')

# Get bash shell in FRR container
oc exec -it -n openshift-frr-k8s "$FRR_POD" -c frr -- bash

# Access FRR vtysh (interactive shell)
oc exec -it -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh
```

### Run Single vtysh Commands

```bash
# Single command execution
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "<command>"

# Multiple commands
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "<command1>" -c "<command2>"

# Common examples (see frr skill for complete vtysh command reference)
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "show running-config"
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "show bgp summary"
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "show bgp neighbors"
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "show bfd peers"
```

For complete list of vtysh commands, see the **frr skill**.

### Access FRR Configuration Files

```bash
# View FRR daemon config
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- cat /etc/frr/frr.conf

# View daemons file
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- cat /etc/frr/daemons

# List FRR config directory
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- ls -la /etc/frr/
```

### Run Commands Across All FRR Pods

```bash
# Execute command on all FRR pods
for pod in $(oc get pods -n openshift-frr-k8s -l app=frr-k8s \
  -o jsonpath='{.items[*].metadata.name}'); do
  echo "=== $pod ==="
  oc exec -n openshift-frr-k8s "$pod" -c frr -- vtysh -c "show bgp summary"
done
```

## FRRK8s Kubernetes Resources

### BGP Session State

```bash
# List BGP session states for all nodes
oc get bgpsessionstates -n openshift-frr-k8s

# Detailed view
oc get bgpsessionstates -n openshift-frr-k8s -o yaml

# Describe specific session
oc describe bgpsessionstates -n openshift-frr-k8s

# Check sessions for specific peer
oc get bgpsessionstates -n openshift-frr-k8s -l frrk8s.metallb.io/peer=<peer-ip>

# Check sessions for specific node
oc get bgpsessionstates -n openshift-frr-k8s -l frrk8s.metallb.io/node=<node-name>
```

### FRR Configuration

```bash
# List FRR configurations
oc get frrconfigurations -n openshift-frr-k8s

# View configuration details
oc get frrconfigurations -n openshift-frr-k8s -o yaml

# Describe configuration
oc describe frrconfigurations -n openshift-frr-k8s
```

### FRR Node States

```bash
# List FRR node states (cluster-scoped resource)
oc get frrnodestates

# Detailed view
oc get frrnodestates -o yaml

# Describe specific node state
oc describe frrnodestates <node-name>
```

## FRR Metrics

```bash
# Get first FRR pod
FRR_POD=$(oc get pods -n openshift-frr-k8s -o jsonpath='{.items[0].metadata.name}')

# Fetch all FRR metrics
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr-metrics -- \
  wget -qO- http://127.0.0.1:7573/metrics

# Filter BGP metrics
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr-metrics -- \
  wget -qO- http://127.0.0.1:7573/metrics | grep -i bgp

# Filter BFD metrics
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr-metrics -- \
  wget -qO- http://127.0.0.1:7573/metrics | grep -i bfd

# Combined BGP and BFD metrics
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr-metrics -- \
  wget -qO- http://127.0.0.1:7573/metrics | grep -E 'bgp|bfd'
```

Key metrics (see **frr skill** for complete metrics reference):
- `frr_bgp_peer_state` - BGP peer connection state
- `frr_bgp_peer_prefixes_received` - Prefixes received
- `frr_bgp_peer_prefixes_sent` - Prefixes sent
- `frr_bfd_peer_state` - BFD peer state

## Logs

```bash
# Controller logs
oc logs -n openshift-frr-k8s -l app=frr-k8s -c controller --tail=50

# FRR daemon logs
oc logs -n openshift-frr-k8s -l app=frr-k8s -c frr --tail=50

# Reloader logs
oc logs -n openshift-frr-k8s -l app=frr-k8s -c reloader --tail=50

# FRR metrics logs
oc logs -n openshift-frr-k8s -l app=frr-k8s -c frr-metrics --tail=50

# Follow logs in real-time
oc logs -n openshift-frr-k8s -l app=frr-k8s -c frr -f

# Logs from specific pod
FRR_POD=$(oc get pods -n openshift-frr-k8s -o jsonpath='{.items[0].metadata.name}')
oc logs -n openshift-frr-k8s "$FRR_POD" -c frr --tail=100

# All container logs from a pod
oc logs -n openshift-frr-k8s "$FRR_POD" --all-containers --tail=50
```

## Advanced: Custom FRR Configuration

Use `FRRConfiguration` CR to apply custom FRR routing configurations beyond
what MetalLB BGPPeer provides.

### FRRConfiguration Resource

```bash
# List custom FRR configurations
oc get frrconfigurations -n openshift-frr-k8s

# View configuration
oc get frrconfigurations -n openshift-frr-k8s -o yaml

# Apply custom FRR config
oc apply -f <frrconfig.yaml>
```

### Common Use Cases

**BGP Unnumbered Peering**
- Point-to-point BGP sessions without interface IP addresses
- Requires NMState configuration for P2P interfaces
- May need custom routing and proxy ARP configuration

**Route Learning and Redistribution**
- Learn routes from BGP peers beyond MetalLB-managed routes
- Configure static routes or route maps
- Handle source IP selection for outbound traffic

**Advanced BGP Features**
- Route filtering with prefix lists
- Community tags and route maps
- BGP policy configuration
- Multi-hop BGP sessions

See **frr skill** for detailed FRR configuration and vtysh commands.

### Verification

```bash
# Check interface configuration on node
oc debug node/<worker-node> -- ip a s <interface>

# Verify BGP sessions (see frr skill for vtysh commands)
FRR_POD=$(oc get pods -n openshift-frr-k8s -l frrk8s.metallb.io/node=<node> \
  -o jsonpath='{.items[0].metadata.name}')
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "show bgp summary"

# Check learned routes
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "show ip route"

# Verify route on node
oc debug node/<worker-node> -- ip route get <destination-ip>

# Check running FRR configuration
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "show running-config"
```

## Troubleshooting

### BGP Session Not Established

```bash
# Check BGP session states
oc get bgpsessionstates -n openshift-frr-k8s -o yaml

# Check FRR logs for errors
oc logs -n openshift-frr-k8s -l app=frr-k8s -c frr | grep -i error

# Check BGP neighbors (see frr skill for complete vtysh reference)
FRR_POD=$(oc get pods -n openshift-frr-k8s -o jsonpath='{.items[0].metadata.name}')
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "show bgp neighbors"

# Verify running config
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "show running-config"
```

### BFD Session Issues

```bash
# Check BFD peers
FRR_POD=$(oc get pods -n openshift-frr-k8s -o jsonpath='{.items[0].metadata.name}')
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- vtysh -c "show bfd peers"

# Check BFD metrics
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr-metrics -- \
  wget -qO- http://127.0.0.1:7573/metrics | grep bfd

# Verify BFD in BGP neighbor config
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- \
  vtysh -c "show running-config" | grep -A20 "neighbor.*bfd"
```

### Configuration Not Applied

```bash
# Check FRR configuration resource
oc get frrconfigurations -n openshift-frr-k8s -o yaml

# Check controller logs
oc logs -n openshift-frr-k8s -l app=frr-k8s -c controller --tail=100

# Check reloader logs
oc logs -n openshift-frr-k8s -l app=frr-k8s -c reloader --tail=100

# Verify FRR daemon config
FRR_POD=$(oc get pods -n openshift-frr-k8s -o jsonpath='{.items[0].metadata.name}')
oc exec -n openshift-frr-k8s "$FRR_POD" -c frr -- cat /etc/frr/frr.conf
```

## Debug Mode

```bash
# Enable debug logging via ConfigMap
cat <<EOF | oc apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: env-overrides
  namespace: openshift-frr-k8s
data:
  frrk8s-loglevel: "--log-level=debug"
EOF

# Restart FRRK8s pods to apply debug logging
oc delete pods -n openshift-frr-k8s --all

# Wait for pods to restart
oc wait --for=condition=Ready pods --all -n openshift-frr-k8s --timeout=120s
```

## Resources

- FRRK8s GitHub: https://github.com/metallb/frr-k8s
- **frr skill** - For FRR vtysh commands and BGP/BFD configuration
- **metallb skill** - For MetalLB integration and configuration
