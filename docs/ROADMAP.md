# Roadmap

Status legend: ⬜ not started · 🟨 in progress · ✅ done

This roadmap implements the requirements defined in
[`docs/requirements/`](requirements/) (functional requirements `RF-XX`,
non-functional requirements `RNF-XX`, use cases `CU-XX`). Each item below
references the requirement(s) it satisfies — if a task doesn't trace back
to a requirement ID, either the task doesn't belong in this phase or the
requirements docs need to be updated first.

## Phase 1 — Storage core (current phase)

Goal: a working, self-hosted, encrypted, erasure-coded storage pool across
a handful of nodes on the **same network** (Docker Compose / LAN). NAT
traversal across real different networks is explicitly deferred to Phase 2.

Implements: RF-01, RF-02, RF-03, RF-05, RF-08 to RF-14, RF-16 to RF-18,
RNF-01, RNF-02, RNF-04, RNF-05, RNF-08, RNF-15 — see
[`use-cases.md`](requirements/use-cases.md) CU-01 to CU-04, CU-06, CU-07.

- ✅ `internal/chunker` — split/reassemble files into fixed-size blocks (RF-08)
- ✅ `internal/erasure` — Reed-Solomon wrapper, encode/decode/reconstruct (RF-09, RNF-05)
- ✅ `internal/crypto` — XChaCha20-Poly1305 encryption + Argon2id key derivation (RF-10, RNF-01, RNF-02)
- ✅ `internal/protocol` — message definitions for put/get/health between nodes (RNF-04)
- ✅ `internal/dht` — peer discovery, reuse/adapt existing Kademlia implementation (RF-05)
- ✅ `noded` — node agent: local shard storage (bbolt-backed), serves put/get (RF-11, RF-12, RF-14)
- ✅ `trackerd` — coordinator: metadata store, basic placement policy (RF-11, RF-13, RF-16)
- ✅ `aldea` CLI — `network create/join`, `alloc`, `put`, `get`, `status` (RF-01, RF-02, RF-03, RF-16, RF-17)
- ✅ Docker Compose topology simulating 8 nodes for local testing
- ✅ Basic repair job: detect a node going offline, reconstruct its shards elsewhere (RF-13, CU-06)
- ✅ README with setup guide and demo script (Docker Compose demo: kill a node, show file is retrievable)

**Definition of done for Phase 1:** CU-01 through CU-04, CU-06, and CU-07
are demonstrable end to end via `docker compose up`, including killing a
node mid-download and confirming the file is still retrievable (RNF-05).

## Phase 2 — Real-world networking

Goal: the same system working across actual separate networks/countries,
not just Docker Compose.

Implements: RF-06, RF-07, RF-15, RNF-03.

- ✅ NAT traversal: STUN-based hole punching (`internal/nat`, via `pion/stun`) (RF-06)
- ✅ Peer-relayed fallback when direct/hole-punched connection fails (RF-07)
- ✅ Node liveness heuristics tuned for intermittent home connections
  (distinguish "restarting router" from "gone for good")
- ⬜ Real cross-region test using free-tier cloud VMs (Oracle Cloud ARM,
  etc.) in at least 2 different regions
- ✅ Bandwidth throttling / allocation enforcement (respect the
  `--bandwidth` limit a node sets) (RF-15)
- ✅ Invite/join flow hardened (secure token exchange, not just a shared
  passphrase) (RNF-03)

**Definition of done for Phase 2:** CU-02 succeeds between two nodes on
genuinely different networks/ISPs/countries, with automatic fallback to a
relay when direct connection fails.

## Phase 3 — Desktop UI/UX

Goal: a simple, intuitive desktop app so a non-technical friend/family
member can join a network, see their storage, and upload/download files
without ever touching the CLI or a config file.

Implements: RF-19 to RF-22, RNF-10, RNF-12. See CU-08.

- ⬜ Desktop app shell using **Wails** (Go backend, calls `internal/` and
  `pkg/client` directly — no duplicated logic in a second language)
- ⬜ Frontend framework decision (Svelte recommended for a small,
  fast bundle; React acceptable if faster to build given prior experience)
  — record the choice in an ADR
- ⬜ Core flows: create/join network (via invite link or QR-style code,
  not a raw passphrase paste), set storage allocation with a slider, view
  connected nodes and their health at a glance, drag-and-drop
  upload/download
- ⬜ System tray integration: `noded` runs in the background, app shows
  status (online, syncing, degraded) without needing the window open
- ⬜ First-run onboarding flow that explains the model in plain language
  (what erasure coding guarantees, in non-technical terms) — this matters
  because trust is the actual product being sold here
- ⬜ Packaged installers per OS (macOS `.dmg`, Windows `.msi`, Linux
  `.AppImage`/`.deb`), signed if budget allows once there's revenue to
  justify a code-signing cert

**Definition of done for Phase 3:** someone with no CLI experience can
install the app, join a network via an invite link, and successfully
upload/download a file, entirely through the GUI.

## Phase 4 — Usability & distribution

Goal: harden distribution so the project is genuinely "clone, build, run" —
this is the phase that turns Aldea from "a repo" into "a thing anyone can
actually set up with their friends/family this weekend."

- ⬜ Single-binary installers / prebuilt releases (GitHub Releases, per-OS)
  for the CLI and headless `noded`/`trackerd`, complementing the Phase 3
  desktop installers
- ⬜ Config wizard (`aldea init`) instead of hand-written YAML for
  CLI/headless users
- ⬜ Metrics/alerting (node health, storage usage) — local only, no
  external service required
- ⬜ Clear self-hosting guide: how to run `trackerd` on a free-tier VM
  (e.g. Oracle Cloud Always Free) so a group has a stable bootstrap point,
  with a documented default of "one member of the group hosts the
  tracker" — no paid or centrally-run service required at any point
- ⬜ "Getting started with your group" doc: end-to-end walkthrough for a
  non-technical group (5 friends, a family) to go from zero to a working
  shared pool

**Aldea is and will remain free and fully self-hosted — there is no paid
tier planned.** The project's value proposition is precisely that a group
of people can turn spare capacity on devices they already own into a
shared drive / lightweight "VPS" without paying anyone, including the
project's own maintainers. Do not introduce monetization, licensing gates,
or feature paywalls in any phase unless this document is explicitly
revised to say otherwise.

## Phase 5 — Compute layer (exploratory, not committed)

Not designed yet. Do not start implementation here without a dedicated
architecture discussion and new ADRs — this phase involves fundamentally
different problems (process isolation, security boundaries between
untrusted host machines, scheduling) that deserve their own design pass
rather than being bolted onto the storage architecture.

Rough shape being considered: lightweight container execution
(`Firecracker` or plain `containerd`) for services explicitly opted into by
node owners, with strict resource and network isolation. **Not started.**

## Explicitly out of scope (any phase)

- Blockchain / token incentive layer
- Public, open, untrusted-node network (Aldea is for closed/trusted groups)
- Multi-tenant SaaS at scale
- **Any paid tier, license gate, or feature paywall.** Aldea is free and
  self-hosted, full stop. If a future contributor proposes monetization,
  that requires a deliberate revision of this document and an ADR — it is
  not an assumed direction of the project.
