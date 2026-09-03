# 0001. Symmetric cipher selection: ChaCha20-Poly1305

## Status

Accepted

## Context

The tech stack document lists two options for shard encryption: `golang.org/x/crypto/chacha20poly1305` and stdlib `crypto/aes` (AES-256-GCM). A decision was needed before implementing `internal/crypto`.

Aldea targets consumer hardware across Linux, macOS, and Windows (RNF-12), with no assumption about CPU capabilities. The node pool includes personal computers ranging from modern desktops to older or low-end machines. AES-GCM performance is only competitive when the CPU provides hardware acceleration (AES-NI); on hardware without it, AES-GCM falls back to a slower software implementation vulnerable to cache-timing attacks. ChaCha20-Poly1305 delivers predictable, constant-time performance across all target architectures regardless of hardware support.

## Decision

Use **ChaCha20-Poly1305** (`golang.org/x/crypto/chacha20poly1305`) as the sole symmetric cipher for shard encryption throughout Aldea.

Key derivation remains Argon2id (`golang.org/x/crypto/argon2`) as specified in RNF-02, producing a 32-byte key suitable for `chacha20poly1305.New`.

A random 24-byte nonce (XChaCha20-Poly1305 variant) is generated per shard using `crypto/rand`. The nonce is prepended to the ciphertext so each shard is a self-contained authenticated blob.

## Consequences

- Consistent, portable encryption performance on all consumer hardware without relying on CPU feature detection.
- Slight throughput disadvantage on machines with AES-NI compared to AES-256-GCM, accepted given the target environment.
- The codebase imports `golang.org/x/crypto` regardless (Argon2id already requires it), so no new dependency is introduced.
- Changing cipher in the future would require a migration plan for stored shards; this decision should be revisited before Phase 3 if benchmarks (RNF-08) show a meaningful gap.
