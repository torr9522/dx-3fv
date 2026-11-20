# DX3FV-0002: Durable multiport runtime orchestration

- Priority: P1 feature
- Status: FIXED_IN_DX3FV
- Baseline: FLVX 3.0.1-rc16
- Release tag: `v0.1.5-feature-multiport-runtime`

## Requirement

Create an independent `GZ -> IX -> Landing` business lifecycle without using a
synthetic FLVX Tunnel. A path profile is static configuration; a route owns its
entry and IX leases plus two runtime lanes per selected protocol.

## Implementation

- Add administrator path and route APIs.
- Expand TCP+UDP into four named GOST services.
- Apply the IX listener before the GZ entry listener.
- Preserve leases while paused or while cleanup is unconfirmed.
- Release leases only after both runtime lanes acknowledge deletion.
- Reconcile pending work periodically and after an Agent reconnect/upgrade.
- Retry automatically allocated ports after a proven runtime bind conflict.
- Reject a manually selected conflicting port without substituting another.

GZ and IX ownership is scoped by immutable Node identity. A shared public IP
does not merge port resources. GZ2 uses its full-port pool with TCP/UDP runtime
occupancy checks and permanent exclusion of SSH port 22; Frontsea IX2 remains
an independent verified NAT mapping pool.

## Modified areas

- `go-backend/internal/http/handler/ix_routes.go`
- `go-backend/internal/http/handler/handler.go`
- `go-backend/internal/http/handler/jobs.go`
- `go-backend/internal/http/handler/upgrade.go`
- `go-backend/internal/store/repo/repository_ix_routes.go`
- focused handler and lifecycle tests

The FLVX Tunnel model, ordinary Forward handlers, GOST configuration schema,
and Agent command protocol are unchanged.

## Verification

- TCP+UDP creates exactly four services.
- IX services are applied before GZ entry services.
- Pause removes listeners and preserves both allocations.
- Resume restores listeners with the same allocations.
- Deleting route A leaves sibling route B and its leases unchanged.
- Leases remain pending until delete acknowledgements arrive.
- Automatic allocation retries a runtime bind conflict.
- Manual allocation reports a runtime bind conflict without substitution.
- `go test ./internal/http/handler -run '^TestIXRoute' -count=1 -v`
- `go test ./...`

Real node data-plane verification remains required before this feature can be
classified as operationally accepted.

## Rollback

From the main development branch, revert the commit pointed to by
`v0.1.5-feature-multiport-runtime`. Do not delete database rows manually; first
reconcile or remove test routes so Agent-owned listeners and leases reach their
terminal state.
