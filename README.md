# Aldea

Aldea is a self-hosted, peer-to-peer distributed storage network. It allows small closed groups (friends, family, or small teams) to pool spare disk space across personal devices into a single storage network without relying on VPS hosting, third-party cloud providers, or central servers.

Data stored in the network is chunked, erasure-coded (Reed-Solomon), and encrypted client-side (XChaCha20-Poly1305) before leaving the host machine.

## Core Design

- **Zero-knowledge encryption**: Files are encrypted client-side using XChaCha20-Poly1305 with Argon2id key derivation and erasure-coded (Reed-Solomon $k=4, m=4$) before leaving the machine. Storage nodes never receive unencrypted content.
- **Metadata separation**: The tracker daemon manages node availability and placement policy metadata only. It does not store or process raw data shards.
- **Auto-repair and churn tolerance**: The storage layer handles node offline events by automatically reconstructing missing shards from remaining healthy nodes and re-allocating them to active peers.
- **Closed groups**: Designed for trusted small networks (5 to 20 nodes) sharing disk capacity.

## Architecture

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
```

### Components

- `cmd/aldea`: CLI client for initializing credentials, uploading, downloading, and inspecting network status.
- `cmd/noded`: Storage node daemon that runs on host devices and stores encrypted data shards locally (`bbolt`).
- `cmd/trackerd`: Coordinator daemon that tracks node health and maintains placement metadata (`bbolt`).

## Quickstart

### 1. Build Binaries

```bash
go build -o bin/aldea ./cmd/aldea
go build -o bin/noded ./cmd/noded
go build -o bin/trackerd ./cmd/trackerd
```

### 2. Local Multi-Node Test with Docker Compose

Spin up a local 8-node storage network + coordinator:

```bash
docker compose up -d --build
```

### 3. Usage via CLI

Initialize local client configuration:

```bash
./bin/aldea init --tracker "http://localhost:8080" --key "aldea-p2p-mesh-secret-key-32b!"
```

Upload a file:

```bash
./bin/aldea put /path/to/myphoto.jpg
# Output: uploaded: /path/to/myphoto.jpg → <fileID>
```

Check network status:

```bash
./bin/aldea status
```

Download file:

```bash
./bin/aldea get <fileID> /path/to/recovered_photo.jpg
```

### 4. Interactive Resiliency Demo

Run the automated demonstration script to test node failure resiliency ($k=4, m=4$ Reed-Solomon reconstruction):

```bash
./scripts/demo.sh
```

## Repository Layout

```
.
├── cmd/
│   ├── aldea/         # CLI client
│   ├── noded/         # Storage node daemon
│   └── trackerd/      # Coordinator daemon
├── internal/
│   ├── chunker/       # File splitting and reassembly
│   ├── config/        # Configuration loading and validation
│   ├── crypto/        # Client-side XChaCha20-Poly1305 & Argon2id
│   ├── erasure/       # Reed-Solomon erasure coding (klauspost/reedsolomon)
│   ├── protocol/      # Authenticated wire protocols (HMAC-SHA256)
│   ├── repair/        # Automatic shard reconstruction and re-allocation
│   └── tracker/       # Metadata engine and placement logic (bbolt)
├── deploy/
│   └── docker/        # Node and tracker deployment configs
├── test/
│   └── e2e/           # Multi-node end-to-end integration test suite
└── scripts/           # Demo and setup scripts
```

## Testing

Run all unit and integration tests with race detector:

```bash
go test -v -race ./...
```
