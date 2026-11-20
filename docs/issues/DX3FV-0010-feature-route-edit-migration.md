# DX3FV-0010: Durable multiport edit and Listener migration

- Priority: P1 resource lifecycle
- Status: IMPLEMENTED, REAL_AGENT_RECOVERY_PENDING
- Baseline: `v0.1.12-feature-multiport-diagnosis`
- Release tag: `v0.1.13-feature-route-edit-migration`
- Schema: 11

## Problem

Editing Landing, protocol, or ports changes Runtime listeners on two Agents.
Releasing the old Lease before both Agents acknowledge the new configuration
can let another business take a port that is still listening. A crash after one
Agent updates can also leave the database and data plane on different versions.

## Implementation

- `ix_route_migration` is a durable edit journal with complete old and new Route
  and Runtime snapshots. Runtime configuration uses a private snapshot type so
  the public `json:"-"` API protection does not remove rollback data.
- Replacement ports are allocated to migration-specific Owners. Existing Route
  Leases remain active during the complete operation.
- The IX Runtime set is applied before the GZ entry Runtime set. The embedded
  Agent's tested `UpdateService` transaction restores the previous service when
  replacement startup fails.
- Removed protocol services are deleted only after replacement updates.
- Only after every command succeeds does one database transaction update the
  Route and desired Runtime rows, transfer new Leases to the Route, release old
  Lease IDs, and close the journal.
- Any failure replays the complete old Runtime set on both nodes. A failed or
  interrupted rollback keeps both lease generations and state
  `rollback_pending`.
- Pending migration journals are included in the periodic Reconciler lookup.
  A reconnect or periodic retry restores the old set before normal route
  reconciliation.
- Active businesses cannot change PathProfile identity during edit.

## UI contract

The edit form loads current GZ and IX ports. Keeping a number retains that
resource when possible. Clearing a port requests a new automatic allocation.
Typing a different port requests that exact port; conflicts fail without silent
substitution. Landing host and port remain separate fields.

## Verification

- `TestIXRouteEditMigratesListenersBeforeReleasingOldLeases`
- `TestIXRouteEditFailureRestoresOldRuntimeAndKeepsOldLeases`
- `TestIXRouteInterruptedEditConvergesByRollingBackAndRetainingOldPorts`
- Existing create, pause, resume, delete isolation, traffic, and port-conflict
  tests continue to pass.
- Frontend TypeScript and production build pass.

## Remaining acceptance

The tests use the real encrypted Panel WebSocket command transport and mock
Agent acknowledgements. Process kill, real Agent restart, Runtime bind failure,
and two-node data-plane continuity still require the authorized test nodes.
Automatic edit allocation currently preserves the old route on a Runtime bind
conflict and reports failure; retrying multiple alternative ports in one edit
request is deferred until real-node behavior is characterized.

## Rollback

Revert the commit pointed to by `v0.1.13-feature-route-edit-migration`. Before
rollback, ensure no `ix_route_migration.active_key` begins with `route:`. Schema
11 may retain the unused migration journal table.
