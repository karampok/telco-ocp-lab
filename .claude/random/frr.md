# FRR (Free Range Routing) vtysh Commands

## Overview

This skill provides FRR vtysh command reference for BGP, BFD, routing, and
interface configuration. For running these commands in Kubernetes/OpenShift
FRRK8s pods, see the **frrk8s skill**.

## FRR vtysh Commands Reference

### Global Configuration and Status

```bash
# Show complete running configuration
show running-config

# Show FRR version and build info
show version

# Show system uptime
show uptime

# Show configured daemons
show daemons

# Show memory usage
show memory

# Show threads
show thread cpu
```

### BGP Commands

```bash
# BGP summary (all peers)
show bgp summary

# BGP summary for specific address family
show bgp ipv4 unicast summary
show bgp ipv6 unicast summary

# All BGP neighbors detailed info
show bgp neighbors

# Specific neighbor detailed info
show bgp neighbors <peer-ip>

# BGP routing table
show bgp ipv4 unicast
show bgp ipv6 unicast

# Routes advertised to neighbor
show bgp ipv4 unicast neighbors <peer-ip> advertised-routes
show bgp ipv6 unicast neighbors <peer-ip> advertised-routes

# Routes received from neighbor
show bgp ipv4 unicast neighbors <peer-ip> received-routes
show bgp ipv6 unicast neighbors <peer-ip> received-routes

# Routes accepted after filtering
show bgp ipv4 unicast neighbors <peer-ip> routes

# BGP community information
show bgp community
show bgp ipv4 unicast community <community>
```

### BFD Commands

```bash
# Show all BFD peers
show bfd peers

# Brief BFD peer summary
show bfd peers brief

# Specific BFD peer
show bfd peer <peer-ip>

# BFD peer counters
show bfd peers counters
```

### Routing Table Commands

```bash
# Show IPv4 routing table
show ip route

# Show IPv6 routing table
show ipv6 route

# Show routes for specific protocol
show ip route bgp
show ip route connected
show ip route static

# Show specific route
show ip route <prefix>

# Show routing table summary
show ip route summary
```

### Interface Commands

```bash
# Show all interfaces
show interface

# Show specific interface
show interface <interface-name>

# Show interface brief
show interface brief

# Show IP addresses on interfaces
show ip interface
show ipv6 interface
```

### Prefix List and Route Map

```bash
# Show prefix lists
show ip prefix-list
show ip prefix-list <name>

# Show route maps
show route-map
show route-map <name>

# Show community lists
show bgp community-list
```

### Logging and Debug

```bash
# Show logging configuration
show logging

# Show BGP debugging status
show debugging bgp

# Enable BGP debugging (use with caution)
debug bgp updates
debug bgp neighbor-events
debug bgp keepalives

# Disable debugging
no debug all
```

### Statistics and Counters

```bash
# BGP statistics
show bgp statistics

# BGP message statistics for neighbor
show bgp neighbors <peer-ip> statistics

# Interface statistics
show interface <interface-name> statistics
```

## Resources

- FRR Documentation: https://docs.frrouting.org/
- FRR vtysh Reference: https://docs.frrouting.org/en/latest/vtysh.html
- FRR BGP Documentation: https://docs.frrouting.org/en/latest/bgp.html
- FRR BFD Documentation: https://docs.frrouting.org/en/latest/bfd.html
- **frrk8s skill** - For running these commands in Kubernetes/OpenShift
