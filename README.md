# Aldea

Aldea is an open-source, self-hosted, peer-to-peer (P2P) distributed storage and compute network designed for closed groups (friends, family, small teams). It enables users to pool spare disk space and compute capacity from personal devices across locations into a unified infrastructure without relying on paid VPS hosting, centralized cloud providers, or third-party servers.

---

## Core Features

### 1. Zero-Knowledge Distributed Storage
- **Client-Side Encryption**: High-security symmetric encryption using XChaCha20-Poly1305 with Argon2id key derivation. Data is encrypted locally before leaving the source host; storage nodes never see unencrypted content.
- **Redundancy & Fault Tolerance**: Reed-Solomon erasure coding ($K=4, M=4$). Every file is split into 8 shards (100% redundancy), allowing data recovery even if up to 4 storage nodes fail or go offline simultaneously.
- **Fixed Chunking**: Files are chunked into 1 MB blocks for uniform distribution across the P2P network.
- **Self-Healing Data Repair**: Automated liveness monitoring detects offline storage nodes (churn) and reconstructs missing shards onto active healthy nodes.

### 2. Distributed Compute Layer (MicroVMs & Containers)
- **Secure Isolation**: Execution of containerized workloads using Kata Containers and Firecracker microVMs (`internal/runtime`), providing kernel-level isolation on Linux hosts.
- **P2P Resource Scheduler**: Balanced distribution of compute instances across peer nodes based on available CPU and RAM capacity.
- **P2P Ingress Routing**: Public and private HTTP routing for running services with automated stateless failover.
- **Stateful Snapshot Volume Backups**: Automated periodic snapshots of persistent volumes stored securely in the P2P zero-knowledge storage pool.

### 3. P2P Networking & NAT Traversal
- **DHT Discovery & Routing**: Kademlia Distributed Hash Table (`internal/dht`) for peer discovery and node routing.
- **NAT & Router Traversal**: Integrated UPnP and STUN support for direct P2P connections across home network routers.
- **Relay Fallback**: Authenticated HMAC-SHA256 relay protocol (TURN/WebSocket) for restrictive symmetric NAT environments.
- **Bandwidth Control**: Per-node upload/download rate limiting and disk quota management.

### 4. User Interfaces
- **CLI (`aldea`)**: Command-line interface for credential setup, file transfers, and container deployment.
- **Desktop GUI (`aldea-desktop`)**: Cross-platform desktop application built with Wails v2 (Go + HTML/CSS/JS) featuring system tray integration (GTK/Windows/macOS), real-time metrics, and visual workload management.

---

## Architecture Overview

```
                       +-------------------+
                       |     trackerd      |
                       |   (Coordinator)   |
                       +---------+---------+
                                 | Metadata / Placement
                                 v
    +-------------------+ +-------------------+ +-------------------+
    |       noded       | |       noded       | |       noded       |
    |    (Storage A)    |<|    (Storage B)    |<|    (Storage C)    |
    +-------------------+ +-------------------+ +-------------------+
              ^                     ^                     ^
              |                     |                     |
              +---------------------+---------------------+
                        Encrypted P2P Mesh
```

### Component Breakdown
- `cmd/aldea`: Interactive CLI client.
- `cmd/noded`: Storage and compute node daemon running on host devices (embedded `bbolt` database).
- `cmd/trackerd`: Coordinator daemon managing node health and placement metadata (embedded `bbolt` database).
- `gui/`: Wails v2 desktop application with metrics monitoring, P2P file explorer, and microVM manager.

---

## Installation & Requirements

- **Go**: Version 1.22 or higher.
- **CGO & Build Toolchain**: `gcc`, `pkg-config`.
- **GUI Dependencies (Linux)**: `webkit2gtk-4.1` (or `webkit2gtk-4.0` on supported distros), `gtk3`.
- **Compute Engine (Optional for microVM nodes)**: Linux kernel with KVM / Kata Containers / Firecracker support.

---

## Quickstart Guide

### 1. Building Binaries

To compile CLI tools and daemons:

```bash
./scripts/build_all.sh
```

Generated executables will be placed in `build/bin/`:
- `build/bin/aldea`
- `build/bin/noded`
- `build/bin/trackerd`

To compile the native Wails desktop application:

```bash
./scripts/build_desktop.sh
```

The resulting executable will be `build/bin/aldea-desktop-linux-amd64` (or target platform equivalent).

---

### 2. Local Testing with Docker Compose

Spin up a local 8-node storage network + coordinator cluster:

```bash
docker compose up -d --build
```

Check cluster status:

```bash
docker compose ps
```

---

### 3. Command Line Interface (CLI) Usage

#### Initialize client credentials:
```bash
./build/bin/aldea init --tracker "http://localhost:8080" --key "secret-32-byte-encryption-key!"
```

#### Check network and active nodes status:
```bash
./build/bin/aldea status
```

#### Upload a file to the P2P network:
```bash
./build/bin/aldea put /path/to/my_image.png
```
*Output:* `uploaded: /path/to/my_image.png -> <fileID>`

#### Download a file from the network:
```bash
./build/bin/aldea get <fileID> /path/to/recovered_image.png
```

#### Deploy a compute container / microVM:
```bash
./build/bin/aldea compute deploy --image nginx:alpine --name my-server --cpus 1 --ram 512
```

#### List running compute workloads:
```bash
./build/bin/aldea compute list
```

#### Terminate a microVM:
```bash
./build/bin/aldea compute terminate <vmID>
```

---

### 4. Automated Resiliency Demo

Run the automated test script to demonstrate Reed-Solomon ($K=4, M=4$) reconstruction during simulated node failure:

```bash
./scripts/demo.sh
```

---

## Repository Structure

```
.
├── cmd/
│   ├── aldea/          # CLI client
│   ├── noded/          # Storage & compute node daemon
│   └── trackerd/       # Metadata coordinator daemon
├── deploy/
│   └── docker/         # Docker deployment configurations
├── gui/                # Wails v2 native desktop app (Go + Web Frontend)
├── internal/
│   ├── bandwidth/      # Quota management & bandwidth throttling
│   ├── chunker/        # 1 MB chunk splitting & reassembly
│   ├── config/         # Configuration loading & validation
│   ├── crypto/         # XChaCha20-Poly1305 & Argon2id client-side encryption
│   ├── dht/            # Kademlia Distributed Hash Table
│   ├── erasure/        # Reed-Solomon erasure coding (klauspost/reedsolomon)
│   ├── ingress/        # P2P HTTP ingress routing & tunneling
│   ├── invite/         # Invitation token generation & validation
│   ├── metrics/        # System & network performance metrics collector
│   ├── nat/            # NAT traversal (STUN & UPnP)
│   ├── protocol/       # Authenticated wire protocols (HMAC-SHA256)
│   ├── relay/          # TURN/WebSocket fallback relay
│   ├── repair/         # Automated shard repair engine
│   ├── runtime/        # Container/microVM runtime isolation (Kata/Firecracker)
│   ├── scheduler/      # Compute workload scheduler
│   ├── snapshot/       # Persistent volume snapshot & backup engine
│   └── tracker/        # Metadata placement & node health database (bbolt)
├── scripts/            # Build & demo scripts
└── test/               # Integration & E2E tests
```

---

## Testing

Run the full test suite with race condition detection:

```bash
go test -v -race ./...
```

---

## License

Distributed under the MIT License.
