# DX3FV-0007: Durable Agent flow outbox and idempotent Panel receipts

- Priority: P1 data integrity
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.9-feature-multiport-traffic`
- Release tag: `v0.1.10-feature-durable-flow-reports`
- Schema: 9

## Problem

Restoring counters in memory protects transient HTTP failures but does not
survive an Agent restart. Retrying after an uncertain ACK can also double count
traffic unless the Panel recognizes the same report.

## Protocol

- The JSON body remains the legacy array for old-Panel compatibility.
- A new Agent adds `X-DX3FV-Flow-Report-ID`, a random 128-bit lowercase hex ID.
- Before sending, the Agent atomically appends the pending batch beside
  `gost.json` as `gost-traffic-outbox.json` with mode 0600.
- Every snapshot has a distinct ID; new traffic is queued behind an uncertain
  older batch instead of being merged into that batch.
- On restart pending batches are sent in order before their receipt is removed.
- The outbox is removed only after the Panel returns `ok`.

## Panel transaction

Schema 9 adds `flow_upload_receipt`, uniquely keyed by Node identity and report
ID. The receipt and these counters commit in one database transaction:

- ordinary Forward, User, and UserTunnel cumulative flow;
- daily/monthly user quota usage;
- DX-3FV multiport route cumulative flow;
- directly classified peer-share usage.

A duplicate ID becomes a successful no-op. A database failure returns HTTP 503
so the Agent retains its outbox. Receipts older than seven days are removed by
daily maintenance; cumulative counters are not affected.

## Verification

- `TestGlobalTrafficManagerRecoversDurableOutboxAfterRestart` proves atomic
  persistence, 0600 permissions, stable ID reuse, payload recovery, and removal
  after ACK.
- `TestGlobalTrafficManagerQueuesNewTrafficBehindUncertainBatch` proves traffic
  generated during an outage receives a second durable ID and survives another
  restart.
- `TestApplyDurableFlowReportIsIdempotentAcrossCoreCounters` applies the same ID
  twice and proves every core counter changes exactly once.
- Backend, embedded go-gost/x, and top-level Agent suites pass.

## Compatibility and residual scope

Agents without the header retain the original ingestion behavior. Old Panels
ignore the header and still parse the unchanged body, but cannot deduplicate an
uncertain ACK. Derived federation cleanup and policy actions run after the
counter transaction and are idempotent control-plane side effects; the core
traffic values are committed exactly once.

## Rollback

Revert the commit pointed to by `v0.1.10-feature-durable-flow-reports`. The new
receipt table can remain unused. Before rolling back an Agent, allow its outbox
to receive an ACK or archive it so a pending batch is not silently discarded.
