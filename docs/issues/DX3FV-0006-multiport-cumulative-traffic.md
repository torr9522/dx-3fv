# DX3FV-0006: Multiport per-business cumulative traffic

- Priority: P1 feature
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.8-fix-flow-report-loss`
- Release tag: `v0.1.9-feature-multiport-traffic`

## Requirement

Each multiport business needs one persistent cumulative traffic total. The two
data-plane segments must not both be charged because they represent the same
payload traversing GZ and IX.

## Ownership rule

Only GZ entry services named `dx3fv_ix_<route>_entry_<protocol>` are eligible.
The Panel then verifies that the reporting Node owns an `entry` runtime row for
that route before updating counters. IX lane service names are ignored.

TCP and UDP entry deltas are both accumulated for a TCP+UDP business. Incoming
bytes map to `in_flow`; outgoing bytes map to `out_flow`. The totals are stored
on `ix_route`, survive Panel restarts, and are formatted with the existing FLVX
traffic utility in the UI.

## Verification

`TestIXRouteTrafficCountsOnlyOwnedEntryLane` sends entry and IX lane reports,
then repeats the entry service name from the IX Node. Only the authentic GZ
entry bytes are persisted.

```bash
cd go-backend
go test ./internal/http/handler -run '^TestIXRoute' -count=1
go test ./...
cd ../vite-frontend
pnpm run build
```

## Residual risk

HTTP failure retention is covered by `DX3FV-0005`. Durable Agent restart
outbox recovery and duplicate-report receipts remain separate work and must be
completed before exact-once traffic accounting can receive operational PASS.

## Rollback

Revert the commit pointed to by `v0.1.9-feature-multiport-traffic`. Existing
stored counters remain in the database but new multiport reports stop updating
them.
