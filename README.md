# Fringe

Fringe is a distributed synchronization platform designed for edge computing environments. It combines efficient data synchronization using Merkle trees with robust peer discovery and failure detection via SWIM-style gossip. This architecture enables lightweight, fault-tolerant, and eventually consistent data sharing across edge nodes with minimal coordination overhead.

---

## Features

- **Merkle Tree-Based Data Sync**  
  Efficient, partial synchronization of data using Merkle trees to detect and transmit only differences between nodes.

- **SWIM-Style Gossip Membership**  
  Scalable and resilient cluster membership using periodic gossip-based health checks and peer discovery.

- **Edge-First Architecture**  
  Designed to function in intermittently connected, bandwidth-constrained, and heterogeneous environments.

- **Pluggable Storage Engine**  
  Easily integrate your own data store or file system abstraction for synchronization.

---

## Architecture Overview

 +-----
| Node A | <----> | Node B | </br>
|------------------|
| SWIM Gossip Layer| <----> | SWIM Gossip Layer| </br>
|------------------|
| Merkle Tree Sync| <----> | Merkle Tree Sync| </br>
|------------------|
| Local Data | | Local Data |


- **Membership Layer:**  
  Periodic pings and indirect probes maintain a dynamic view of the cluster.

- **Sync Layer:**  
  On topology change or scheduled intervals, nodes compare Merkle tree hashes and exchange only the diffs.

---


## Use Cases

- Sensor data synchronization across edge gateways
- Peer-to-peer configuration propagation
- Offline-first applications with periodic sync
- Collaborative document or content delivery networks at the edge

---



