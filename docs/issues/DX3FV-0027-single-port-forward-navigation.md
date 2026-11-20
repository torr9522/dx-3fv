# DX3FV-0027: Single-port Forward and Forward navigation

- Priority: P0 product workflow
- Baseline: `v0.1.30-tunnel-transport-default-tcp`
- Target release: `v0.1.31-single-port-forward-navigation`

## Problem and design basis

The original Forward page required a Tunnel ID in both the UI and handler.
That made a simple entry-node-to-Landing rule impossible without first creating
a Type-1 Tunnel. A hidden or synthetic Tunnel would make ownership and deletion
semantics misleading, so direct rules need an explicit representation.

## Implementation

- `forward.tunnel_id=0` identifies a direct single-port Forward. Existing rows
  keep their positive Tunnel ID and are not migrated.
- `forward_port` remains the source of truth for the entry Node, bind address,
  and allocated port. It retains the existing transactional collision checks.
- `forward.protocol` stores `tcp`, `udp`, or `tcp+udp`. Existing rows receive
  the `tcp+udp` default, preserving the prior behavior where two services were
  always generated.
- The same Forward API, flow counters, lifecycle commands, service names, and
  GOST forward handler are reused. No IXPathProfile, Pool, Lease, IXRoute, or
  synthetic Tunnel is created.
- Direct create/edit is administrator-only and currently requires an Agent
  node. Tunnel-backed Forward behavior and user permissions are unchanged.

## Navigation and compatibility

Desktop navigation now contains a top-level Forward group with:

1. Single-port Forward (`/single-forward`)
2. Tunnel Forward (`/forward`)
3. Multiport Forward (`/multiport`)

The independent Tunnel resource page remains `/tunnel`. Existing `/forward`
and `/multiport` links remain valid. Mobile uses `/forwards` as the Forward
chooser and exposes the same three destinations.

## Tests

- Repository test proves a direct rule persists `tunnel_id=0`, the selected
  entry Node/port, and protocol.
- Control-plane tests prove TCP, UDP, and TCP+UDP generate exactly the expected
  Runtime services and reject transport labels such as TLS.
- Full backend and frontend builds cover the legacy Tunnel Forward paths.
- Browser acceptance creates, edits, pauses, resumes, and deletes direct rules
  through the new page, then checks the Agent Runtime listeners.

## Rollback

Switch the native Panel release to v0.1.30 with the matching pre-upgrade
database backup. Schema 12 adds one non-destructive column, but use the backup
for a complete rollback. Delete any test direct rule through the v0.1.31 UI
before rollback so its Runtime listener is removed cleanly.
