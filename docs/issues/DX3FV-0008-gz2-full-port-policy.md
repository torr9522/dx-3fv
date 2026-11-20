# DX3FV-0008: Enforce the GZ2 full-port policy in the resource layer

- Priority: P1 resource integrity
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.10-feature-durable-flow-reports`
- Release tag: `v0.1.11-fix-gz2-full-port-policy`

## Problem

The HTTP handler normalized a direct-node pool to `1-65535`, but the repository
accepted a narrower range for the same `full` pool kind. An internal caller or
future migration could therefore accidentally apply an IX provider range to
GZ2 and violate the product contract.

## Fix

`Repository.CreateIXPortPool` now makes `full` mean `1-65535` regardless of the
caller. The allocation cursor may still start at an explicitly selected port;
that cursor is an allocation preference, not an eligibility boundary.

Port 22 is rejected in the allocator even if no resource row exists. Other
reserved ports are represented by excluded resources, existing FLVX ownership,
Agent Runtime observations, and bind-conflict observations. Automatic Runtime
bind conflicts are retried; manual conflicts fail without substitution.

Ownership is keyed by Node ID, internal port, and protocol. GZ2 and Frontsea
IX2 therefore remain independent even though they share a public IP.

## Verification

- `TestIXFullPoolExcludesSSHAndAllocatesTCPUDPAtomically`
- `TestIXFullPoolKeepsTCPAndUDPOwnershipIndependent`
- `TestIXAllocationIsScopedByNodeIdentityNotPublicIP`
- `TestIXConcurrentAllocationDoesNotDuplicateProtocolLease`
- `TestIXLeaseReleaseRequiresExplicitCompletion`
- `TestIXAllocationHonorsDeclaredRuntimeObservation`
- `TestIXRouteAutomaticPortRetriesRuntimeBindConflict`
- `TestIXRouteManualPortConflictFailsWithoutSubstitution`
- `TestIXRouteDeleteIsIsolatedAndReleasesAfterRuntimeAck`

These are control-plane tests. Real GZ2 Listener occupancy and restart behavior
remain pending until the GZ2 SSH identity and Agent connection are verified.

## Rollback

Revert the commit pointed to by `v0.1.11-fix-gz2-full-port-policy`. Existing
pool rows remain readable; rolling back only weakens creation-time validation.
