# Acceptance Test Plan

Automated tests support development but do not replace browser and data-plane
acceptance. Every product milestone records the exact commit, artifact digest,
environment, timestamp, and protocol-specific result.

## Original FLVX regression

Verify all original routes, Forward and Tunnel CRUD, Node lifecycle, Agent
commands, permissions, monitoring, limits, sharing, configuration, backup, and
desktop/mobile layouts against the frozen reference deployment.

## GZ2 allocation correction

1. Manually allocate an available GZ2 entry port.
2. Automatically allocate a GZ2 entry port from the full eligible space.
3. Verify TCP-only, UDP-only, and atomic TCP+UDP allocation.
4. Verify port 22 is never offered or allocated.
5. Verify actual system listeners and explicit reservations conflict.
6. Verify concurrent creates cannot receive duplicate protocol resources.
7. Verify delete releases only after listener removal is observed.
8. Verify Agent restart preserves owner and lease.
9. Verify deleting one business leaves sibling businesses unchanged.
10. Verify GZ2 and Frontsea IX2 ownership never collides despite the shared
    public IP.

## Browser acceptance

At 1440px and 390px, a real Chromium session must log in, find each new menu,
create Path Profiles and businesses, use automatic and manual ports, inspect
ownership, edit, pause, resume, delete, recreate, view total traffic, and run
segment diagnostics. API calls and direct database fixtures cannot substitute
for this final workflow.

## Data plane and recovery

TCP and UDP are reported separately. UDP requires nonce-bearing payload proof,
not ICMP or a TCP result. Exercise Panel, Agent, runtime, WebSocket, and partial
deployment failures; then verify desired state, observations, leases, listeners,
traffic counters, and unrelated businesses converge correctly.
# Multiport edit migration

- Change Landing without changing ports; verify old Leases remain the same.
- Change manual GZ and IX ports; verify new listeners acknowledge before old
  Lease IDs become released.
- Clear each port to request automatic migration.
- Add and remove TCP/UDP lanes on the same numeric port atomically.
- Fail IX apply, fail GZ apply after IX succeeds, and fail rollback itself.
- Restart Panel during each migration state and verify the journal restores the
  old Runtime and retains old ownership until recovery completes.
- Delete a sibling route after an edit and prove its services and Leases never
  appear in edit commands.
