# DX3FV-0005: Agent drops counters after a failed flow report

- Priority: P1 data integrity
- Status: FIXED_IN_DX3FV
- Baseline: FLVX 3.0.1-rc16
- Release tag: `v0.1.8-fix-flow-report-loss`

## Problem

The Agent reset every service counter before sending its five-second HTTP flow
batch. A timeout, non-200 response, or invalid response permanently discarded
that batch. Traffic for a just-deleted service was also removed before the
Panel acknowledged its final report.

## Root cause

`GlobalTrafficManager.collectAndReport` treated snapshot collection as
delivery. It had no rollback path for an unsuccessful `sendBatchTrafficReport`.

## Fix

- Restore the exact failed batch into each service counter.
- Merge restoration with bytes collected concurrently during the request.
- Remove stale service counters only after the Panel returns a successful ACK.
- Preserve the existing report body and old-Panel compatibility.

## Verification

`TestGlobalTrafficManagerRestoresCountersAfterReportFailure` injects an HTTP
failure and proves that both byte directions remain pending. The existing stale
service test proves that a successful ACK still clears the completed entry.

```bash
cd go-gost/x
go test ./service -run '^TestGlobalTrafficManager' -count=1
go test ./...
```

## Residual risk

This fix protects transient HTTP failures but the pending counters are still
memory-only. Agent crash/restart recovery and duplicate-report idempotency need
a durable outbox and Panel receipt protocol in a separate change.

## Rollback

Revert the commit pointed to by `v0.1.8-fix-flow-report-loss`. Rollback restores
the known traffic-loss behavior and is not recommended.
