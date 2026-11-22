# DX3FV-0073: Multiport entry automatic port policy

- Priority: P1
- Affected version: v1.0.1
- State: FIXED_IN_DX3FV

## Evidence and root cause

Multiport entry ports used the generic full-pool allocator. An empty entry port
therefore followed `next_port` from the full 1-65535 range and allocated in
ascending order. The create, edit-clear and Runtime bind-conflict replacement
paths also constructed allocation requests independently.

## Fix

- Entry auto-allocation now uses a random, non-repeating traversal of the
  eligible 10000-65535 range.
- Existing resource exclusions, runtime observations and active Lease
  uniqueness checks remain authoritative.
- Create, edit-clear and bind-conflict replacement use one request helper.
- Manual entry ports and IX NAT resource allocation retain their existing
  behavior.
- The create/edit form describes the backend policy exactly.

## Verification

- Repository tests cover the minimum port, randomized permutation, conflicts,
  uniqueness and manual ports below 10000.
- Handler tests cover automatic create and edit-clear behavior while confirming
  the IX lane still allocates from its verified NAT pool.
- Full backend and frontend build suites are required before release.

## Rollback

Revert the focused commit and redeploy the previous stable native release. No
schema or stored route migration is involved.
