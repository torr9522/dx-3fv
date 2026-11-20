# DX3FV-0016: Block IX-origin probes in multiport diagnosis

> Superseded by DX3FV-0023 after the operator clarified that only external
> development-machine probes to an IX egress public IP are prohibited.

- Priority: P0 operational safety
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.19-fix-agent-transport-advisories`
- Release tag: `v0.1.20-fix-ix-route-egress-probes`
- Schema change: none

## Problem and evidence

`runIXRouteDiagnosis` dispatched `TcpPing` and `EchoProbe` commands to the IX
Agent with the configured Landing as the target. The optional payload mode also
generated synthetic end-to-end Echo traffic. This violates the operator's
explicit rule that IX egress may carry normal business forwarding but must not
originate diagnostic traffic.

The affected call chain was:

1. `POST /api/v1/ix-route/diagnose`;
2. `runIXRouteDiagnosis`;
3. `tcpConnectIXLeg` or `echoIXLeg` with `profile.IXNodeID`;
4. Agent `TcpPing` or `EchoProbe`;
5. direct socket activity from IX to Landing.

No production or test IX command was executed while confirming this source
path.

## Fix

- Keep the GZ-to-IX TCP Connect check, scoped to the route's registered public
  NAT address and port.
- Return `NOT_PROBED` for IX-to-Landing with
  `disabled_by_egress_policy`; no command is sent to the IX Agent.
- Return `NOT_PROBED` for end-to-end synthetic diagnosis. Real business
  observations must originate through the GZ entry and remain distinct from
  active probes.
- Leave UDP GZ-to-IX as `UNVERIFIED` until real business payload evidence is
  available. UDP connect is not promoted to success.
- Ignore the legacy `payloadEcho` request flag and persist it as false.
- Remove the browser Echo toggle and show the egress policy explicitly.
- Do not change GOST services, Runtime configuration, normal business
  forwarding, port leases, or traffic accounting.

## Verification

- `TestIXRouteDiagnosisDoesNotDispatchIXEgressOrSyntheticPayload` asserts that
  a diagnosis sends exactly one TCP command for GZ-to-IX, sends no `UdpPing` or
  `EchoProbe`, and persists `NOT_PROBED` evidence.
- `TestIXRouteDiagnosisNeverInfersIXEgressFromGZToIXSuccess` proves successful
  entry reachability cannot become an overall PASS.
- Frontend type checking and production build verify `NOT_PROBED` rendering and
  removal of the active Echo control.
- Browser acceptance must confirm the modal exposes the policy and cannot
  trigger IX egress probes.

## Rollback

Revert the commit pointed to by `v0.1.20-fix-ix-route-egress-probes`. Rollback
would restore prohibited IX-origin probes and therefore must not be deployed
while the current operator policy remains in force.
