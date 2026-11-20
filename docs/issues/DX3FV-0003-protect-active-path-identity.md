# DX3FV-0003: Protect active PathProfile identity

- Priority: P1 data integrity
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.5-feature-multiport-runtime`
- Release tag: `v0.1.6-fix-active-path-identity`

## Problem

The PathProfile update API allowed its GZ Node, IX Node, or either bound pool
to change while active routes still referenced it. Existing leases and Runtime
services retained the old identities, so the management view could claim that
the business belonged to different resources than the data plane actually
used.

## Root cause

Path updates validated that the submitted Node/pool combination was internally
valid, but did not compare it with the stored identity or lock against route
creation.

## Fix

- Serialize PathProfile identity updates and route creation on the Path row.
- Reject Node or pool changes while a non-deleted route references the Path.
- Continue to allow safe name and enabled-status changes.
- Validate status as either enabled or disabled.

## Verification

`TestIXPathProfileIdentityUpdateProtectsActiveBusinesses` proves that a safe
name/status update succeeds, an identity migration fails, and the rejected
update does not partially alter the stored PathProfile.

Run:

```bash
cd go-backend
go test ./internal/store/repo -run '^TestIXPathProfile' -count=1
go test ./...
```

## Rollback

Revert the commit pointed to by `v0.1.6-fix-active-path-identity`. Rolling back
reopens a management/data-plane consistency risk and is not recommended while
the Path update API is available.
