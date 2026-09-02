# Aldea

Aldea is a self-hosted, peer-to-peer distributed storage system. It allows small closed groups (friends, family, or small teams) to pool spare disk space across their personal devices into a single storage network without relying on third-party cloud providers or central servers.

Data stored in the network is chunked, erasure-coded, and encrypted client-side before being transmitted to peer storage nodes.

## Core Design

- **Zero-knowledge encryption**: Files are encrypted client-side using AES-256-GCM and erasure-coded (Reed-Solomon) before leaving the host machine. Storage nodes never receive unencrypted file contents.
- **Metadata separation**: The tracker daemon manages node availability and placement policy metadata only. It does not store or process raw data shards.
- **Churn tolerance**: The storage layer assumes consumer devices will join and leave the network unpredictably.
- **Closed groups**: Targeted at small node groups (5 to 20 nodes) sharing storage capacity within a known group.

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

- `cmd/aldea`: CLI client for uploading, retrieving, and inspecting stored files.
- `cmd/noded`: Storage node daemon that runs on host devices and stores encrypted data shards.
- `cmd/trackerd`: Coordinator daemon that tracks node health and maintains placement metadata.

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
│   ├── crypto/        # Client-side encryption
│   ├── dht/           # Peer discovery
│   ├── erasure/       # Reed-Solomon erasure coding
│   ├── nat/           # NAT traversal
│   ├── protocol/      # Wire protocols and serialization
│   └── tracker/       # Metadata engine and placement logic
├── pkg/
│   └── client/        # Go client library
├── docs/              # System documentation and specifications
└── test/
    └── integration/   # Multi-node integration test setups
```

## Development

Building from source requires Go 1.22 or higher.

```bash
# Build all binary targets
go build -o bin/aldea ./cmd/aldea
go build -o bin/noded ./cmd/noded
go build -o bin/trackerd ./cmd/trackerd
```

## License

MIT License
