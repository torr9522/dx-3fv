# DX3FV-0019: Block service monitors on IX egress nodes

> Superseded by DX3FV-0023 after the operator clarified that authenticated
> Agent diagnostics against configured business targets are allowed.

- Priority: P0 operational safety
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.22-fix-manual-tunnel-egress-diagnosis`
- Release tag: `v0.1.23-fix-ix-service-monitor-probes`
- Schema change: none

## Problem and evidence

Enabled Service Monitor records are scanned in the background and dispatch
`ServiceMonitorCheck` to their selected Agent. TCP and ICMP checks can therefore
create recurring synthetic egress traffic from an IX node. The create, update,
manual-run, and scheduler paths had no node-role policy check.

Agent startup itself does not create these checks; the Panel scheduler is the
automatic source.

## Fix

- The repository classifies a node as egress-restricted when any of these
  persistent facts apply:
  - it is an `IXPathProfile.IXNodeID`;
  - it owns a non-`full` IX NAT/list port pool;
  - it is a Tunnel middle or exit node (`chain_type` 2 or 3).
- A GZ entry node owning only a `full` listener pool is not restricted.
- Monitor create/update rejects an enabled record on a restricted node.
- The scheduler and manual RunOnce path re-check persisted identity before
  dispatch, so old records cannot bypass the policy.
- Blocked checks return `NOT_PROBED` without calling the Agent.
- Normal GOST forwarding, flow reporting, and Agent lifecycle commands are not
  affected.

## Verification

- Repository tests cover full-entry, NAT-IX, and Tunnel-exit classification.
- Health checker tests prove `ServiceMonitorCheck` is never dispatched for a
  restricted IX node.
- API contract tests reject creation of an enabled monitor on an IX NAT node.
- Existing non-IX monitor CRUD and scheduler tests remain passing.

## Rollback

Revert the commit pointed to by `v0.1.23-fix-ix-service-monitor-probes`.
Rollback restores prohibited scheduled IX-origin checks and must not be
deployed under the current operator policy.
