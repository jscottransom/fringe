# Fringe

Fringe is a **distributed synchronization platform** purpose-built for edge computing environments.  
It combines efficient Merkle tree–based data synchronization** with robust SWIM-style gossip membership and failure detection.  
This architecture enables lightweight, fault-tolerant, and eventually consistent data sharing** across edge nodes with minimal coordination overhead**.

---

## ✨ Features

- **Merkle Tree–Based Data Sync**  
  Efficient partial synchronization that transmits only differences between nodes.

- **SWIM-Style Gossip Membership**  
  Scalable, resilient cluster membership with periodic health checks and peer discovery.

- **Edge-First Architecture**  
  Designed for intermittently connected, bandwidth-constrained, and heterogeneous environments.

- **Pluggable Storage Engine**  
  Easily integrate your own data store or file system abstraction.

---

## Architecture Overview

![image](arch.png)



**Components:**

- **Membership Layer**  
  Periodically pings peers and uses indirect probes to maintain a dynamic view of the cluster.

- **Sync Layer**  
  On topology change or scheduled intervals, nodes compare Merkle tree hashes and exchange only the diffs.

---

## 📚 Use Cases

- Synchronizing sensor data across edge gateways.
- Peer-to-peer configuration propagation.
- Offline-first applications with periodic synchronization.

