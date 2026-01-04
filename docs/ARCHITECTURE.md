# Architecture Overview

## System Design

InterMesh is built on a decentralized mesh networking architecture that allows devices to share internet connectivity across a global network.

### Core Components

#### 1. Node
- Represents a device in the mesh network
- Manages peer connections
- Tracks internet connectivity status
- Stores peer information

#### 2. Mesh Manager
- Initializes and manages the mesh network
- Handles peer discovery
- Manages network lifecycle

#### 3. Router
- Routes packets through the mesh network
- Maintains routing table with routes and costs
- Selects best next hop based on route metrics

#### 4. Proxy Manager
- Manages proxy connections for internet sharing
- Tracks active proxy connections
- Selects best available proxy
- Manages proxy statistics

#### 5. Personal Network Manager
- Creates and manages sub-mesh groups
- Defines network policies
- Controls access and resource sharing
- Manages network membership

### Network Topology

```
        ┌─────────────┐
        │   Node A    │ (Internet)
        │  (Proxy)    │
        └──────┬──────┘
               │
        ┌──────┼──────┐
        │      │      │
    ┌───▼──┐ ┌──▼─┐ ┌──▼─┐
    │Node B│ │Node C│ │Node D│
    │(WiFi)│ │(WiFi)│ │(WiFi)│
    └──────┘ └──────┘ └──────┘
```

### Data Flow

1. **Peer Discovery**: Nodes broadcast presence on local network
2. **Routing**: Nodes learn routes to other nodes via exchanges
3. **Proxy Selection**: Non-internet nodes select best proxy for routing
4. **Packet Forwarding**: Packets routed through mesh to destination
5. **Internet Access**: Proxy nodes forward internet traffic

### Personal Networks

Personal networks (sub-mesh) provide:
- **Device Grouping**: Organize devices into logical groups
- **Policy Management**: Define access and sharing policies
- **Member Management**: Add/remove members
- **Proxy Selection**: Choose proxy for the group

## Communication Protocol

### Peer Discovery
- Broadcast: `HELLO <NodeID> <IP> <MAC> <HasInternet>`
- Response: `ACK <NodeID> <IP>`

### Route Advertisement
- Message: `ROUTE <Destination> <NextHop> <HopCount> <Cost>`

### Data Packets
- Header: `<SourceID> <DestinationID> <ProxyID> <TTL>`
- Payload: Application data

## Security Considerations

- **Authentication**: Nodes must authenticate before joining mesh
- **Encryption**: Data should be encrypted in transit
- **Policy Enforcement**: Personal networks enforce access policies
- **Rate Limiting**: Prevent proxy abuse

## Performance Optimization

- **Route Caching**: Cache frequently used routes
- **Proxy Load Balancing**: Distribute connections across multiple proxies
- **Signal Strength**: Use RSSI for route selection
- **Connection Pooling**: Reuse proxy connections
