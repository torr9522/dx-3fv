# DX3FV-0017: Stop automatic probes from tunnel exit nodes

> Superseded by DX3FV-0023 after the operator clarified that authenticated
> Agent diagnostics against configured business targets are allowed.

- Priority: P0 operational safety
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.20-fix-ix-route-egress-probes`
- Release tag: `v0.1.21-fix-automatic-exit-egress-probes`
- Schema change: none

## Problem and evidence

The original tunnel-quality scheduler starts five seconds after Panel startup
and is enabled when its configuration key is absent. Its default interval is
one second. For type-2 tunnels it dispatched `TcpPing` from every selected and
candidate exit node to `www.bing.com:443` or a configured public target.
Best-exit scoring also required the same exit-origin public probe.

This behavior is automatic, is not normal GOST business forwarding, and can
cause an IX egress address to contact a protected external target repeatedly.
It violates the operator's IX egress policy.

No real node command was sent while reproducing the source call chain.

## Fix

- Type-2 tunnel quality keeps entry-to-exit reachability but never performs the
  exit-to-public-target leg.
- Exit candidate topology no longer schedules public-target probes.
- Best-exit scoring uses entry-to-exit latency/loss only. It never calls the
  candidate exit Agent. Best-exit probing is skipped for paths whose owner is a
  middle node because that would originate synthetic traffic from IX.
- API snapshots expose `egressProbeStatus: NOT_PROBED`; latency/loss use `-1`
  as the historical-storage sentinel.
- The monitor UI renders `NOT_PROBED` and the policy reason instead of showing
  a false success/failure.
- Type-1 entry-origin public monitoring remains unchanged because it does not
  originate at an IX/exit node.
- Runtime services and normal forwarding are unchanged.

## Verification

- Best-exit tests assert the only command is entry-to-exit and the exit public
  fields are `NOT_PROBED` sentinels.
- Tunnel-quality tests assert online backup exits and candidate exits never
  originate calls to the public target.
- Frontend type checking/build verifies the new status rendering.
- Browser acceptance must verify the monitor displays `NOT_PROBED` without
  changing route operation.

## Rollback

Revert the commit pointed to by
`v0.1.21-fix-automatic-exit-egress-probes`. Rollback restores prohibited
automatic exit-origin traffic and must not be deployed under the current
operator policy.
