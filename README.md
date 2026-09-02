# Aldea

> Free, self-hosted, peer-to-peer distributed storage network.

Aldea lets a closed group of people (friends, family, a small team) pool spare disk space from devices they already own — across cities or countries — into a single unified storage pool, without renting a VPS, without paying subscription fees, and without any central server owning the data.

Turn devices you already own into your own shared storage pool for **$0**.

---

## 🌟 Core Principles

1. **Zero-Knowledge by Default**: Files are chunked, erasure-coded (Reed-Solomon), and encrypted client-side before leaving the originating machine. No other node can read your plaintext data.
2. **No Single Point of Failure**: The coordinator/tracker stores only placement metadata, never raw shard content. Losing the tracker never means losing your data.
3. **Designed for Unreliable Nodes**: Tolerates nodes joining and leaving the network unpredictably.
4. **Small Closed Groups**: Optimized for trusted small networks (5–20 nodes) operating under peer-to-peer encryption.
5. **100% Free & Self-Hosted**: No cloud lock-in, no paid tiers, no external dependencies.

---

## 🏗️ System Architecture

```
                               ┌─────────────────┐
                               │    trackerd     │
                               │  (Coordinator)  │
                               └────────┬────────┘
                                        │ Metadata / Placement
                                        ▼
 ┌─────────────────┐           ┌─────────────────┐           ┌─────────────────┐
 │     noded       │◄─────────►│     noded       │◄─────────►│     noded       │
 │   (Storage A)   │  P2P Mesh │   (Storage B)   │  P2P Mesh │   (Storage C)   │
 └─────────────────┘           └─────────────────┘           └─────────────────┘
```

### Components

- **`cmd/aldea`**: Command-line client for users to upload, retrieve, and manage stored data.
- **`cmd/noded`**: Storage node daemon running on participating devices to hold encrypted data shards.
- **`cmd/trackerd`**: Network coordinator daemon managing node health, shard placement policy, and group metadata.

---

## 📂 Repository Structure

```
.
├── cmd/
│   ├── aldea/         # User CLI client
│   ├── noded/         # Storage node daemon
│   └── trackerd/      # Network coordinator daemon
├── internal/
│   ├── chunker/       # File splitting and reassembly
│   ├── config/        # Configuration management
│   ├── crypto/        # Client-side encryption (AES-256-GCM)
│   ├── dht/           # Peer discovery & routing
│   ├── erasure/       # Reed-Solomon erasure coding
│   ├── nat/           # NAT traversal (STUN/TURN/ICE)
│   ├── protocol/      # P2P wire message protocols
│   └── tracker/       # Metadata store & placement policy
├── pkg/
│   └── client/        # Go client library
├── docs/              # Specifications, requirements, & architecture
└── test/
    └── integration/   # Multi-node Docker Compose tests
```

---

## 🚀 Quick Start (Development)

### Prerequisites
- **Go**: 1.22 or higher

### Building
```bash
go build -o bin/aldea ./cmd/aldea
go build -o bin/noded ./cmd/noded
go build -o bin/trackerd ./cmd/trackerd
```

---

## 📄 License

Open source & self-hosted under the MIT License.
