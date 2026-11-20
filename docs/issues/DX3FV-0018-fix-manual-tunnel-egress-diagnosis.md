# DX3FV-0018: Block middle/exit-origin manual diagnosis

> Superseded by DX3FV-0023 after the operator clarified that authenticated
> Agent diagnostics against configured business targets are allowed.

- Priority: P0 operational safety
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.21-fix-automatic-exit-egress-probes`
- Release tag: `v0.1.22-fix-manual-tunnel-egress-diagnosis`
- Schema change: none

## Problem and evidence

The original Forward and Tunnel diagnosis builders created work items from
chain-type 2 and 3 nodes. For type-2 tunnels this included Exit-to-public-target
and, for multi-hop paths, middle-to-next-hop probes. Clicking either diagnosis
button dispatched `TcpPing`/`UdpPing` from those Agents. QIX used as a type-2
exit would therefore originate extra traffic outside normal GOST forwarding.

## Fix

- Diagnosis work items whose source role is middle (`chain_type=2`) or exit
  (`chain_type=3`) are retained for topology visibility but marked
  `NOT_PROBED` before execution.
- The executor returns the policy reason without sending an Agent command.
- Entry-origin checks remain available for registered next-hop reachability.
- Stream progress counts `NOT_PROBED` separately instead of treating it as a
  failed network check.
- Forward and Tunnel desktop/mobile result components render an amber
  `NOT_PROBED` state, not success or failure.
- Runtime services and business traffic are unchanged.

## Verification

- Contract tests require every middle/exit-origin Forward and Tunnel result to
  carry `NOT_PROBED`.
- Handler tests cover both synchronous and streamed diagnosis paths.
- Frontend type checking/build covers the explicit status in both layouts.
- Browser acceptance must click the original Forward and Tunnel diagnosis
  controls and confirm the policy state remains visible.

## Rollback

Revert the commit pointed to by
`v0.1.22-fix-manual-tunnel-egress-diagnosis`. Rollback restores prohibited
middle/exit Agent probes and must not be deployed under the current policy.
