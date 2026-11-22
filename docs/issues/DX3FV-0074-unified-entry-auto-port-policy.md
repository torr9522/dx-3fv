# DX3FV-0074: Unified entry automatic port policy

- Priority: P1
- Affected versions: v1.0.1-v1.0.3
- State: FIXED_IN_DX3FV

## Root cause

Single-port and tunnel Forward rules shared a generic allocator, but it accepted
automatic candidates below 10000 and a legacy setting could switch allocation
back to ascending order. Clearing a port during edit retained the old value, and
a real Runtime bind conflict retried the same port. Multiport had already gained
the desired high random policy, but the three forwarding surfaces could drift.

## Fix

- A shared repository policy defines automatic entry ports as random available
  candidates in 10000-65535.
- Single-port, tunnel Forward and multiport paths use that policy while keeping
  their existing atomic ownership checks.
- Explicit manual ports, including 8000 and 9999, continue to use the node's
  full configured range and occupancy checks.
- Edit-clear requests now trigger a new automatic allocation.
- Forward Runtime bind conflicts exclude failed ports and atomically reallocate,
  with a bounded eight-attempt retry.
- The legacy sequential-allocation setting no longer overrides the policy and
  is removed from the UI. Stored legacy values remain harmless.
- IX NAT resources retain their independent verified mapping behavior.

## Rollback

Redeploy v1.0.3 and restore the pre-upgrade database backup. Schema 12 is
unchanged.
