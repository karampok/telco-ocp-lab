# Lab info

```
KUBECONFIG ~/.kube/lab0.yaml
DOCKER_HOST tcp://10.1.104.10:2375
```

Network topology is dynamic topo.clab.yaml. On start give me the topology
with OCP nodes and the external containerlab.


## Access external Gateways

```
#host level commands
DOCKER_HOST=tcp://10.1.104.10:2375 docker exec clab-vlab-gw1 bash -c 'ip a s'

#vtysh commands
DOCKER_HOST=tcp://10.1.104.10:2375 docker exec clab-vlab-gw1 vtysh -c 'show running-config'
```

## Testing agnhost netexec endpoints

For agnhost netexec (registry.k8s.io/e2e-test-images/agnhost:2.40), use simple
curl commands without URL encoding:

```bash
# Simple command - no URL encoding needed
curl -s http://4.4.4.1:5555/shell?cmd="env|grep -i node" | jq -r '.output'

# For complex commands with special characters, use URL encoding alias
alias urlencode="python3 -c \"import sys, urllib.parse; print(urllib.parse.quote(''.join(sys.stdin.readlines())))\""
curl -s http://4.4.4.1:5555/shell?cmd="$(echo "your complex command" | urlencode)" | jq -r '.output'
```

# Testing Blue MetalLB BGP Peering

## Blue Configuration
- **Peer**: 10.10.10.1 (ASN 65001)
- **Local ASN**: 7003
- **BFD**: Enabled (2s RX, 1s TX, 3x multiplier)
- **IP Pool**: 4.4.4.1/32, 2001:db8:0:0:0:0:4:1/128 (autoAssign: false)
- **Nodes**: Workers only

## Quick Setup

```bash
# Deploy blue peering config (BFDProfile, BGPPeer, BGPAdvertisement, IPAddressPool)
oc apply -f day2/blue-peering.yaml

# Deploy test app
oc apply -f day2/blue-pod-one.yaml
```

## Blue-Specific Checks

```bash
# Verify blue resources are created
oc get bgppeer,bfdprofile,bgpadvertisement,ipaddresspool -n metallb-system | grep blue

# Check blue BGP session state
oc get bgpsessionstates -n openshift-frr-k8s -o yaml | grep -A10 "10.10.10.1"

# Check blue service status
oc get servicebgpstatuses -n metallb-system -o yaml | grep -A20 "blue"

# List blue services and their IPs
oc get svc -n default -l 'app in (blue,blue-two)' -o wide

# Test blue HTTP service
curl -s http://4.4.4.1:5555/shell?cmd="env|grep -i node" | jq -r '.output'
```

## Test Applications

### blue-pod-one.yaml
Single replica deployment with three services sharing 4.4.4.1:
- `blue-svc-http`: 5555 → 8080 (agnhost netexec)
- `blue-svc-iperf-tcp`: 60000/TCP (iperf3)
- `blue-svc-iperf-udp`: 60000/UDP (iperf3)

Containers: snife (privileged sidecar), agnhost (HTTP), iperf3 (performance)
Traffic policy: Local
