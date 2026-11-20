# DX3FV-0009: Evidence-graded multiport diagnosis

- Priority: P2 diagnostics
- Status: IMPLEMENTED, REAL_DATA_PLANE_PENDING
- Baseline: `v0.1.11-fix-gz2-full-port-policy`
- Release tag: `v0.1.12-feature-multiport-diagnosis`
- Schema: 10

## Problem

The original Agent's `UdpPing` only calls `net.DialTimeout("udp", ...)` and
closes the socket. UDP dial success does not prove that a datagram reached the
peer or that a response returned. It therefore cannot support a UDP PASS.

TCP Connect also does not prove application payload completed the full route.
The UI needs to distinguish segment reachability, controlled payload evidence,
and facts that remain unverified.

## Implementation

- `EchoProbe` sends a caller-provided random nonce and requires the exact bytes
  to return before reporting success. TCP and UDP use real payload I/O and a
  deadline.
- Default diagnosis runs GZ-to-IX and IX-to-Landing TCP Connect checks.
- TCP end-to-end and every UDP leg remain `UNVERIFIED` by default.
- An explicit UI toggle enables nonce probes only when Landing is a controlled
  TCP/UDP echo fixture.
- UDP GZ-to-IX uses the IX route listener and therefore continues to Landing;
  its method is `udp_nonce_echo_via_ix_route`, not a pure link RTT.
- Per-leg method, status, latency, timestamp, and error are persisted in
  `ix_route_diagnosis` and returned to the UI.
- Segment values are displayed independently. They are not added and are not
  represented as application latency.

## Verification

- `TestEchoProbeRequiresExactPayloadRoundTrip` verifies real TCP and UDP nonce
  round trips against controlled local fixtures.
- `TestEchoProbeDoesNotTreatUDPDialAsPayloadSuccess` proves an open local UDP
  socket that never replies is a failure.
- `TestIXRouteDiagnosisDoesNotMisreportUDPConnectAsPayloadSuccess` proves the
  Panel never calls legacy `UdpPing` or promotes its semantics.
- `TestIXRouteDiagnosisRequiresNonceEchoForPayloadPass` proves PASS requires
  nonce evidence and diagnosis evidence persists.
- Frontend TypeScript and production build pass.

## Remaining acceptance

No GZ2, IX2, QIX, Blindbox, NAT, Shadowsocks, or controlled external echo data
plane was contacted in this change. Real TCP/UDP results remain UNKNOWN until
node identities and Agent access are verified and the controlled fixture is
deployed in the authorized test environment.

## Rollback

Revert the commit pointed to by `v0.1.12-feature-multiport-diagnosis`. Schema 10
can retain the unused diagnosis table. Older Agents do not know `EchoProbe`, so
the Panel will show an explicit command failure rather than a false PASS.
