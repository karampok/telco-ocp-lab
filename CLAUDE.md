# Lab info

```
KUBECONFIG ~/.kube/lab0.yaml
DOCKER_HOST tcp://lab0:2375
```

Network topology is dynamic topo.clab.yaml. On start give me the topology
with OCP nodes and the external containerlab.

## Lab not working

```
 dig +short api.5gc.telco.vlab
 #ask user to sudo wg-quick up lab0
```


## Access external Gateways

```
#host level commands
DOCKER_HOST=tcp://lab0:2375 docker exec clab-vlab-gw1 bash -c 'ip a s'

#vtysh commands
DOCKER_HOST=tcp://lab0:2375 docker exec clab-vlab-gw1 vtysh -c 'show running-config'
```

## tcpdump in Lab Components

### OCP Nodes
Use oc debug with snife image (includes tcpdump).

Quick inline capture (text output):
```bash
oc debug node/w1 --image quay.io/karampok/snife:latest -- tcpdump -i bond0.11 -n -c 10
```

Capture to pcap file (use oc cp to retrieve):
```bash
# Start debug pod with --preserve-pod flag, capture packets, then sleep
oc debug node/w1 --image quay.io/karampok/snife:latest --preserve-pod -- \
  bash -c 'tcpdump -i bond0.11 -n -w /tmp/capture.pcap -c 20 && sleep 60' &

# Generate traffic (in another terminal or background)
# ... your test commands ...

# Find the debug pod name
oc get pods | grep w1-debug

# Copy pcap file from pod to local
oc cp w1-debug-xxxxx:/tmp/capture.pcap /tmp/w1-capture.pcap

# Analyze locally
tcpdump -r /tmp/w1-capture.pcap -n
wireshark /tmp/w1-capture.pcap

# Clean up
oc delete pod w1-debug-xxxxx
rm /tmp/w1-capture.pcap
```

### Docker Containers (Gateways, Clients)
To capture traffic in any docker container, use snife with shared network namespace:

```bash
# Generic pattern for any container
DOCKER_HOST=tcp://lab0:2375 docker run --rm --net=container:<container-name> \
  quay.io/karampok/snife:latest tcpdump -i <interface> -n -c 10

# Example: Capture on gateway gw1
DOCKER_HOST=tcp://lab0:2375 docker run --rm --net=container:clab-vlab-gw1 \
  quay.io/karampok/snife:latest tcpdump -i eth1.green -n host 5.5.5.1 -c 10

# Example: Capture on client
DOCKER_HOST=tcp://lab0:2375 docker run --rm --net=container:clab-vlab-green-client \
  quay.io/karampok/snife:latest tcpdump -i eth0 -n icmp -c 10
```

### Switch (sw1)
Switch is a Linux bridge on lab0 host. SSH to lab0 and run tcpdump:
```bash
# SSH to lab0 host
ssh lab0

# Run tcpdump on the bridge or its ports
sudo tcpdump -i sw1 -n -c 20
sudo tcpdump -i veth-w1 -n vlan -c 20
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

# Packet Path Visualization

## Network Diagram
```
┌──────────────┐            ┌───────────────────────────┐            ┌────────┐            ┌────────────────────────────────────────────────────┐            ┌─────────────────┐
│              │            │                           │            │        │            │                       w1                           │            │                 │
│ green-client │            │           gw1             │            │  sw1   │            │                                                    │            │    blue pod     │
│              │            │                           │            │ bridge │            │                                                    │            │                 │
│         eth0 │────────────│eth2              eth1.blue│────────────│        │────────────│bond0.10 ─── br-ex ─── ovn-k8s-mp0              veth│────────────│eth0             │
│              │            │                           │            │        │            │                                                    │            │                 │
└──────────────┘            └───────────────────────────┘            └────────┘            └────────────────────────────────────────────────────┘            └─────────────────┘
```

## Packet Flow Example (curl http://4.4.4.1:5555)

Forward Path:
```
200.200.200.10.45678 > 4.4.4.1.5555                          # green-client eth0 (egress)
200.200.200.10.45678 > 4.4.4.1.5555                          # gw1 eth2 (ingress)
200.200.200.10.45678 > 4.4.4.1.5555, vlan 10                 # gw1 eth1.blue (egress, VLAN 10 tagged)
200.200.200.10.45678 > 4.4.4.1.5555, vlan 10                 # sw1 bridge (forwarding)
200.200.200.10.45678 > 4.4.4.1.5555                          # w1 bond0.10 (ingress, VLAN 10)
200.200.200.10.45678 > 4.4.4.1.5555                          # w1 br-ex (OVN pre-DNAT)
200.200.200.10.45678 > 10.131.1.128.8080                     # w1 ovn-k8s-mp0 (post-DNAT to pod IP)
200.200.200.10.45678 > 10.131.1.128.8080                     # Pod eth0 (blue-xxx pod namespace)
```

# Setup MetalLB BGP on blue

BLUE stands for testing on the primary interface on the node.
Primary is where the default route exists and br-ex is attached


```bash
# Deploy blue peering config (BFDProfile, BGPPeer, BGPAdvertisement, IPAddressPool)
oc apply -f day2/blue-peering.yaml

# Deploy test app
oc apply -f day2/blue-pod-one.yaml
```

## Blue Configuration Example

```
BGP Peering Configuration (blue)
  Peer: 10.10.10.1 (ASN 65001)
  Local ASN: 7003 
  Participating Nodes: Workers only (node-role.kubernetes.io/worker)
  BFD Profile: simple (RX: 2s, TX: 1s, Multiplier: 3x)
  IP Pool: 4.4.4.1/32, 2001:db8:0:0:0:0:4:1/128 (autoAssign: false)
  Resources: BGPPeer/blue, BFDProfile/simple, BGPAdvertisement/blue, IPAddressPool/blue
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

# Comparing MetalLB Configurations

When asked to compare blue and green configurations from day2/ files, present in table format:

## BGP Peering Configuration

| Aspect | Blue | Green |
|--------|------|-------|
| **BGP Peer Address** | 10.10.10.1 | 11.11.11.1 |
| **Peer ASN** | 65001 | 8011 |
| **Local ASN** | 7003 | 7003 (same) |
| **BFD Profile Name** | simple | green |
| **Node Selector** | All workers | Only w1 |
| **Hold Time** | Not specified (default) | 60s |
| **Graceful Restart** | Disabled | Not specified (default) |
| **IP Pool** | 4.4.4.1/32, 2001:db8::4:1/128 | 5.5.5.1/32, 5.5.5.2/32, 2001:db8::5:1/128 |

## Pod Deployment Configuration

| Aspect | Blue | Green |
|--------|------|-------|
| **Node Selector** | Not specified | w0 |
| **Replicas** | 1 | 1 |
| **Anti-Affinity** | Not specified | Required |
| **Containers** | snife (sidecar), agnhost, iperf3 | snife, agnhost, iperf3 |
| **Number of SVCs** | 3 (all on 4.4.4.1) | 2 active (5.5.5.1, 5.5.5.2) |
| **SVC 1 Name** | blue-svc-http | green-svc-http-cluster |
| **SVC 1 IP** | 4.4.4.1:5555 | 5.5.5.1:5555 |
| **SVC 1 Traffic Policy** | Local | Cluster |
| **SVC 2 Name** | blue-svc-iperf-tcp | green-svc-http-local |
| **SVC 2 IP** | 4.4.4.1:60000/TCP | 5.5.5.2:5555 |
| **SVC 2 Traffic Policy** | Local | Local |
| **SVC 3 Name** | blue-svc-iperf-udp | N/A (commented out) |
| **SVC 3 IP** | 4.4.4.1:60000/UDP | N/A |
| **SVC 3 Traffic Policy** | Local | N/A |
